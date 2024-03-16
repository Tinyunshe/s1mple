package rcd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"s1mple/pkg/config"
	"s1mple/rcd/img"
	"strings"
	"sync"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type Document struct {
	CloudId            string         `json:"cloudId"`
	Jira               string         `json:"jira"`
	Version            string         `json:"version"`
	AssigneeEmail      string         `json:"assignee_email"`
	Subject            string         `json:"subject"`
	Content            string         `json:"content"`
	ContentAttachments string         `json:"contentAttachments"`
	Comments           string         `json:"comments"`
	Assignee           string         `json:"assignee,omitempty"`
	PageId             string         `json:",omitempty"`
	ReleaserToken      string         `json:",omitempty"`
	Imgs               []img.Img      `json:",omitempty"`
	Config             *config.Config `json:",omitempty"`
	Logger             *zap.Logger    `json:",omitempty"`
}

var (
	lock sync.Mutex
)

// 如果jira为<空>，返回空字段
func (d *Document) fixJira() {
	if d.Jira == "<空>" {
		d.Jira = ""
	}
}

// 判断工单受理人决定使用的token,发布到对应受理人的confluence
func (d *Document) identifyReleaserToken() error {
	for _, v := range d.Config.ConfluenceSpec.Parts {
		if v.Username == d.AssigneeEmail {
			d.ReleaserToken = v.Token
			d.Logger.Info("Current document user is", zap.String("user", d.AssigneeEmail))
			return nil
		}
	}
	err := errors.New("")
	d.Logger.Error("Error Do not Identify any user", zap.Error(err))
	return err
}

// 修饰内容:
func (d *Document) adorn() error {
	// 删除内容中的“-----”号，修饰文档内容
	rjFunc := func(s string) string {
		return strings.Replace(s, "----------------------------------------", "", -1)
	}
	d.Comments = rjFunc(d.Comments)
	d.Content = rjFunc(d.Content)

	if strings.Contains(d.ContentAttachments, "</li>") && strings.Contains(d.ContentAttachments, "</ul>") {
		html, err := goquery.NewDocumentFromReader(strings.NewReader(d.ContentAttachments))
		if err != nil {
			d.Logger.Error("Error Adorn", zap.Error(err))
			return err
		}

		// 删除<ul>标签
		html.Find("ul").Each(func(_ int, s *goquery.Selection) {
			s.Contents().Unwrap()
		})

		// 删除<li>标签
		html.Find("li").Each(func(_ int, s *goquery.Selection) {
			s.Contents().Unwrap()
		})

		d.ContentAttachments, err = html.Html()
		if err != nil {
			d.Logger.Error("Error Adorn", zap.Error(err))
			return err
		}
	}
	return nil
}

// htmlHandler分为3个部分
// 1、解析html tag中的“img”或者“a”，然后Attr其中的“src”或者“href”
// 2、识别到后将其中的地址和存放img的目录，传入初始化img对象的函数
// 2、将img替换为confluence所识别的ac:image
func (d *Document) htmlHandler(dst string, tag string, childtag string) error {
	ga := ""
	switch dst {
	case "comments":
		ga = d.Comments
	case "content":
		ga = d.Content
	case "contentAttachments":
		ga = d.ContentAttachments
	}

	// 初始化解析html数据格式的对象
	html, err := goquery.NewDocumentFromReader(strings.NewReader(ga))
	if err != nil {
		d.Logger.Error("Error htmlHandler new html", zap.Error(err))
		return err
	}

	lock.Lock()
	// 如果找到的img长度不等于0，认为是存在img的
	if html.Find(tag).Length() != 0 {
		d.Logger.Info("Replace", zap.String("tag", tag))
		// 否则存在img，则实例化Img对象传入img http地址和img本地存放的目录
		html.Find(tag).Each(func(i int, s *goquery.Selection) {
			c, _ := s.Attr(childtag)

			// 初始化img对象，传入存放img文件的目录
			img := img.NewImg(c, d.Config.DocumentImgDirectory)
			d.Logger.Info("New img", zap.Any("", img))

			// 将img替换为confluence所识别的ac:image
			newTag := fmt.Sprintf(`<ac:image><ri:attachment ri:filename="%v" /></ac:image>`, img.Name)
			s.ReplaceWithHtml(newTag)

			// 追加到imgs对象列表
			d.Imgs = append(d.Imgs, *img)

		})
		replaceImgAfterHtml, err := html.Html()
		if err != nil {
			d.Logger.Error("Error htmlHandler result html", zap.Error(err))
			return err
		}

		switch dst {
		case "comments":
			d.Comments = replaceImgAfterHtml
		case "content":
			d.Content = replaceImgAfterHtml
		case "contentAttachments":
			d.ContentAttachments = replaceImgAfterHtml
		}
	} else {
		d.Logger.Info("", zap.String("No find img", dst))
	}

	lock.Unlock()
	return nil
}

// 由于html格式字符串无法直接传到json中,需要创建对象去构造,并返回post请求需要的reader
func (d *Document) constructReleaseBody(documentHtmlContent string) (*strings.Reader, error) {
	/*
		body 示例
				{
				    "type": "page",
				    "title": "hahaha",
				    "space": {"key": "~stwu"},
				    "body": {
				        "storage": {
				            "value": "",
				            "representation": "storage"
				        }
				    }
				}
	*/
	// 创建符合body json的匿名结构体传入数据
	crb := &struct {
		Title string `json:"title"`
		Type  string `json:"type"`
		Space struct {
			Key string `json:"key"`
		} `json:"space"`
		Body struct {
			Storage struct {
				Value          string `json:"value"`
				Representation string `json:"representation"`
			} `json:"storage"`
		} `json:"body"`
	}{
		Title: fmt.Sprintf(d.CloudId + "-" + d.Subject),
		Type:  "page",
		Space: struct {
			Key string "json:\"key\""
		}{"~" + d.Assignee},
		Body: struct {
			Storage struct {
				Value          string "json:\"value\""
				Representation string "json:\"representation\""
			} "json:\"storage\""
		}{Storage: struct {
			Value          string "json:\"value\""
			Representation string "json:\"representation\""
		}{Value: documentHtmlContent, Representation: "storage"}},
	}
	body, err := json.Marshal(crb)
	if err != nil {
		d.Logger.Error("Error construct confluence release post body, json marshal error", zap.Error(err))
		return nil, err
	}
	d.Logger.Info("construct confluence release post body success")
	return strings.NewReader(string(body)), nil
}

// 将Document中的所有字段数据 渲染到 -> 故障文档模板 ,返回的是html格式的大字符串,可理解为文档
func (d *Document) render() (string, error) {
	// 打开模板文件句柄
	file, err := os.Open(d.Config.GotemplatePath)
	if err != nil {
		d.Logger.Error("Error render open file", zap.Error(err))
		return "", err
	}
	defer file.Close()
	document, err := io.ReadAll(file)
	if err != nil {
		d.Logger.Error("Error render ioread document", zap.Error(err))
		return "", err
	}
	// 解析文件内容返回template对象
	t := template.Must(template.New("").Parse(string(document)))
	buf := &bytes.Buffer{}
	// 执行解析
	if err := t.Execute(buf, d); err != nil {
		d.Logger.Error("Error render execute gotemplate", zap.Error(err))
		return "", err
	}
	d.Logger.Info("Render document")
	return buf.String(), nil
}

// 传入文档,将文档发布到confluence,返回当前assignee在confluence的token
func (d *Document) release(documentHtmlContent string) error {
	// 通过对象构造body数据,返回reader
	payload, err := d.constructReleaseBody(documentHtmlContent)
	if err != nil {
		d.Logger.Error("Error construct release body", zap.Error(err))
		return err
	}

	// 声明发布到confluence请求的更多数据
	url := d.Config.ConfluenceUrl + "/rest/api/content"
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		d.Logger.Error("Error new request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.ReleaserToken)
	req.Header.Set("Connection", "keep-alive")

	var done bool
	for i := 0; i < d.Config.RetryCount; i++ {
		done = true
		resp, err := d.Config.ConfluenceSpec.HttpClient.Do(req)
		if err != nil {
			done = false
			d.Logger.Error("Error Release document, post to confluence url error", zap.Error(err))
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			d.Logger.Error("Error Read respone body", zap.Error(err))
			return err
		}
		if resp.StatusCode != http.StatusOK {
			err := errors.New(string(body))
			d.Logger.Error("Error respone code not 200", zap.Error(err))
			return err
		}
		// 发布confluence文档后，从confluence返回的响应body中，获取页面的pageId
		d.PageId = gjson.Get(string(body), "id").String()
		if d.PageId == "" {
			msg := "error release response: page id is empty"
			err := errors.New(msg)
			d.Logger.Error(msg, zap.Error(err))
			return err
		}
		resp.Body.Close()
		d.Logger.Info("Release document success", zap.String("Respone confluence pageId", d.PageId), zap.String("Assignee", d.Assignee))
		break
	}
	if !done {
		return errors.New("timeout")
	}

	return nil
}

// img处理
func (d *Document) imgHander() {
	if len(d.Imgs) != 0 {
		for i := 0; i < len(d.Imgs); i++ {

			go func(this int) {
				done := make(chan bool)
				go func() {
					d.Imgs[this].Download(done)
				}()
				go func() {
					d.Imgs[this].Upload(d.Config.ConfluenceUrl, d.PageId, d.ReleaserToken, d.Config.HttpClient, d.Config.RetryCount, done)
				}()
			}(i)

		}
	} else {
		d.Logger.Info("No exisit img", zap.String("PageId", d.PageId))
	}
}

// 并发处理comments，content，contentAttachments中的img和a标签，并替换数据
func parallelHTMLProcess(d *Document, logger *zap.Logger) {
	wg := sync.WaitGroup{}
	chanerr := make(chan error, 20)
	wg.Add(3)
	go func(err chan error) {
		err <- d.htmlHandler("comments", "img", "src")
		wg.Done()
	}(chanerr)
	go func(err chan error) {
		err <- d.htmlHandler("content", "img", "src")
		wg.Done()
	}(chanerr)
	go func(err chan error) {
		err <- d.htmlHandler("contentAttachments", "a", "href")
		wg.Done()
	}(chanerr)
	wg.Wait()
	close(chanerr)

	errdone := false
	for err := range chanerr {
		if err != nil {
			errdone = true
			logger.Error("Error goroutine html handler", zap.Error(err))
		}
	}
	if errdone {
		return
	}
}

// 处理来自udesk触发器的post请求中的json数据构造为Document文档对象
func newDocument(r *http.Request, config *config.Config, logger *zap.Logger) (*Document, error) {
	// 将config和logger传到document结构体中
	d := &Document{
		Logger: logger,
		Config: config,
	}

	// 从body内将json解析
	r.ParseForm()
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		d.Logger.Error("Error json decode", zap.Error(err))
		return nil, err
	}

	// 处理jira链接
	d.fixJira()

	// 处理工单处理人的名字，传入的是admin@alauda.io，返回admin
	d.Assignee = fmt.Sprintf(strings.Split(d.AssigneeEmail, "@")[0])

	// 确定发布者token
	err := d.identifyReleaserToken()
	if err != nil {
		return nil, err
	}

	// 修饰内容
	err = d.adorn()
	if err != nil {
		return nil, err
	}

	d.Logger.Info("New document success")
	return d, nil
}

// 发布confluence文档入口
func ReleaseConfluenceDocument(w http.ResponseWriter, r *http.Request, config *config.Config, logger *zap.Logger) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	doc, err := newDocument(r, config, logger)
	if err != nil {
		logger.Error("", zap.Error(err))
		http.Error(w, "Error create document", http.StatusBadRequest)
		return
	}

	parallelHTMLProcess(doc, logger)

	documentAfterRender, err := doc.render()
	if err != nil {
		logger.Error("", zap.Error(err))
		http.Error(w, "Error render document", http.StatusBadRequest)
		return
	}

	err = doc.release(documentAfterRender)
	if err != nil {
		logger.Error("", zap.Error(err))
		http.Error(w, "Error release document", http.StatusBadRequest)
		return
	}

	doc.imgHander()

	w.Write([]byte("ok"))
}
