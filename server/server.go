package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"s1mple/auth"
	"s1mple/config"
	"s1mple/rcd"
	"syscall"

	"go.uber.org/zap"
)

type Server struct {
	Config *config.Config
	Logger *zap.Logger
}

// 加载需要注册的URL
func (s *Server) loadUrl() {
	// auth.BasicAuth以rcdHander为参数，为rcdHander在正式执行之前添加了BasicAuth的功能
	// http.HandeFunc的最终执行者在auth.BasicAuth函数中的next.ServeHTTP(w,r)
	http.HandleFunc("/release_confluence_document", auth.BasicAuth(s.rcdHandler))
	http.HandleFunc("/health", auth.BasicAuth(s.healthHandler))
}

// health接口的处理
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}

// 处理发布到confluence文档功能的闭包函数，作用是将外部的config属性传入到功能入口
func (s *Server) rcdHandler(w http.ResponseWriter, r *http.Request) {
	func() {
		rcd.ReleaseConfluenceDocument(w, r, s.Config, s.Logger)
	}()
}

func (s *Server) rcdCleanImgs() {
	// 待处理
}

func (s *Server) waitForShutdown() {
	// 捕获程序退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// 程序退出时刷新日志
	s.Logger.Info("Shutting down server")
}

// server启动入口
func (s *Server) Run() {
	s.loadUrl()

	go http.ListenAndServe(":8080", nil)
	s.Logger.Info("Start server success :8080")

	// go s.rcdCleanImgs()

	s.waitForShutdown()
}
