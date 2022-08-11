package api

import (
	"net/http"

	"github.com/go-zoo/bone"
	"github.com/vietquy/alpha"
)

// MakeHandler returns a HTTP API handler with version and metrics.
func MakeHandler(svcName string) http.Handler {
	r := bone.New()
	r.GetFunc("/version", alpha.Version(svcName))

	return r
}
