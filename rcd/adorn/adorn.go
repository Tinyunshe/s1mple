package adorn

import (
	"fmt"
	"s1mple/rcd/img"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

type Adorner struct {
	Logger     *zap.Logger
	htmlParser *goquery.Document
}

func NewAdorner(logger *zap.Logger) *Adorner {
	return &Adorner{
		Logger: logger,
	}
}

func (a *Adorner) Execute(data *string) *Adorner {
	// 初始化解析html数据格式的对象
	html, err := goquery.NewDocumentFromReader(strings.NewReader(*data))
	if err != nil {
		a.Logger.Error("Error Execute new htmlParser", zap.Error(err))
		return nil
	}
	a.htmlParser = html
	return a
}

func (a *Adorner) ReverseComments(comments string) string {
	builder := strings.Builder{}
	old := strings.Split(comments, "\n----------------------------------------\n")[1:]
	new := make([]string, len(old))
	for i := 1; i <= len(old); i++ {
		new[i-1] = old[len(old)-i]
	}
	for _, v := range new {
		builder.WriteString(v)
	}
	return builder.String()
}

func (a *Adorner) ImgTagHandler(tag string, childtag string, imgdir string, imgChan chan<- *img.Img) (string, error) {
	// 如果找到的img长度不等于0，认为是存在img的
	if a.htmlParser.Find(tag).Length() != 0 {
		a.Logger.Info("Replace", zap.String("HtmlTag", tag))
		// 否则存在img，则实例化Img对象传入img http地址和img本地存放的目录
		a.htmlParser.Find(tag).Each(func(i int, s *goquery.Selection) {
			c, _ := s.Attr(childtag)

			// 初始化img对象，传入存放img文件的目录
			img := img.NewImg(c, imgdir)
			if img != nil {
				a.Logger.Info("New img", zap.Any("", img))
				// 将img替换为confluence所识别的ac:image
				newTag := fmt.Sprintf(`<ac:image ac:height="400"><ri:attachment ri:filename="%v" /></ac:image>`, img.Name)
				s.ReplaceWithHtml(newTag)

				// 追加到imgs channel
				imgChan <- img
			} else {
				// 否则不是一个img格式的文件就提示此处文字
				s.ReplaceWithHtml("<在工单回复或者附件中,此处存在非图片格式的文件,确认后可以删除这段文字>")
			}
		})
	} else {
		a.Logger.Info("", zap.String("No find img", tag))
	}
	afterHtml, err := a.htmlParser.Html()
	if err != nil {
		a.Logger.Error("Error ImgTagHandler", zap.Error(err))
		return "", err
	}
	return afterHtml, nil
}

func (a *Adorner) DeleteSpareHtmlTag(tag string) (string, error) {
	a.htmlParser.Find(tag).Each(func(_ int, s *goquery.Selection) {
		s.Contents().Unwrap()
	})
	afterHtml, err := a.htmlParser.Html()
	if err != nil {
		a.Logger.Error("Error DeleteSpareHtmlTag", zap.Error(err))
		return "", err
	}
	return afterHtml, nil
}

func (a *Adorner) DeleteSpecialString(str string) string {
	if str == "<空>" {
		return ""
	}
	return str
}

func (a *Adorner) AdornProductClass(str string) string {
	return strings.Replace(str, ",", "-", -1)
}

func (a *Adorner) AdornVersion(str string) string {
	return strings.Replace(str, "v", "", -1)
}

func (a *Adorner) DeleteMacros(str string, macros []string) string {
	for _, s := range macros {
		str = strings.Replace(str, s, "", -1)
	}
	return str
}
