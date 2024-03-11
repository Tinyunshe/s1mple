package img

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// 使用groutine将comments中的img下载到本地，上传到confluence
// 同时将上传后的img名称替换到comments中
// 下载与上传img的动作与后续的动作形成异步，不阻塞后续动作，已提高性能
type Img struct {
	HttpAddress string
	LocalPath   string
	Name        string
}

// 下载图片
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

// 上传图片到confluence
func (i *Img) Upload(cf string) {

}

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
