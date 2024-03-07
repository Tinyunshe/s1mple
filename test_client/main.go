package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
	payload := strings.NewReader(`{"id":"1","jira":"http://xxx","assignee":"stwu@alauda.io","subject":"docker异常","content":"docker异常怀疑机器导致","comments":"1、排查与机器有关"}`)
	url := "http://127.0.0.1:8080/release_confluence_document"
	req, _ := http.NewRequest(http.MethodPost, url, payload)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth("admin", "1qazXSW@")
	client := http.Client{}
	resp, _ := client.Do(req)

	rb, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Printf("result:\n%v\n", string(rb))

}

// func test(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintln(w, "ok")
// }

// func main() {
// 	http.HandleFunc("/", auth.BasicAuth(test))
// 	http.ListenAndServe(":30002", nil)
// }
