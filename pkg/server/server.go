package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"s1mple/notify"
	"s1mple/pkg/auth"
	"s1mple/pkg/config"
	"s1mple/rcd"
	"s1mple/rcd/img"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	Config                 *config.Config
	Logger                 *zap.Logger
	rcdCleanImgsTickerTime time.Duration
	rcdCleanImgsFileCycle  time.Duration
}

// 加载需要注册的URL
func (s *Server) loadUrl() {
	// auth.BasicAuth以rcdHander为参数，为rcdHander在正式执行之前添加了BasicAuth的功能
	// http.HandeFunc的最终执行者在auth.BasicAuth函数中的next.ServeHTTP(w,r)
	http.HandleFunc("/release_confluence_document", auth.BasicAuth(s.rcdHandler))
	http.HandleFunc("/health", auth.BasicAuth(s.healthHandler))
	http.HandleFunc("/notify/review", auth.BasicAuth(s.notifyHandler))
	http.HandleFunc("/notify/verificationcode", auth.BasicAuth(s.notifyHandler))
	http.HandleFunc("/notify/auto_remind", auth.BasicAuth(s.notifyHandler))
}

// 处理企微通知功能的闭包函数的功能入口
func (s *Server) notifyHandler(w http.ResponseWriter, r *http.Request) {
	func() {
		notify.Notify(w, r, s.Config, s.Logger)
	}()
}

// 处理发布到confluence文档功能的闭包函数，作用是将外部的config属性传入到功能入口
func (s *Server) rcdHandler(w http.ResponseWriter, r *http.Request) {
	func() {
		rcd.ReleaseConfluenceDocument(w, r, s.Config, s.Logger)
	}()
}

// 启动清理img目录定时器
func (s *Server) rcdCleanImgs() {
	// 每 1分钟，遍历存放目录下 以图片格式结尾的文件
	// 并且 文件创建时间 超过 30秒 的
	// 删除掉
	s.rcdCleanImgsTickerTime = 30 * time.Minute
	s.rcdCleanImgsFileCycle = 1 * time.Minute

	// 初始化定时器
	ticker := time.NewTicker(s.rcdCleanImgsTickerTime)
	defer ticker.Stop()

	// 定义执行函数
	cleanImgsFunc := func() {
		err := filepath.Walk(s.Config.DocumentImgDirectory, func(file string, info os.FileInfo, err error) error {
			if err != nil {
				s.Logger.Error("Error clean img accessing file", zap.Error(err))
				return err
			}
			if !info.IsDir() && img.HasImgFileType(file) {
				createTime := info.ModTime()
				if time.Since(createTime) > s.rcdCleanImgsFileCycle {
					err := os.Remove(file)
					if err != nil {
						s.Logger.Error("Error clean img remove file", zap.Error(err))
						return err
					}
					s.Logger.Info("clean img remove file", zap.String("file", file), zap.Any("time", createTime))
				}
			} else {
				s.Logger.Info("No clean img")
			}
			return nil
		})
		if err != nil {
			s.Logger.Error("Error clean img file walk", zap.Error(err))
			return
		}
	}

	// 程序运行后的首次执行
	cleanImgsFunc()
	// 启动定时器清理
	for range ticker.C {
		cleanImgsFunc()
	}
}

// 创建img存放目录
func (s *Server) rcdCreateImgDir() {
	if _, err := os.Stat(s.Config.DocumentImgDirectory); os.IsNotExist(err) {
		// 如果目录不存在，则创建它
		err := os.MkdirAll(s.Config.DocumentImgDirectory, 0755) // 0755 是目录权限
		if err != nil {
			s.Logger.Error("", zap.Error(err))
			panic("Error rcd create img directory")
		}
		s.Logger.Info("Create img directory", zap.String("", s.Config.DocumentImgDirectory))
	} else {
		s.Logger.Warn("Warning img directory exisit", zap.String("", s.Config.DocumentImgDirectory))
	}
}

// 监听退出信号
func (s *Server) waitForShutdown() {
	// 捕获程序退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// 程序退出时刷新日志
	s.Logger.Info("Shutting down server")
}

// health接口的处理
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}

// server启动入口
func (s *Server) Run() {
	s.rcdCreateImgDir()

	s.loadUrl()

	go http.ListenAndServe(":8080", nil)
	s.Logger.Info("Start server success :8080")

	go s.rcdCleanImgs()

	s.waitForShutdown()
}

func NewServer(config *config.Config, logger *zap.Logger) *Server {
	return &Server{
		Config: config,
		Logger: logger,
	}
}
