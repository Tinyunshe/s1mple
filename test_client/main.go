package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
	postData := `{"id":"3","jira":"http://xxx","assignee":"stwu@alauda.io","subject":"docker异常","content":"docker异常怀疑机器导致","comments":"1、排查与机器有关"}`
	contentType := "application/json"
	datar := strings.NewReader(postData)
	resp, _ := http.Post("http://43.140.214.63:8081/release_confluence_document", contentType, datar)

	rb, _ := io.ReadAll(resp.Body)

	defer resp.Body.Close()
	fmt.Printf("result:\n%v\n", string(rb))

}
