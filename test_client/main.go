package main

import (
	"fmt"
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
	htmlContent := `\n----------------------------------------\n<p>您好，如果确认问题解决，请您关闭工单，有其他问题，请再联系我们，感谢您的咨询与反馈，祝您生活愉快，再见。</p>\n----------------------------------------\n<p>zabbix普通用户执行的进程不属于TKE的任何容器组件，可以顺着进程号查询父进程是否是docker运行</p>\n----------------------------------------\n<p>您好，您的工单我们已做升级处理，我是本次为您服务的高级工程师，您的问题我这边正在处理当中，稍后为您同步进展，请您稍等。</p>\n----------------------------------------\n<p>您好，排查如图<img src=\"https://pro-cs-freq.udeskcs.com/icon/tid99781/mceclip0_1667978009817_qe6m6.png\" width=\"451\" height=\"178\" />这边为您做升级处理</p>\n----------------------------------------\n<p>您好，您反馈的问题已经收到，这边先看下，请您稍等。</p>\n----------------------------------------\n<p>您好，您反馈的问题已分配任务给工程师，请您耐心等待，稍后工程师将尽快进行处理，感谢您的等待与理解。</p>`

	// http.Get("https://pro-cs-freq.udeskcs.com/icon/tid99781/mceclip0_1667978009817_qe6m6.png")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Println(err)
	}
	se := doc.Find("img")
	se.Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists {
			fmt.Println(src)
		}
	})
	se.ReplaceWithHtml()
}
