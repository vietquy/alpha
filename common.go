package alpha

import (
	"os"
	"encoding/json"
	"net/http"
)

const version string = "1.0"

type VersionInfo struct {
	// Service name.
	Service string `json:"service"`

	// Current version value.
	Version string `json:"version"`
}

// HTTP handler for retrieving service version.
func Version(service string) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		res := VersionInfo{service, version}

		data, _ := json.Marshal(res)

		rw.Write(data)
	})
}

// Reads specified environment variable. If no value has been found,
// fallback is returned.
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

// Interface contains HTTP response specific methods.
type Response interface {
	// HTTP response code.
	Code() int

	// Map of HTTP headers with their values.
	Headers() map[string]string

	// Indicates if HTTP response has content.
	Empty() bool
}


