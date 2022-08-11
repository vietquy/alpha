package http

import (
	"fmt"
	"net/http"

	"github.com/vietquy/alpha"
)

var (
	_ alpha.Response = (*removeRes)(nil)
	_ alpha.Response = (*thingRes)(nil)
	_ alpha.Response = (*viewThingRes)(nil)
	_ alpha.Response = (*thingsPageRes)(nil)
	_ alpha.Response = (*projectRes)(nil)
	_ alpha.Response = (*viewProjectRes)(nil)
	_ alpha.Response = (*projectsPageRes)(nil)
	_ alpha.Response = (*connectionRes)(nil)
	_ alpha.Response = (*disconnectionRes)(nil)
)

type removeRes struct{}

func (res removeRes) Code() int {
	return http.StatusNoContent
}

func (res removeRes) Headers() map[string]string {
	return map[string]string{}
}

func (res removeRes) Empty() bool {
	return true
}

type thingRes struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name,omitempty"`
	Key      string                 `json:"key"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	created  bool
}

func (res thingRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res thingRes) Headers() map[string]string {
	if res.created {
		return map[string]string{
			"Location":           fmt.Sprintf("/things/%s", res.ID),
			"Warning-Deprecated": "This endpoint will be depreciated in v1.0.0. It will be replaced with the bulk endpoint currently found at /things/bulk.",
		}
	}

	return map[string]string{}
}

func (res thingRes) Empty() bool {
	return true
}

type thingsRes struct {
	Things  []thingRes `json:"things"`
	created bool
}

func (res thingsRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res thingsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res thingsRes) Empty() bool {
	return false
}

type viewThingRes struct {
	ID       string                 `json:"id"`
	Owner    string                 `json:"-"`
	Name     string                 `json:"name,omitempty"`
	Key      string                 `json:"key"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (res viewThingRes) Code() int {
	return http.StatusOK
}

func (res viewThingRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewThingRes) Empty() bool {
	return false
}

type thingsPageRes struct {
	pageRes
	Things []viewThingRes `json:"things"`
}

func (res thingsPageRes) Code() int {
	return http.StatusOK
}

func (res thingsPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res thingsPageRes) Empty() bool {
	return false
}

type projectRes struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	created  bool
}

func (res projectRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res projectRes) Headers() map[string]string {
	if res.created {
		return map[string]string{
			"Location":           fmt.Sprintf("/projects/%s", res.ID),
			"Warning-Deprecated": "This endpoint will be depreciated in v1.0.0. It will be replaced with the bulk endpoint currently found at /projects/bulk.",
		}
	}

	return map[string]string{}
}

func (res projectRes) Empty() bool {
	return true
}

type projectsRes struct {
	Projects []projectRes `json:"projects"`
	created  bool
}

func (res projectsRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res projectsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res projectsRes) Empty() bool {
	return false
}

type viewProjectRes struct {
	ID       string                 `json:"id"`
	Owner    string                 `json:"-"`
	Name     string                 `json:"name,omitempty"`
	Things   []viewThingRes         `json:"connected,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (res viewProjectRes) Code() int {
	return http.StatusOK
}

func (res viewProjectRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewProjectRes) Empty() bool {
	return false
}

type projectsPageRes struct {
	pageRes
	Projects []viewProjectRes `json:"projects"`
}

func (res projectsPageRes) Code() int {
	return http.StatusOK
}

func (res projectsPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res projectsPageRes) Empty() bool {
	return false
}

type connectionRes struct{}

func (res connectionRes) Code() int {
	return http.StatusOK
}

func (res connectionRes) Headers() map[string]string {
	return map[string]string{
		"Warning-Deprecated": "This endpoint will be depreciated in v1.0.0. It will be replaced with the bulk endpoint found at /connect.",
	}
}

func (res connectionRes) Empty() bool {
	return true
}

type createConnectionsRes struct{}

func (res createConnectionsRes) Code() int {
	return http.StatusOK
}

func (res createConnectionsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res createConnectionsRes) Empty() bool {
	return true
}

type disconnectionRes struct{}

func (res disconnectionRes) Code() int {
	return http.StatusNoContent
}

func (res disconnectionRes) Headers() map[string]string {
	return map[string]string{}
}

func (res disconnectionRes) Empty() bool {
	return true
}

type pageRes struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}

type errorRes struct {
	Err string `json:"error"`
}
