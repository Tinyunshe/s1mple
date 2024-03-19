package notify

import (
	"fmt"
	"net/http"
	"s1mple/pkg/config"

	"go.uber.org/zap"
)

type Notifier interface {
	contentHander() error
	Send() error
}

// 通知功能入口
func Notify(w http.ResponseWriter, req *http.Request, config *config.Config, logger *zap.Logger) {
	defer req.Body.Close()

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var notifier Notifier
	switch req.URL.Path {
	case "/notify/review":
		err := notifier.Send()
		if err != nil {
			logger.Error("", zap.Error(err))
			return
		}
		err = notifier.contentHander()
		if err != nil {
			logger.Error("", zap.Error(err))
			return
		}
	case "/notify/auto_remind":
		err := notifier.Send()
		if err != nil {
			logger.Error("", zap.Error(err))
			return
		}
		err = notifier.contentHander()
		if err != nil {
			logger.Error("", zap.Error(err))
			return
		}
	case "/notify/verificationcode":
		err := notifier.Send()
		if err != nil {
			logger.Error("", zap.Error(err))
			return
		}
		err = notifier.contentHander()
		if err != nil {
			logger.Error("", zap.Error(err))
			return
		}
	}

	w.Write([]byte("ok"))
}

func Entrances(notifier Notifier) {
	err := notifier.contentHander()
	if err != nil {
		fmt.Println(err)
	}
	err = notifier.Send()
	if err != nil {
		fmt.Println(err)
	}
}
