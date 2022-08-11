package api

import (
	"net/http"

	"github.com/vietquy/alpha"
	"github.com/vietquy/alpha/reader"
)

var _ alpha.Response = (*pageRes)(nil)

type pageRes struct {
	reader.PageMetadata
	Total    uint64            `json:"total"`
	Messages []reader.Message `json:"messages,omitempty"`
}

func (res pageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res pageRes) Code() int {
	return http.StatusOK
}

func (res pageRes) Empty() bool {
	return false
}

type errorRes struct {
	Err string `json:"error"`
}
