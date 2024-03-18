package adorn

import (
	"fmt"
	"s1mple/pkg/config"
	"s1mple/rcd/img"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

type Adorner struct {
	Config *config.Config
	Logger *zap.Logger
	data   *string
}

// htmlTag分为4个部分
// 1、解析html tag中的“img”或者“a”，然后Attr其中的“src”或者“href”
// 2、识别到后将其中的地址和存放img的目录，传入初始化img对象的函数
// 3、将img替换为confluence所识别的ac:image
// 4、修饰不需要的html tag
func (a *Adorner) htmlTag(imgChan chan<- *img.Img) (*string, error) {
	// 初始化解析html数据格式的对象
	html, err := goquery.NewDocumentFromReader(strings.NewReader(*a.data))
	if err != nil {
		a.Logger.Error("Error htmlHandler new html", zap.Error(err))
		return nil, err
	}

	findImgFunc := func(tag string, childtag string) {
		// 如果找到的img长度不等于0，认为是存在img的
		if html.Find(tag).Length() != 0 {
			a.Logger.Info("Replace", zap.String("tag", tag))
			// 否则存在img，则实例化Img对象传入img http地址和img本地存放的目录
			html.Find(tag).Each(func(i int, s *goquery.Selection) {
				c, _ := s.Attr(childtag)

				// 初始化img对象，传入存放img文件的目录
				img := img.NewImg(c, a.Config.DocumentImgDirectory)
				if img != nil {
					a.Logger.Info("New img", zap.Any("", img))
					// 将img替换为confluence所识别的ac:image
					newTag := fmt.Sprintf(`<ac:image><ri:attachment ri:filename="%v" /></ac:image>`, img.Name)
					s.ReplaceWithHtml(newTag)

					// 追加到imgs channel
					imgChan <- img
				} else {
					// 否则不是一个img格式的文件就删掉tag
					s.ReplaceWithHtml("")
				}
			})
		} else {
			a.Logger.Info("", zap.String("No find img", tag))
		}
	}

	adornContentAttachmentsFunc := func(tag string) {
		html.Find(tag).Each(func(_ int, s *goquery.Selection) {
			s.Contents().Unwrap()
		})
	}

	findImgFunc("img", "src")
	findImgFunc("a", "href")
	adornContentAttachmentsFunc("ul")
	adornContentAttachmentsFunc("li")

	afterHtml, err := html.Html()
	if err != nil {
		a.Logger.Error("Error htmlHandler result html", zap.Error(err))
		return nil, err
	}
	close(imgChan)
	return &afterHtml, nil
}

// 修饰文本相关
func (a *Adorner) text() {
	*a.data = strings.Replace(*a.data, "----------------------------------------", "", -1)
}

func Execute(data *string, imgChan chan *img.Img, config *config.Config, logger *zap.Logger) (*string, error) {
	a := &Adorner{Config: config, Logger: logger, data: data}
	var err error
	a.data, err = a.htmlTag(imgChan)
	if err != nil {
		a.Logger.Error("", zap.Error(err))
		return nil, err
	}
	a.text()
	a.Logger.Info("Adorn success")
	a.Logger.Debug("Debug adorn", zap.String("", *a.data))
	return a.data, nil
}
