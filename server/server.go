package server

import (
	"fmt"
	"net/http"
	"os"
	"s1mple/auth"
	"s1mple/config"
	"s1mple/rcd"
)

type Server struct {
	Config *config.Config
}

// 加载需要注册的URL
func (s *Server) loadUrl() {
	// auth.BasicAuth以rcdHander为参数，为rcdHander在正式执行之前添加了BasicAuth的功能
	// http.HandeFunc的最终执行者在auth.BasicAuth函数中的next.ServeHTTP(w,r)
	http.HandleFunc("/release_confluence_document", auth.BasicAuth(s.rcdHander))
	http.HandleFunc("/health", auth.BasicAuth(s.healthHander))
}

// 处理发布到confluence文档功能的闭包函数，作用是将外部的config属性传入到功能入口
func (s *Server) rcdHander(w http.ResponseWriter, r *http.Request) {
	func() {
		rcd.ReleaseConfluenceDocument(w, r, s.Config)
	}()
}

// health接口的处理
func (s *Server) healthHander(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}

// server启动入口
func (s *Server) Run() {
	s.loadUrl()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
