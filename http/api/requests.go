package api

import (
	"github.com/vietquy/alpha/messaging"
)

type publishReq struct {
	msg   messaging.Message
	token string
}
