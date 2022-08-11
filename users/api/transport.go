package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/vietquy/alpha/errors"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/vietquy/alpha"
	log "github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/users"
)

const contentType = "application/json"

var (
	// ErrUnsupportedContentType indicates unacceptable or lack of Content-Type
	ErrUnsupportedContentType = errors.New("unsupported content type")
	errMissingRefererHeader   = errors.New("missing referer header")
	errInvalidToken           = errors.New("invalid token")
	errNoTokenSupplied        = errors.New("no token supplied")
	// ErrFailedDecode indicates failed to decode request body
	ErrFailedDecode = errors.New("failed to decode request body")
	logger          log.Logger
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc users.Service, l log.Logger) http.Handler {
	logger = l

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	mux := bone.New()

	mux.Post("/users", kithttp.NewServer(
		registrationEndpoint(svc),
		decodeCredentials,
		encodeResponse,
		opts...,
	))

	mux.Get("/users", kithttp.NewServer(
		userInfoEndpoint(svc),
		decodeViewInfo,
		encodeResponse,
		opts...,
	))

	mux.Put("/users", kithttp.NewServer(
		updateUserEndpoint(svc),
		decodeUpdateUser,
		encodeResponse,
		opts...,
	))

	mux.Patch("/password", kithttp.NewServer(
		passwordChangeEndpoint(svc),
		decodePasswordChange,
		encodeResponse,
		opts...,
	))

	mux.Post("/tokens", kithttp.NewServer(
		loginEndpoint(svc),
		decodeCredentials,
		encodeResponse,
		opts...,
	))

	mux.GetFunc("/version", alpha.Version("users"))

	return mux
}

func decodeViewInfo(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewUserInfoReq{
		token: r.Header.Get("Authorization"),
	}
	return req, nil
}

func decodeUpdateUser(_ context.Context, r *http.Request) (interface{}, error) {
	var req updateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	req.token = r.Header.Get("Authorization")
	return req, nil
}

func decodeCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}

	var user users.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return nil, errors.Wrap(users.ErrMalformedEntity, err)
	}

	return userReq{user}, nil
}

func decodePasswordResetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}

	var req passwResetReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(ErrFailedDecode, err)
	}

	req.Host = r.Header.Get("Referer")
	return req, nil
}

func decodePasswordReset(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}

	var req resetTokenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(ErrFailedDecode, err)
	}

	return req, nil
}

func decodePasswordChange(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, ErrUnsupportedContentType
	}

	var req passwChangeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(ErrFailedDecode, err)
	}

	req.Token = r.Header.Get("Authorization")

	return req, nil
}

func decodeToken(_ context.Context, r *http.Request) (interface{}, error) {
	vals := bone.GetQuery(r, "token")
	if len(vals) > 1 {
		return "", errInvalidToken
	}

	if len(vals) == 0 {
		return "", errNoTokenSupplied
	}
	t := vals[0]
	return t, nil

}
func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	if ar, ok := response.(alpha.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch errorVal := err.(type) {
	case errors.Error:
		w.Header().Set("Content-Type", contentType)
		switch {
		case errors.Contains(errorVal, users.ErrMalformedEntity):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, users.ErrUnauthorizedAccess):
			w.WriteHeader(http.StatusForbidden)
		case errors.Contains(errorVal, users.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		case errors.Contains(errorVal, ErrUnsupportedContentType):
			w.WriteHeader(http.StatusUnsupportedMediaType)
		case errors.Contains(errorVal, ErrFailedDecode):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, io.ErrUnexpectedEOF):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, io.EOF):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, users.ErrUserNotFound):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, users.ErrRecoveryToken):
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		if errorVal.Msg() != "" {
			if err := json.NewEncoder(w).Encode(errorRes{Err: errorVal.Msg()}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
