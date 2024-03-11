package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// func main() {
// 	payload := strings.NewReader(`{"id":"1","jira":"http://xxx","assignee":"stwu@alauda.io","subject":"docker异常","content":"docker异常怀疑机器导致","comments":"1、排查与机器有关"}`)
// 	url := "http://127.0.0.1:8080/release_confluence_document"
// 	req, _ := http.NewRequest(http.MethodPost, url, payload)
// 	req.Header.Add("Content-Type", "application/json")
// 	req.SetBasicAuth("admin", "1qazXSW@")
// 	client := http.Client{}
// 	resp, _ := client.Do(req)

// 	rb, _ := io.ReadAll(resp.Body)
// 	defer resp.Body.Close()
// 	fmt.Printf("result:\n%v\n", string(rb))

// }

// func test(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintln(w, "ok")
// }

// func main() {
// 	http.HandleFunc("/", auth.BasicAuth(test))
// 	http.ListenAndServe(":30002", nil)
// }

func main() {
	htmlContent := `\n----------------------------------------\n<p>您好，如果确认问题解决，请您关闭工单，有其他问题，请再联系我们，感谢您的咨询与反馈，祝您生活愉快，再见。</p>\n----------------------------------------\n<p>zabbix普通用户执行的进程不属于TKE的任何容器组件，可以顺着进程号查询父进程是否是docker运行</p>\n----------------------------------------\n<p>您好，您的工单我们已做升级处理，我是本次为您服务的高级工程师，您的问题我这边正在处理当中，稍后为您同步进展，请您稍等。</p>\n----------------------------------------\n<p>您好，排查如图<img src="https://pro-cs-freq.udeskcs.com/icon/tid99781/mceclip0_1667978009817_qe6m6.png" width=\"451\" height=\"178\" />这边为您做升级处理</p>\n----------------------------------------\n<p>您好，您反馈的问题已经收到，这边先看下，请您稍等。</p>\n----------------------------------------\n<p>您好，您反馈的问题已分配任务给工程师，请您耐心等待，稍后工程师将尽快进行处理，感谢您的等待与理解。<img src="https://pro-cs-freq.udeskcs.com/icon/tid99781/mceclip2_1710123492549_lfelo.jpg" />`

	html, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Println(err)
		return
	}

	src := []string{"./123.png", "./321.jpg"}
	html.Find("img").Each(func(i int, s *goquery.Selection) {
		// src, _ := s.Attr("src")
		// imgsMap[i] = src
		newNode := fmt.Sprintf(`<ac:image><ri:url ri:value="%s"/></ac:image>`, src[i])

		// 使用新内容替换原 img 标签
		s.ReplaceWithHtml(newNode)
	})
	content, _ := html.Html()
	fmt.Println(content)

	imgsMap := make(map[int]string, 5)
	imgsFileList := make([]string, 0, 5)
	for _, v := range imgsMap {
		resp, err := http.Get(v)
		if err != nil {
			fmt.Println(err)
			return
		}
		splitStr := strings.Split(v, "/")
		imgName := splitStr[len(splitStr)-1]
		filePath := "./" + imgName
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		file.Close()
		resp.Body.Close()
		imgsFileList = append(imgsFileList, filePath)
	}

	file, _ := os.Open(imgsFileList[0])
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", imgsFileList[0])
	io.Copy(part, file)
	writer.WriteField("minorEdit", "true")
	// writer.WriteField("comment", "Example attachment comment")
	if err != nil {
		fmt.Println("无法创建表单文件:", err)
		return
	}
	err = writer.Close()
	if err != nil {
		fmt.Println("Error closing writer:", err)
		return
	}

	pageId := "188188885"
	url := fmt.Sprintf("http://192.168.143.145:31871/rest/api/content/%v/child/attachment", pageId)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Atlassian-Token", "nocheck")
	req.Header.Set("Authorization", "Bearer NjMxOTQwNjQxNDkwOqDnS2dgE7gQ/dS62RhN8gg1L3L8")
	client := &http.Client{}
	resp, _ := client.Do(req)

	c, _ := io.ReadAll(resp.Body)
	fmt.Println(string(c))

}

type Img struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}
