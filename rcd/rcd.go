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
	"s1mple/rcd/adorn"
	"s1mple/rcd/img"
	"strings"
	"text/template"
	"time"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type Document struct {
	CloudId            string                            `json:"cloudId"`
	Jira               string                            `json:"jira"`
	Version            string                            `json:"version"`
	ProductClass       string                            `json:"productClass"`
	AssigneeEmail      string                            `json:"assigneeEmail"`
	Subject            string                            `json:"subject"`
	Content            string                            `json:"content"`
	ContentAttachments string                            `json:"contentAttachments"`
	Comments           string                            `json:"comments"`
	PageId             string                            `json:",omitempty"`
	ReleaserToken      string                            `json:",omitempty"`
	ImgChan            chan *img.Img                     `json:",omitempty"`
	Config             *config.ReleaseConfluenceDocument `json:",omitempty"`
	Logger             *zap.Logger                       `json:",omitempty"`
	HttpClient         *http.Client                      `json:",omitempty"`
}

// 判断工单受理人决定使用的token,发布到对应受理人的confluence
func (d *Document) identifyReleaserToken() error {
	for _, v := range d.Config.Parts {
		d.Logger.Debug("Debug parts have", zap.String("assigneeEmail", v.Username))
		if strings.Contains(d.AssigneeEmail, v.Username) {
			d.ReleaserToken = v.Token
			d.Logger.Info("Current document user is", zap.String("user", d.AssigneeEmail))
			return nil
		}
	}
	err := errors.New("")
	d.Logger.Error("Error Do not Identify any assigneeEmail", zap.String("current assigneeEmail", d.AssigneeEmail), zap.Error(err))
	return err
}

// 由于html格式字符串无法直接传到json中,需要创建对象去构造,并返回post请求需要的reader
func (d *Document) constructReleaseBody(documentHtmlContent *string) (*strings.Reader, error) {
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
		Ancestors []struct {
			Id string `json:"id"`
		} `json:"ancestors"`
		Body struct {
			Storage struct {
				Value          string `json:"value"`
				Representation string `json:"representation"`
			} `json:"storage"`
		} `json:"body"`
	}{
		Title: fmt.Sprintf("%s-%s-%s", d.ProductClass, d.Subject, d.CloudId),
		Type:  "page",
		Space: struct {
			Key string "json:\"key\""
		}{d.Config.ReleaseSpace},
		Ancestors: []struct {
			Id string "json:\"id\""
		}{
			{Id: d.Config.ReleaseChildPageId},
		},
		Body: struct {
			Storage struct {
				Value          string "json:\"value\""
				Representation string "json:\"representation\""
			} "json:\"storage\""
		}{Storage: struct {
			Value          string "json:\"value\""
			Representation string "json:\"representation\""
		}{Value: *documentHtmlContent, Representation: "storage"}},
	}
	body, err := json.Marshal(crb)
	if err != nil {
		d.Logger.Error("Error construct confluence release post body, json marshal error", zap.Error(err))
		return nil, err
	}
	d.Logger.Info("construct confluence release post body success")
	d.Logger.Debug("Debug construct confluence release post body", zap.String("", string(body)))
	return strings.NewReader(string(body)), nil
}

// 将Document中的所有字段数据 渲染到 -> 故障文档模板 ,返回的是html格式的大字符串,可理解为文档
func (d *Document) render() (*string, error) {
	// 打开模板文件句柄
	file, err := os.Open(d.Config.GotemplatePath)
	if err != nil {
		d.Logger.Error("Error render open file", zap.Error(err))
		return nil, err
	}
	defer file.Close()
	document, err := io.ReadAll(file)
	if err != nil {
		d.Logger.Error("Error render ioread document", zap.Error(err))
		return nil, err
	}
	// 解析文件内容返回template对象
	t := template.Must(template.New("").Parse(string(document)))
	buf := &bytes.Buffer{}
	// 执行解析
	if err := t.Execute(buf, d); err != nil {
		d.Logger.Error("Error render execute gotemplate", zap.Error(err))
		return nil, err
	}
	data := buf.String()
	d.Logger.Info("Render document")
	d.Logger.Debug("Debug render document", zap.String("", data))
	return &data, nil
}

func (d *Document) adorn() {
	// err := errors.New("")
	a := adorn.NewAdorner(d.Logger)

	// 处理产品分类的字符串
	d.ProductClass = a.AdornProductClass(d.ProductClass)

	// 处理版本中的“v”
	d.Version = a.AdornVersion(d.Version)

	// 处理<空>值
	d.ContentAttachments = a.DeleteSpecialString(d.ContentAttachments)
	d.Jira = a.DeleteSpecialString(d.Jira)

	// 反转回复
	d.Comments = a.ReverseComments(d.Comments)

	// 删除“宏”
	d.Comments = a.DeleteMacros(d.Comments, d.Config.Macros)

	//lint:ignore SA4017 Ignore "New doesn't have side effects and its return value is ignored" warning
	//lint:ignore SA4006 Ignore "this value of err is never used" warning
	err := errors.New("")
	d.Comments, err = a.Execute(&d.Comments).ImgTagHandler("img", "src", d.Config.DocumentImgDirectory, d.ImgChan)
	if err != nil {
		d.Logger.Error("", zap.Error(err))
		return
	}
	d.Content, err = a.Execute(&d.Content).ImgTagHandler("img", "src", d.Config.DocumentImgDirectory, d.ImgChan)
	if err != nil {
		d.Logger.Error("", zap.Error(err))
		return
	}
	d.ContentAttachments, err = a.Execute(&d.ContentAttachments).ImgTagHandler("a", "href", d.Config.DocumentImgDirectory, d.ImgChan)
	if err != nil {
		d.Logger.Error("", zap.Error(err))
		return
	}
	// ImgTagHandler处理完后，要关闭ImgChan通道
	close(d.ImgChan)

	d.ContentAttachments, err = a.DeleteSpareHtmlTag("ul")
	if err != nil {
		d.Logger.Error("", zap.Error(err))
		return
	}
	d.ContentAttachments, err = a.DeleteSpareHtmlTag("li")
	if err != nil {
		d.Logger.Error("", zap.Error(err))
		return
	}
}

// 传入文档,将文档发布到confluence,并打上页面的标签
func (d *Document) release(documentHtmlContent *string) error {
	// 通过对象构造body数据,返回reader
	payload, err := d.constructReleaseBody(documentHtmlContent)
	if err != nil {
		d.Logger.Error("Error construct release body", zap.Error(err))
		return err
	}

	// 准备发布confluence请求的数据
	url := d.Config.ConfluenceUrl + "/rest/api/content"
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		d.Logger.Error("Error new request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.ReleaserToken)
	req.Header.Set("Connection", "keep-alive")

	// 发布confluence文档
	var done bool
	for i := 0; i < d.Config.RetryCount; i++ {
		done = true
		resp, err := d.HttpClient.Do(req)
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
		resp.Body.Close()

		// 发布confluence文档后，从confluence返回的响应body中，获取页面的pageId
		d.PageId = gjson.Get(string(body), "id").String()
		if d.PageId == "" {
			msg := "error release response: page id is empty"
			err := errors.New(msg)
			d.Logger.Error(msg, zap.Error(err))
			return err
		}

		// 创建页面标签
		err = d.createPageLabel()
		if err != nil {
			d.Logger.Error("Error create page label", zap.Error(err))
			return err
		}

		d.Logger.Info("Release document success", zap.String("Respone confluence pageId", d.PageId))
		break
	}
	if !done {
		return errors.New("timeout")
	}

	return nil
}

// 创建页面page的label
func (d *Document) createPageLabel() error {
	/*
		构造pageLabel请求体
		[{"prefix":"global","name":"kb-troub"},{"prefix":"global","name":"test"}]
	*/
	type pageLabelBody struct {
		Prefix string `json:"prefix"`
		Name   string `json:"name"`
	}

	cpl := make([]pageLabelBody, 0)
	for _, v := range d.Config.PageLabels {
		data := pageLabelBody{Prefix: "global", Name: v}
		cpl = append(cpl, data)
	}
	cplbody, err := json.Marshal(cpl)
	if err != nil {
		d.Logger.Error("Error createPageLabel json marshal", zap.Error(err))
		return err
	}
	payload := strings.NewReader(string(cplbody))

	// 准备请求content label接口
	// 接口示例 https: //confluence.alauda.cn/rest/api/content/214860307/label
	url := d.Config.ConfluenceUrl + "/rest/api/content/" + d.PageId + "/label"
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		d.Logger.Error("Error create page label new request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.ReleaserToken)
	req.Header.Set("Connection", "keep-alive")

	// 创建标签
	resp, err := d.HttpClient.Do(req)
	if err != nil {
		d.Logger.Error("Error create page label, http send failed", zap.Error(err))
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.Logger.Error("Error create page label read respone body", zap.Error(err))
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err := errors.New(string(body))
		d.Logger.Error("Error create page label respone code not 200", zap.Error(err))
		return err
	}
	resp.Body.Close()

	return nil
}

// 并发下载与上传img
func parallelIMGProcess(d *Document) {
	num := len(d.ImgChan)
	if num != 0 {
		for i := range d.ImgChan {
			this := i
			done := make(chan bool)
			go func() {
				this.Download(d.HttpClient, done)
			}()
			go func() {
				this.Upload(d.Config.ConfluenceUrl, d.PageId, d.ReleaserToken, d.HttpClient, d.Config.RetryCount, done)
			}()
		}
	} else {
		d.Logger.Info("No exisit img", zap.String("PageId", d.PageId))
	}
}

// 处理来自udesk触发器的post请求中的json数据构造为Document文档对象
func newDocument(r *http.Request, config *config.Config, logger *zap.Logger) (*Document, error) {
	// 将config和logger传到document结构体中
	d := &Document{
		Logger: logger,
		Config: &config.ReleaseConfluenceDocument,
		HttpClient: &http.Client{
			Timeout: time.Duration(config.ReleaseConfluenceDocument.ConfluenceSpec.Timeout) * time.Second,
		},
	}

	// 从body内将json解析
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		d.Logger.Error("Error json decode", zap.Error(err))
		return nil, err
	}

	// 确定发布者token
	err := d.identifyReleaserToken()
	if err != nil {
		return nil, err
	}

	// 初始化img channel，默认允许工单中出现50个img
	d.ImgChan = make(chan *img.Img, 50)

	d.Logger.Info("New document success", zap.String("cloudId", d.CloudId), zap.String("subject", d.Subject))
	d.Logger.Debug("Debug new document success", zap.String("cloudId", d.CloudId), zap.String("assigneeEmail", d.AssigneeEmail), zap.String("subject", d.Subject), zap.String("comments", d.Comments), zap.String("content", d.Content), zap.String("contentAttachments", d.ContentAttachments))
	return d, nil
}

// 发布confluence文档入口
func ReleaseConfluenceDocument(w http.ResponseWriter, r *http.Request, config *config.Config, logger *zap.Logger) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	d, err := newDocument(r, config, logger)
	if err != nil {
		logger.Error("", zap.Error(err))
		http.Error(w, "Error create document", http.StatusBadRequest)
		return
	}

	d.adorn()

	documentAfterRender, err := d.render()
	if err != nil {
		logger.Error("", zap.Error(err))
		http.Error(w, "Error render document", http.StatusBadRequest)
		return
	}

	err = d.release(documentAfterRender)
	if err != nil {
		logger.Error("", zap.Error(err))
		http.Error(w, "Error release document", http.StatusBadRequest)
		return
	}

	parallelIMGProcess(d)

	w.Write([]byte("ok"))
}
