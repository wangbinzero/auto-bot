package serve

import "errors"

type Request struct {
	Id      string   `json:"id"`
	Biz     string   `json:"biz"`
	Op      string   `json:"op"`
	Channel []string `json:"channel"`
}

func (r Request) Valid() (error, bool) {
	if r.Id == "" {
		return errors.New("客户端标识错误"), false
	}
	return nil, true
}
