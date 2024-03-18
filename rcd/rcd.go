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
	CloudId            string         `json:"cloudId"`
	Jira               string         `json:"jira"`
	Version            string         `json:"version"`
	ProductClass       string         `json:"productClass"`
	AssigneeEmail      string         `json:"assignee_email"`
	Subject            string         `json:"subject"`
	Content            string         `json:"content"`
	ContentAttachments string         `json:"contentAttachments"`
	Comments           string         `json:"comments"`
	Assignee           string         `json:"assignee,omitempty"`
	PageId             string         `json:",omitempty"`
	ReleaserToken      string         `json:",omitempty"`
	ImgChan            chan *img.Img  `json:",omitempty"`
	Config             *config.Config `json:",omitempty"`
	Logger             *zap.Logger    `json:",omitempty"`
	HttpClient         *http.Client   `json:",omitempty"`
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
		}{"~" + d.Assignee},
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

// 传入文档,将文档发布到confluence,返回当前assignee在confluence的token
func (d *Document) release(documentHtmlContent *string) error {
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
		// 发布confluence文档后，从confluence返回的响应body中，获取页面的pageId
		d.PageId = gjson.Get(string(body), "id").String()
		if d.PageId == "" {
			msg := "error release response: page id is empty"
			err := errors.New(msg)
			d.Logger.Error(msg, zap.Error(err))
			return err
		}
		resp.Body.Close()
		d.Logger.Info("Release document success", zap.String("Respone confluence pageId", d.PageId))
		break
	}
	if !done {
		return errors.New("timeout")
	}

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
		Config: config,
		HttpClient: &http.Client{
			Timeout: time.Duration(config.ConfluenceSpec.Timeout) * time.Second,
		},
	}

	// 从body内将json解析
	r.ParseForm()
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		d.Logger.Error("Error json decode", zap.Error(err))
		return nil, err
	}

	// 处理工单处理人的名字，传入的是admin@alauda.io，返回admin
	d.Assignee = fmt.Sprintf(strings.Split(d.AssigneeEmail, "@")[0])

	// 处理产品分类的字符串
	d.ProductClass = strings.Replace(d.ProductClass, ",", "-", -1)

	// 处理版本中的“v”
	d.Version = strings.Replace(d.Version, "v", "", -1)

	// 确定发布者token
	err := d.identifyReleaserToken()
	if err != nil {
		return nil, err
	}

	// 初始化img channel，默认允许工单中出现50个img
	d.ImgChan = make(chan *img.Img, 50)

	d.Logger.Info("New document success", zap.String("cloudId", d.CloudId), zap.String("subject", d.Subject), zap.String("assignee", d.Assignee))
	d.Logger.Debug("Debug new document success", zap.Any("", d))
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

	documentAfterRender, err := d.render()
	if err != nil {
		logger.Error("", zap.Error(err))
		http.Error(w, "Error render document", http.StatusBadRequest)
		return
	}

	documentAfterHtmlHandler, err := adorn.Execute(documentAfterRender, d.ImgChan, config, logger)
	if err != nil {
		logger.Error("", zap.Error(err))
		http.Error(w, "Error adorn document", http.StatusBadRequest)
		return
	}

	err = d.release(documentAfterHtmlHandler)
	if err != nil {
		logger.Error("", zap.Error(err))
		http.Error(w, "Error release document", http.StatusBadRequest)
		return
	}

	parallelIMGProcess(d)

	w.Write([]byte("ok"))
}
