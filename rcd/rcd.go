package rcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"s1mple/config"
	"strings"
	"text/template"
)

type Document struct {
	Id       string         `json:"id"`
	Jira     string         `json:"jira"`
	Version  string         `json:"version"`
	Assignee string         `json:"assignee"`
	Subject  string         `json:"subject"`
	Content  string         `json:"content"`
	Comments string         `json:"comments"`
	Config   *config.Config `json:",omitempty"`
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
		if v.Username == d.Assignee {
			token = v.Token
		}
	}

	// 声明发布到confluence请求的更多数据
	url := d.Config.ConfluenceSpec.ConfluenceUrl + "/rest/api/content"
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Connection", "keep-alive")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	defer resp.Body.Close()

	resbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	fmt.Println(string(resbody))
	return nil
}

// 由于html格式字符串无法直接传到json中,需要创建对象去构造,并返回post请求需要的reader
func (d *Document) constructReleaseBody(documentHtmlContent string) (*strings.Reader, error) {
	// 处理工单处理人的名字，传入的是admin@alauda.io，返回admin
	fixNameFunc := func() string {
		return fmt.Sprintf("~" + strings.Split(d.Assignee, "@")[0])
	}

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
		Title: fmt.Sprintf(d.Id + "-" + d.Subject),
		Type:  "page",
		Space: struct {
			Key string "json:\"key\""
		}{fixNameFunc()},
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

	// 将config传到document结构体中
	d.Config = config
	return d, nil
}

// 发布confluence文档入口
func ReleaseConfluenceDocument(w http.ResponseWriter, r *http.Request, config *config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	doc, err := newDocument(r, config)
	if err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	documentHtmlContent, err := doc.render()
	if err != nil {
		http.Error(w, "Error render doc", http.StatusBadRequest)
		return
	}

	err = doc.release(documentHtmlContent)
	if err != nil {
		http.Error(w, "Error release doc", http.StatusBadRequest)
		return
	}

	w.Write([]byte("ok"))
}
