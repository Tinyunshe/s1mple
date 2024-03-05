package server

import (
	"fmt"
	"net/http"
	"os"
	"s1mple/config"
	"s1mple/rcd"
)

type Server struct {
	Config *config.Config
}

func (s *Server) LoadUrl() {
	http.HandleFunc("/release_confluence_document", func(w http.ResponseWriter, r *http.Request) {
		func() {
			rcd.ReleaseConfluenceDocument(w, r, s.Config)
		}()
	})
	http.HandleFunc("/health", healthHander)
}

func healthHander(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}

func (s *Server) Run() {
	s.LoadUrl()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
