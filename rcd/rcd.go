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
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type Document struct {
	CloudId       string         `json:"cloudId"`
	Jira          string         `json:"jira"`
	Version       string         `json:"version"`
	AssigneeEmail string         `json:"assignee_email"`
	Subject       string         `json:"subject"`
	Content       string         `json:"content"`
	Comments      string         `json:"comments"`
	Assignee      string         `json:"assignee,omitempty"`
	PageId        string         `json:",omitempty"`
	ReleaserToken string         `json:",omitempty"`
	Imgs          []img.Img      `json:",omitempty"`
	Config        *config.Config `json:",omitempty"`
	Logger        *zap.Logger    `json:",omitempty"`
}

// 处理工单处理人的名字，传入的是admin@alauda.io，返回admin
func (d *Document) fixAssignee() string {
	return fmt.Sprintf(strings.Split(d.AssigneeEmail, "@")[0])
}

// commentsHandler分为3个部分
// 1、识别html中的img tag，初始化img对象传入document.Imgs切片
// 2、将img替换为confluence所识别的ac:image
// 3、删除内容中的“-----”号，修饰文档内容
func (d *Document) commentsHandler() error {
	// 修饰内容
	d.Comments = strings.Replace(d.Comments, "----------------------------------------", "", -1)
	// 初始化解析html数据格式的对象
	html, err := goquery.NewDocumentFromReader(strings.NewReader(d.Comments))
	if err != nil {
		d.Logger.Error("CommentsHandler new html error", zap.Error(err))
		return err
	}
	// 如果找到的img长度不等于0，认为是存在img的
	if html.Find("img").Length() != 0 {
		d.Logger.Info("Replace img")
		// 否则存在img，则实例化Img对象传入img http地址和img本地存放的目录
		html.Find("img").Each(func(i int, s *goquery.Selection) {
			src, _ := s.Attr("src")

			// 初始化img对象，传入存放img文件的目录
			img := img.NewImg(src, d.Config.DocumentImgDirectory)
			d.Logger.Info("find img", zap.Any("", img))
			// 将img替换为confluence所识别的ac:image
			newTag := fmt.Sprintf(`<ac:image><ri:attachment ri:filename="%v" /></ac:image>`, img.Name)
			s.ReplaceWithHtml(newTag)

			// 追加到imgs对象列表
			d.Imgs = append(d.Imgs, *img)
		})
		replaceImgAfterHtml, err := html.Html()
		if err != nil {
			d.Logger.Error("CommentsHandler result html error", zap.Error(err))
			return err
		}
		d.Comments = replaceImgAfterHtml
	}
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
		d.Logger.Error("constructReleaseBody json marshal error", zap.Error(err))
		return nil, err
	}
	d.Logger.Info("ConstructReleaseBody")
	return strings.NewReader(string(body)), nil
}

// 将Document中的所有字段数据 渲染到 -> 故障文档模板 ,返回的是html格式的大字符串,可理解为文档
func (d *Document) render() (string, error) {
	// 打开模板文件句柄
	file, err := os.Open(d.Config.GotemplatePath)
	if err != nil {
		d.Logger.Error("Render open file error", zap.Error(err))
		return "", err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		d.Logger.Error("Render read content error", zap.Error(err))
		return "", err
	}
	// 解析文件内容返回template对象
	t := template.Must(template.New("").Parse(string(content)))
	buf := &bytes.Buffer{}
	// 执行解析
	if err := t.Execute(buf, d); err != nil {
		d.Logger.Error("Render execute gotemplate error", zap.Error(err))
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
		d.Logger.Error("construct release body error", zap.Error(err))
		return err
	}

	// 判断工单受理人决定使用的token,发布到对应受理人的confluence
	for _, v := range d.Config.ConfluenceSpec.Parts {
		if v.Username == d.AssigneeEmail {
			d.ReleaserToken = v.Token
		}
	}

	// 声明发布到confluence请求的更多数据
	url := d.Config.ConfluenceUrl + "/rest/api/content"
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		d.Logger.Error("New request error", zap.Error(err))
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
			d.Logger.Error("Release document error, post to confluence url error", zap.Error(err))
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			d.Logger.Error("Read respone body error", zap.Error(err))
			return err
		}
		if resp.StatusCode != http.StatusOK {
			err := errors.New(string(body))
			d.Logger.Error("respone code not 200", zap.Error(err))
			return err
		}
		// 发布confluence文档后，从confluence返回的响应body中，获取页面的pageId
		d.PageId = gjson.Get(string(body), "id").String()
		if d.PageId == "" {
			msg := "release response error: page id is empty"
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
	for i := 0; i < len(d.Imgs); i++ {
		go func(this int) {
			done := make(chan bool)
			go func() {
				d.Imgs[this].Download(done)
			}()
			go func() {
				d.Imgs[this].Upload(d.Config.ConfluenceUrl, d.PageId, d.ReleaserToken, done)
			}()
		}(i)
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
		d.Logger.Error("json decode", zap.Error(err))
		return nil, err
	}

	// 获取正确的Assignee名称
	d.Assignee = d.fixAssignee()

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
		doc.Logger.Error("", zap.Error(err))
		http.Error(w, "Error create document", http.StatusBadRequest)
		return
	}

	err = doc.commentsHandler()
	if err != nil {
		doc.Logger.Error("", zap.Error(err))
		http.Error(w, "Error fix document comments", http.StatusBadRequest)
		return
	}

	documentHtmlContent, err := doc.render()
	if err != nil {
		doc.Logger.Error("", zap.Error(err))
		http.Error(w, "Error render document", http.StatusBadRequest)
		return
	}

	err = doc.release(documentHtmlContent)
	if err != nil {
		doc.Logger.Error("", zap.Error(err))
		http.Error(w, "Error release document", http.StatusBadRequest)
		return
	}

	if len(doc.Imgs) != 0 {
		doc.imgHander()
	} else {
		logger.Info("No exisit img", zap.String("PageId", doc.PageId))
	}

	w.Write([]byte("ok"))
}
