package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/vietquy/alpha"
	"github.com/vietquy/alpha/errors"
	"github.com/vietquy/alpha/things"
)

const (
	contentType = "application/json"
	offset      = "offset"
	limit       = "limit"
	name        = "name"
	metadata    = "metadata"

	defOffset = 0
	defLimit  = 10
)

var (
	errUnsupportedContentType = errors.New("unsupported content type")
	errInvalidQueryParams     = errors.New("invalid query params")
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc things.Service) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	r := bone.New()

	r.Post("/things", kithttp.NewServer(
		createThingEndpoint(svc),
		decodeThingCreation,
		encodeResponse,
		opts...,
	))

	r.Post("/things/bulk", kithttp.NewServer(
		createThingsEndpoint(svc),
		decodeThingsCreation,
		encodeResponse,
		opts...,
	))

	r.Patch("/things/:id/key", kithttp.NewServer(
		updateKeyEndpoint(svc),
		decodeKeyUpdate,
		encodeResponse,
		opts...,
	))

	r.Put("/things/:id", kithttp.NewServer(
		updateThingEndpoint(svc),
		decodeThingUpdate,
		encodeResponse,
		opts...,
	))

	r.Delete("/things/:id", kithttp.NewServer(
		removeThingEndpoint(svc),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Get("/things/:id", kithttp.NewServer(
		viewThingEndpoint(svc),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Get("/things/:id/projects", kithttp.NewServer(
		listProjectsByThingEndpoint(svc),
		decodeListByConnection,
		encodeResponse,
		opts...,
	))

	r.Get("/things", kithttp.NewServer(
		listThingsEndpoint(svc),
		decodeList,
		encodeResponse,
		opts...,
	))

	r.Post("/projects", kithttp.NewServer(
		createProjectEndpoint(svc),
		decodeProjectCreation,
		encodeResponse,
		opts...,
	))

	r.Post("/projects/bulk", kithttp.NewServer(
		createProjectsEndpoint(svc),
		decodeProjectsCreation,
		encodeResponse,
		opts...,
	))

	r.Put("/projects/:id", kithttp.NewServer(
		updateProjectEndpoint(svc),
		decodeProjectUpdate,
		encodeResponse,
		opts...,
	))

	r.Delete("/projects/:id", kithttp.NewServer(
		removeProjectEndpoint(svc),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Get("/projects/:id", kithttp.NewServer(
		viewProjectEndpoint(svc),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Get("/projects/:id/things", kithttp.NewServer(
		listThingsByProjectEndpoint(svc),
		decodeListByConnection,
		encodeResponse,
		opts...,
	))

	r.Get("/projects", kithttp.NewServer(
		listProjectsEndpoint(svc),
		decodeList,
		encodeResponse,
		opts...,
	))

	r.Put("/projects/:projectId/things/:thingId", kithttp.NewServer(
		connectEndpoint(svc),
		decodeConnection,
		encodeResponse,
		opts...,
	))

	r.Post("/connect", kithttp.NewServer(
		createConnectionsEndpoint(svc),
		decodeCreateConnections,
		encodeResponse,
		opts...,
	))

	r.Delete("/projects/:projectId/things/:thingId", kithttp.NewServer(
		disconnectEndpoint(svc),
		decodeConnection,
		encodeResponse,
		opts...,
	))

	r.GetFunc("/version", alpha.Version("things"))

	return r
}

func decodeThingCreation(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := createThingReq{token: r.Header.Get("Authorization")}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(things.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeThingsCreation(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := createThingsReq{token: r.Header.Get("Authorization")}
	if err := json.NewDecoder(r.Body).Decode(&req.Things); err != nil {
		return nil, errors.Wrap(things.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeThingUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := updateThingReq{
		token: r.Header.Get("Authorization"),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(things.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeKeyUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := updateKeyReq{
		token: r.Header.Get("Authorization"),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(things.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeProjectCreation(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := createProjectReq{token: r.Header.Get("Authorization")}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(things.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeProjectsCreation(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := createProjectsReq{token: r.Header.Get("Authorization")}

	if err := json.NewDecoder(r.Body).Decode(&req.Projects); err != nil {
		return nil, errors.Wrap(things.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeProjectUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := updateProjectReq{
		token: r.Header.Get("Authorization"),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(things.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeView(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewResourceReq{
		token: r.Header.Get("Authorization"),
		id:    bone.GetValue(r, "id"),
	}

	return req, nil
}

func decodeList(_ context.Context, r *http.Request) (interface{}, error) {
	o, err := readUintQuery(r, offset, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := readUintQuery(r, limit, defLimit)
	if err != nil {
		return nil, err
	}

	n, err := readStringQuery(r, name)
	if err != nil {
		return nil, err
	}

	m, err := readMetadataQuery(r, "metadata")
	if err != nil {
		return nil, err
	}

	req := listResourcesReq{
		token:    r.Header.Get("Authorization"),
		offset:   o,
		limit:    l,
		name:     n,
		metadata: m,
	}

	return req, nil
}

func decodeListByConnection(_ context.Context, r *http.Request) (interface{}, error) {
	o, err := readUintQuery(r, offset, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := readUintQuery(r, limit, defLimit)
	if err != nil {
		return nil, err
	}

	req := listByConnectionReq{
		token:  r.Header.Get("Authorization"),
		id:     bone.GetValue(r, "id"),
		offset: o,
		limit:  l,
	}

	return req, nil
}

func decodeConnection(_ context.Context, r *http.Request) (interface{}, error) {
	req := connectionReq{
		token:   r.Header.Get("Authorization"),
		projectID:  bone.GetValue(r, "projectId"),
		thingID: bone.GetValue(r, "thingId"),
	}

	return req, nil
}

func decodeCreateConnections(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := createConnectionsReq{token: r.Header.Get("Authorization")}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(things.ErrMalformedEntity, err)
	}

	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", contentType)

	if ar, ok := response.(alpha.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}

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
		case errors.Contains(errorVal, things.ErrMalformedEntity):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, things.ErrUnauthorizedAccess):
			w.WriteHeader(http.StatusForbidden)
		case errors.Contains(errorVal, things.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Contains(errorVal, things.ErrConflict):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Contains(errorVal, errUnsupportedContentType):
			w.WriteHeader(http.StatusUnsupportedMediaType)
		case errors.Contains(errorVal, errInvalidQueryParams):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, io.ErrUnexpectedEOF):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, io.EOF):
			w.WriteHeader(http.StatusBadRequest)
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

func readUintQuery(r *http.Request, key string, def uint64) (uint64, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return 0, errInvalidQueryParams
	}

	if len(vals) == 0 {
		return def, nil
	}

	strval := vals[0]
	val, err := strconv.ParseUint(strval, 10, 64)
	if err != nil {
		return 0, errInvalidQueryParams
	}

	return val, nil
}

func readStringQuery(r *http.Request, key string) (string, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return "", errInvalidQueryParams
	}

	if len(vals) == 0 {
		return "", nil
	}

	return vals[0], nil
}

func readMetadataQuery(r *http.Request, key string) (map[string]interface{}, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return nil, errInvalidQueryParams
	}

	if len(vals) == 0 {
		return nil, nil
	}

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(vals[0]), &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
