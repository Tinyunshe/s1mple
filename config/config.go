package config

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ReleaseConfluenceDocument `yaml:"releaseConfluenceDocument"`
}

type ConfluenceUser struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type ReleaseConfluenceDocument struct {
	ConfluenceUrl  string           `yaml:"confluenceUrl"`
	GotemplatePath string           `yaml:"gotemplatePath"`
	Parts          []ConfluenceUser `yaml:"parts"`
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
	return config, nil
}
