package review

import (
	"encoding/json"
	"io"
	"net/http"
	"s1mple/pkg/config"

	"go.uber.org/zap"
)

type Review struct {
	CloudId       string         `json:"cloudId"`
	Subject       string         `json:"subject"`
	AssigneeEmail string         `json:"assigneeEmail"`
	HttpClient    *http.Client   `json:",omitempty"`
	Config        *config.Config `json:",omitempty"`
	Logger        *zap.Logger    `json:",omitempty"`
}

func (r *Review) contentHandler() {

}

// https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=93911e50-0447-425d-81ef-b4e4e00b1083
func (r *Review) Send(body io.Reader) error {
	url := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=693axxx6-7aoc-4bc4-97a0-0ec2sifa5aaa"
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		r.Logger.Error("", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.HttpClient.Do(req)
	if err != nil {
		r.Logger.Error("", zap.Error(err))
		return err
	}
	defer resp.Body.Close()
	return nil
}

func NewReview(req *http.Request, config *config.Config, logger *zap.Logger) (*Review, error) {
	r := &Review{
		Config: config,
		Logger: logger,
	}
	err := json.NewDecoder(req.Body).Decode(&r)
	if err != nil {
		logger.Error("", zap.Error(err))
		return nil, err
	}
	return r, nil
}
