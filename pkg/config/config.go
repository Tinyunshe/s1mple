package config

import (
	"flag"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// 发布confluence文档功能配置
	ReleaseConfluenceDocument ReleaseConfluenceDocument `yaml:"releaseConfluenceDocument"`
	LogLevel                  string                    `yaml:"logLevel,omitempty"`
}

type ReleaseConfluenceDocument struct {
	// 关于confluence的配置
	ConfluenceSpec `yaml:"confluenceSpec"`
	// 故障模版gotemplate的文件位置
	GotemplatePath string `yaml:"gotemplatePath"`
	// html img临时存放的路径
	DocumentImgDirectory string `yaml:"documentImgDirectory"`
	// 发布到confluence的目标空间
	ReleaseSpace string `yaml:"releaseSpace"`
	// 发布到confluence目标空间的子页面id
	ReleaseChildPageId string `yaml:"releaseChildPageId"`
	// 需要清理掉的“宏”文字
	Macros []string `yaml:"macros"`
	// confluence发布文档时对应的成员
	Parts []ConfluenceUser `yaml:"parts"`
}

type ConfluenceSpec struct {
	// confluence访问地址，http://xxx
	ConfluenceUrl string `yaml:"url"`
	// 请求confluence超时时间,默认10s
	Timeout int `yaml:"timeout,omitempty"`
	// 重试次数，默认2
	RetryCount int `yaml:"retryCount,omitempty"`
}

type ConfluenceUser struct {
	// confluence 用户与用户token
	// username:admin@alauda.io  /  token:xxxxxx
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

// default args
func (config *Config) defaultValue() {
	if config.ReleaseConfluenceDocument.Timeout == 0 {
		config.ReleaseConfluenceDocument.Timeout = 10
	}
	if config.ReleaseConfluenceDocument.ConfluenceSpec.RetryCount == 0 {
		config.ReleaseConfluenceDocument.ConfluenceSpec.RetryCount = 2
	}
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
}

func NewConfig() *Config {
	// 声明flag名称,默认值,帮助提示,返回字符串指针
	configPtr := flag.String("config", "", "path to config file")
	// 解析命令行参数
	flag.Parse()
	// 检查是否提供了配置文件路径
	if *configPtr == "" {
		panic("Usage: s1mple --config /path/to/config.yaml")
	}

	// 声明config对象
	config := &Config{}
	file, err := os.Open(*configPtr)
	if err != nil {
		panic(err)
	}
	yamlDecoder := yaml.NewDecoder(file)
	err = yamlDecoder.Decode(config)
	if err != nil {
		panic(err)
	}

	// 初始化default args
	config.defaultValue()

	return config
}
