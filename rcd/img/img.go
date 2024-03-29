package img

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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
	lock        sync.Mutex
	ImgFileType = []string{".jpg", ".png", ".gif", ".ico", ".svg", ".jpeg", ".pdf"}
)

// 从httpaddress中分离img名称
func identifyName(addr string) string {
	u, _ := url.Parse(addr)
	u.RawQuery = ""
	splitStr := strings.Split(u.String(), "/")
	return splitStr[len(splitStr)-1]
}

// 下载img
func (i *Img) Download(client *http.Client, done chan<- bool) {
	req, err := http.NewRequest(http.MethodGet, i.HttpAddress, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// 保存到本地路径下
	file, err := os.Create(i.LocalPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("download success", resp.Status, i.HttpAddress)

	done <- true
}

// 上传img到confluence，传入confluence地址，发布文档后返回的pageId，和当前assignee在confluence的token
func (i *Img) Upload(cf string, pageId string, token string, client *http.Client, retryCount int, done <-chan bool) {
	<-done

	// 打开img本地路径的句柄
	file, err := os.Open(i.LocalPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// 创建buffer，通过multipart包创建媒体文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", i.LocalPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 将img写入请求缓存中
	lock.Lock()
	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println(err)
		lock.Unlock()
		return
	}
	lock.Unlock()

	// 添加媒体类请求中的body参数
	writer.WriteField("minorEdit", "true")
	// writer.WriteField("comment", "Example attachment comment")
	lock.Lock()
	err = writer.Close()
	if err != nil {
		fmt.Println("Error closing writer:", err)
		lock.Unlock()
		return
	}
	lock.Unlock()

	// 发起上传img到confluence的请求
	url := fmt.Sprintf("%v/rest/api/content/%v/child/attachment", cf, pageId)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Atlassian-Token", "nocheck")
	req.Header.Set("Authorization", "Bearer "+token)

	lock.Lock()
	status := ""
	for i := 0; i < retryCount; i++ {
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Println(string(body))
		}
		status = resp.Status

		defer resp.Body.Close()
		break
	}
	lock.Unlock()
	fmt.Println("upload success", status, i.LocalPath)
}

// 定义筛选文件后缀的函数
func HasImgFileType(name string) bool {
	for _, v := range ImgFileType {
		if strings.HasSuffix(name, v) {
			return true
		}
	}
	return false
}

// 初始化Img对象
func NewImg(address string, dir string) *Img {
	name := identifyName(address)
	if HasImgFileType(name) {
		return &Img{
			HttpAddress: address,
			Name:        name,
			LocalPath:   fmt.Sprintf("%v/%v", dir, name),
		}
	} else {
		return nil
	}
}
