package config

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// 发布confluence文档功能配置
	ReleaseConfluenceDocument `yaml:"releaseConfluenceDocument"`
}

type ReleaseConfluenceDocument struct {
	// 关于confluence的配置
	ConfluenceSpec `yaml:"confluenceSpec"`
	// 故障模版gotemplate的文件位置
	GotemplatePath string `yaml:"gotemplatePath"`
	// html img临时存放的路径
	DocumentImgDirectory string `yaml:"documentImgDirectory"`
}

type ConfluenceSpec struct {
	// confluence发布文档时对应的成员
	Parts []ConfluenceUser `yaml:"parts"`
	// confluence访问地址，http://xxx
	ConfluenceUrl string `yaml:"url"`
	// 请求confluence超时时间,默认10s
	Timeout int `yaml:"timeout,omitempty"`
	// 重试次数，默认2
	RetryCount int `yaml:"retryCount,omitempty"`
	// 请求confluence的http client
	HttpClient *http.Client `yaml:",omitempty"`
}

type ConfluenceUser struct {
	// confluence 用户与用户token
	// username:admin@alauda.io  /  token:xxxxxx
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

func (c *Config) initConfluenceHttpClient() {
	if c.ConfluenceSpec.Timeout == 0 {
		c.ConfluenceSpec.Timeout = 10
	}
	if c.ConfluenceSpec.RetryCount == 0 {
		c.ConfluenceSpec.RetryCount = 2
	}
	c.ConfluenceSpec.HttpClient = &http.Client{
		Timeout: time.Duration(c.ConfluenceSpec.Timeout) * time.Second,
	}
}

func NewConfig() (*Config, error) {
	// 声明flag名称,默认值,帮助提示,返回字符串指针
	configPtr := flag.String("config", "", "path to config file")
	// 解析命令行参数
	flag.Parse()
	// 检查是否提供了配置文件路径
	if *configPtr == "" {
		fmt.Println("Usage: s1mple --config /path/to/config.yaml")
		os.Exit(1)
	}

	// 声明config对象
	config := &Config{}
	file, err := os.Open(*configPtr)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	yamlDecoder := yaml.NewDecoder(file)
	err = yamlDecoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// 初始化请求confluence的http client
	config.initConfluenceHttpClient()

	return config, nil
}
