package img

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
)

// 使用groutine将comments中的img下载到本地，上传到confluence
// 同时将上传后的img名称替换到comments中
// 下载与上传img的动作与后续的动作形成异步，不阻塞后续动作，已提高性能
type Img struct {
	HttpAddress string
	LocalPath   string
	Name        string
}

var (
	lock sync.Mutex
)

func splitName(addr string) string {
	splitStr := strings.Split(addr, "/")
	return splitStr[len(splitStr)-1]
}

func NewImg(address string, dir string) *Img {
	return &Img{
		HttpAddress: address,
		Name:        splitName(address),
		LocalPath:   fmt.Sprintf("%v/%v", dir, splitName(address)),
	}
}

// 下载img
func (i *Img) Download() {
	resp, err := http.Get(i.HttpAddress)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// 保存到本地路径下
	file, err := os.Create(i.LocalPath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println(err)
	}
}

// 上传img到confluence，传入confluence地址，发布文档后返回的pageId，和当前assignee在confluence的token
func (i *Img) Upload(cf string, pageId string, token string) {
	// 打开img本地路径的句柄
	file, err := os.Open(i.LocalPath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	// 创建buffer，通过multipart包创建媒体文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", i.LocalPath)
	if err != nil {
		fmt.Println(err)
	}

	// 将img写入请求缓存中
	lock.Lock()
	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println(err)
		lock.Unlock()
	}
	lock.Unlock()

	// 添加媒体类请求中的body参数
	writer.WriteField("minorEdit", "true")
	// writer.WriteField("comment", "Example attachment comment")
	err = writer.Close()
	if err != nil {
		fmt.Println("Error closing writer:", err)
	}

	// 初始化http client
	client := &http.Client{}

	// 发起上传img到confluence的请求
	url := fmt.Sprintf("%v/rest/api/content/%v/child/attachment", cf, pageId)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Atlassian-Token", "nocheck")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	// 打印返回body
	c, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(c))
}
