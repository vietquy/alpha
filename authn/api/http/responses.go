package http

import (
	"net/http"
	"time"

	"github.com/vietquy/alpha"
)

var (
	_ alpha.Response = (*issueKeyRes)(nil)
	_ alpha.Response = (*revokeKeyRes)(nil)
)

type issueKeyRes struct {
	ID        string     `json:"id,omitempty"`
	Value     string     `json:"value,omitempty"`
	IssuedAt  time.Time  `json:"issued_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func (res issueKeyRes) Code() int {
	return http.StatusCreated
}

func (res issueKeyRes) Headers() map[string]string {
	return map[string]string{}
}

func (res issueKeyRes) Empty() bool {
	return res.Value == ""
}

type revokeKeyRes struct {
}

func (res revokeKeyRes) Code() int {
	return http.StatusNoContent
}

func (res revokeKeyRes) Headers() map[string]string {
	return map[string]string{}
}

func (res revokeKeyRes) Empty() bool {
	return true
}

type errorRes struct {
	Err string `json:"error"`
}
