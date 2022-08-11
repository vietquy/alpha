package grpc

import "github.com/vietquy/alpha/things"

type AccessByKeyReq struct {
	thingKey string
	projectID   string
}

func (req AccessByKeyReq) validate() error {
	if req.projectID == "" || req.thingKey == "" {
		return things.ErrMalformedEntity
	}

	return nil
}

type accessByIDReq struct {
	thingID string
	projectID  string
}

func (req accessByIDReq) validate() error {
	if req.thingID == "" || req.projectID == "" {
		return things.ErrMalformedEntity
	}

	return nil
}

type identifyReq struct {
	key string
}

func (req identifyReq) validate() error {
	if req.key == "" {
		return things.ErrMalformedEntity
	}

	return nil
}
