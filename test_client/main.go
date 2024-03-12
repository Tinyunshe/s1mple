package main

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

// 上传img
// func main() {
// 	htmlContent := `\n----------------------------------------\n<p>您好，如果确认问题解决，请您关闭工单，有其他问题，请再联系我们，感谢您的咨询与反馈，祝您生活愉快，再见。</p>\n----------------------------------------\n<p>zabbix普通用户执行的进程不属于TKE的任何容器组件，可以顺着进程号查询父进程是否是docker运行</p>\n----------------------------------------\n<p>您好，您的工单我们已做升级处理，我是本次为您服务的高级工程师，您的问题我这边正在处理当中，稍后为您同步进展，请您稍等。</p>\n----------------------------------------\n<p>您好，排查如图<img src="https://pro-cs-freq.udeskcs.com/icon/tid99781/mceclip0_1667978009817_qe6m6.png" width=\"451\" height=\"178\" />这边为您做升级处理</p>\n----------------------------------------\n<p>您好，您反馈的问题已经收到，这边先看下，请您稍等。</p>\n----------------------------------------\n<p>您好，您反馈的问题已分配任务给工程师，请您耐心等待，稍后工程师将尽快进行处理，感谢您的等待与理解。<img src="https://pro-cs-freq.udeskcs.com/icon/tid99781/mceclip2_1710123492549_lfelo.jpg" />`

// 	html, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	src := []string{"./123.png", "./321.jpg"}
// 	html.Find("img").Each(func(i int, s *goquery.Selection) {
// 		// src, _ := s.Attr("src")
// 		// imgsMap[i] = src
// 		newNode := fmt.Sprintf(`<ac:image><ri:url ri:value="%s"/></ac:image>`, src[i])

// 		// 使用新内容替换原 img 标签
// 		s.ReplaceWithHtml(newNode)
// 	})
// 	content, _ := html.Html()
// 	fmt.Println(content)

// 	imgsMap := make(map[int]string, 5)
// 	imgsFileList := make([]string, 0, 5)
// 	for _, v := range imgsMap {
// 		resp, err := http.Get(v)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		splitStr := strings.Split(v, "/")
// 		imgName := splitStr[len(splitStr)-1]
// 		filePath := "./" + imgName
// 		file, err := os.Create(filePath)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		_, err = io.Copy(file, resp.Body)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		file.Close()
// 		resp.Body.Close()
// 		imgsFileList = append(imgsFileList, filePath)
// 	}

// 	file, _ := os.Open(imgsFileList[0])
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	part, err := writer.CreateFormFile("file", imgsFileList[0])
// 	io.Copy(part, file)
// 	writer.WriteField("minorEdit", "true")
// 	// writer.WriteField("comment", "Example attachment comment")
// 	if err != nil {
// 		fmt.Println("无法创建表单文件:", err)
// 		return
// 	}
// 	err = writer.Close()
// 	if err != nil {
// 		fmt.Println("Error closing writer:", err)
// 		return
// 	}

// 	pageId := "188188885"
// 	url := fmt.Sprintf("http://192.168.143.145:31871/rest/api/content/%v/child/attachment", pageId)
// 	req, _ := http.NewRequest("POST", url, body)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())
// 	req.Header.Set("X-Atlassian-Token", "nocheck")
// 	req.Header.Set("Authorization", "Bearer NjMxOTQwNjQxNDkwOqDnS2dgE7gQ/dS62RhN8gg1L3L8")
// 	client := &http.Client{}
// 	resp, _ := client.Do(req)

// 	c, _ := io.ReadAll(resp.Body)
// 	fmt.Println(string(c))

// }

// type Img struct {
// 	Title string `json:"title"`
// 	Type  string `json:"type"`
// }

// import (
// 	"fmt"

// 	"github.com/tidwall/gjson"
// )

// func main() {
// 	jsonStr := `{"results":[{"id":"188188894","type":"attachment","status":"current","title":"mceclip0_1667978009817_qe6m6.png","version":{"by":{"type":"known","username":"stwu","userKey":"939086d6715493da0174ae81a03d0044","profilePicture":{"path":"/download/attachments/75079352/user-avatar","width":48,"height":48,"isDefault":false},"displayName":"Shuting Wu","_links":{"self":"http://192.168.143.145:31871/rest/api/user?key=939086d6715493da0174ae81a03d0044"},"_expandable":{"status":""}},"when":"2024-03-11T19:24:06.964+08:00","number":1,"minorEdit":true,"hidden":false,"_links":{"self":"http://192.168.143.145:31871/rest/experimental/content/188188894/version/1"},"_expandable":{"content":"/rest/api/content/188188894"}},"container":{"id":"188188885","type":"page","status":"current","title":"36540-ACP生产环境日志节点进程问题","extensions":{"position":"none"},"_links":{"webui":"/pages/viewpage.action?pageId=188188885","edit":"/pages/resumedraft.action?draftId=188188885","tinyui":"/x/1Yg3Cw","self":"http://192.168.143.145:31871/rest/api/content/188188885"},"_expandable":{"container":"/rest/api/space/~stwu","metadata":"","operations":"","children":"/rest/api/content/188188885/child","restrictions":"/rest/api/content/188188885/restriction/byOperation","history":"/rest/api/content/188188885/history","ancestors":"","body":"","version":"","descendants":"/rest/api/content/188188885/descendant","space":"/rest/api/space/~stwu"}},"metadata":{"mediaType":"image/png","_expandable":{"currentuser":"","frontend":"","editorHtml":"","properties":""}},"extensions":{"mediaType":"image/png","fileSize":1035478},"_links":{"webui":"/pages/viewpage.action?pageId=188188885&preview=%2F188188885%2F188188894%2Fmceclip0_1667978009817_qe6m6.png","download":"/download/attachments/188188885/mceclip0_1667978009817_qe6m6.png?version=1&modificationDate=1710156246964&api=v2","thumbnail":"/download/thumbnails/188188885/mceclip0_1667978009817_qe6m6.png?api=v2","self":"http://192.168.143.145:31871/rest/api/content/188188894"},"_expandable":{"operations":"","children":"/rest/api/content/188188894/child","restrictions":"/rest/api/content/188188894/restriction/byOperation","history":"/rest/api/content/188188894/history","ancestors":"","body":"","descendants":"/rest/api/content/188188894/descendant","space":"/rest/api/space/~stwu"}}],"size":1,"_links":{"base":"http://192.168.143.145:31871","context":""}}`

// 	// for _, v := range gjson.Get(jsonStr, `results`).Array() {
// 	// 	fmt.Println(v.Get("id").String())
// 	// }
// }

// func main() {
// 	htmlContent := `\n----------------------------------------\n<p>您好，如果确认问题解决，请您关闭工单，有其他问题，请再联系我们，感谢您的咨询与反馈，祝您生活愉快，再见。</p>\n----------------------------------------\n<p>zabbix普通用户执行的进程不属于TKE的任何容器组件，可以顺着进程号查询父进程是否是docker运行</p>\n----------------------------------------\n<p>您好，您的工单我们已做升级处理，我是本次为您服务的高级工程师，您的问题我这边正在处理当中，稍后为您同步进展，请您稍等。</p>\n----------------------------------------\n<p>您好，排查如图<img src="https://pro-cs-freq.udeskcs.com/icon/tid99781/mceclip0_1667978009817_qe6m6.png" width=\"451\" height=\"178\" />这边为您做升级处理</p>\n----------------------------------------\n<p>您好，您反馈的问题已经收到，这边先看下，请您稍等。</p>\n----------------------------------------\n<p>您好，您反馈的问题已分配任务给工程师，请您耐心等待，稍后工程师将尽快进行处理，感谢您的等待与理解。<img src="https://pro-cs-freq.udeskcs.com/icon/tid99781/mceclip2_1710123492549_lfelo.jpg" />`

// 	// html, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }

// 	// html.Find("p").Each(func(i int, s *goquery.Selection) {
// 	// 	fmt.Println(s.Html())
// 	// })
// }

// func main() {
// 	logger, _ := zap.NewProduction()
// }
