package config

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// 发布confluence文档功能配置
	ReleaseConfluenceDocument `yaml:"releaseConfluenceDocument"`
	Logger                    *zap.Logger `yaml:",omitempty"`
}

type ReleaseConfluenceDocument struct {
	// 关于confluence的配置
	ConfluenceSpec `yaml:"confluenceSpec"`
	// 故障模版gotemplate的文件位置
	GotemplatePath string `yaml:"gotemplatePath"`
	// html img临时存放的路径
	CommentsImgDirectory string `yaml:"commentsImgDirectory"`
}

type ConfluenceSpec struct {
	// confluence发布文档时对应的成员
	Parts []ConfluenceUser `yaml:"parts"`
	// confluence访问地址，http://xxx
	ConfluenceUrl string `yaml:"url"`
}

type ConfluenceUser struct {
	// confluence 用户与用户token
	// username:admin@alauda.io  /  token:xxxxxx
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
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

	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		panic("Initialize zap log error")
	}
	config.Logger = logger

	return config, nil
}
