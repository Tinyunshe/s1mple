package rcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"s1mple/config"
	"s1mple/rcd/img"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
)

type Document struct {
	CloudId       string         `json:"cloudId"`
	PageId        string         `json:",omitempty"`
	Jira          string         `json:"jira"`
	Version       string         `json:"version"`
	AssigneeEmail string         `json:"assignee_email"`
	Assignee      string         `json:"assignee,omitempty"`
	Subject       string         `json:"subject"`
	Content       string         `json:"content"`
	Comments      string         `json:"comments"`
	Imgs          []img.Img      `json:",omitempty"`
	Config        *config.Config `json:",omitempty"`
}

// 将Document中的数据渲染到故障文档模板中,返回的是html格式的大字符串,可理解为文档
func (d *Document) render() (string, error) {
	// 打开模板文件句柄
	file, err := os.Open(d.Config.GotemplatePath)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	// 解析文件内容返回template对象
	t := template.Must(template.New("").Parse(string(content)))
	buf := &bytes.Buffer{}
	// 执行解析
	if err := t.Execute(buf, d); err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return buf.String(), nil
}

// 判断comments中是否有html img标签和img src的地址列表
// func (d *Document) isExisitHtmlImg(html *goquery.Document) bool {
// 	// 如果找到的img长度等于0，判断没有img
// 	if html.Find("img").Length() == 0 {
// 		return false
// 	} else {
// 		// 否则存在img，则实例化Img对象传入img http地址和img本地存放的目录
// 		imgs := make([]img.Img, 5)
// 		html.Find("img").Each(func(i int, s *goquery.Selection) {
// 			src, _ := s.Attr("src")
// 			imgs = append(imgs, *img.NewImg(src, d.Config.CommentsImgDirectory))
// 		})
// 		return true
// 	}
// }

// commentsHandler分为3个部分
// 1、识别html中的img同时并发下载img
// 2、将img替换为confluence所识别的ac:image
// 3、删除内容中的“-----”号，修饰文档内容
func (d *Document) commentsHandler() error {
	// 实例化html数据格式的解析对象
	html, err := goquery.NewDocumentFromReader(strings.NewReader(d.Comments))
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	// 如果找到的img长度不等于0，认为是存在img的
	if html.Find("img").Length() != 0 {
		// 否则存在img，则实例化Img对象传入img http地址和img本地存放的目录
		imgList := make([]img.Img, 5)
		html.Find("img").Each(func(i int, s *goquery.Selection) {
			src, _ := s.Attr("src")
			img := img.NewImg(src, d.Config.CommentsImgDirectory)
			// 并发下载img
			go img.Download()

			// 将img替换为confluence所识别的ac:image
			newTag := fmt.Sprintf(`<ac:image><ri:attachment ri:filename="%v" /></ac:image>`, img.Name)
			s.ReplaceWithHtml(newTag)

			// 追加到img对象列表
			imgList = append(imgList, *img)
		})

		d.Comments, err = html.Html()
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

// 传入文档,将文档发布到confluence
func (d *Document) release(documentHtmlContent string) error {
	// 通过对象构造body数据,返回reader
	payload, err := d.constructReleaseBody(documentHtmlContent)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	// 判断工单受理人决定使用的token,发布到对应受理人的confluence
	token := ""
	for _, v := range d.Config.ConfluenceSpec.Parts {
		if v.Username == d.AssigneeEmail {
			token = v.Token
		}
	}

	// 声明发布到confluence请求的更多数据
	url := d.Config.ConfluenceUrl + "/rest/api/content"
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Connection", "keep-alive")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	fmt.Println(string(body))

	// 发布confluence文档后，从confluence返回的响应body中，获取页面的pageId
	if !gjson.Get(string(body), "results").IsArray() {
		return fmt.Errorf("confluence respon body json error, results not found or type error")
	}
	for _, v := range gjson.Get(string(body), "results").Array() {
		d.PageId = v.Get("id").String()
	}
	return nil
}

// 处理工单处理人的名字，传入的是admin@alauda.io，返回admin
func (d *Document) fixAssignee() string {
	return fmt.Sprintf(strings.Split(d.AssigneeEmail, "@")[0])
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
		fmt.Println(err)
		return nil, fmt.Errorf(err.Error())
	}
	return strings.NewReader(string(body)), nil
}

// 处理来自udesk触发器的post请求中的json数据构造为Document文档对象
func newDocument(r *http.Request, config *config.Config) (*Document, error) {
	d := &Document{}

	// 从body内将json解析
	r.ParseForm()
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// 获取正确的Assignee名称
	d.Assignee = d.fixAssignee()

	// 将config传到document结构体中
	d.Config = config
	return d, nil
}

// 发布confluence文档入口
func ReleaseConfluenceDocument(w http.ResponseWriter, r *http.Request, config *config.Config) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	doc, err := newDocument(r, config)
	if err != nil {
		http.Error(w, "Error create document", http.StatusBadRequest)
		return
	}

	err = doc.commentsHandler()
	if err != nil {
		http.Error(w, "Error fix document comments", http.StatusBadRequest)
		return
	}

	documentHtmlContent, err := doc.render()
	if err != nil {
		http.Error(w, "Error render document", http.StatusBadRequest)
		return
	}

	err = doc.release(documentHtmlContent)
	if err != nil {
		http.Error(w, "Error release document", http.StatusBadRequest)
		return
	}

	for i := 0; i <= len(doc.Imgs); i++ {
		go doc.Imgs[i].Upload(doc.Config.ConfluenceUrl, doc.PageId)
	}

	w.Write([]byte("ok"))
}
