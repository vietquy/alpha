package api

import (
	"github.com/vietquy/alpha/reader"
)

type apiReq interface {
	validate() error
}

type listMessagesReq struct {
	projectID   string
	pageMeta reader.PageMetadata
}

func (req listMessagesReq) validate() error {
	if req.pageMeta.Limit < 1 || req.pageMeta.Offset < 0 {
		return errInvalidQueryParams
	}
	if req.pageMeta.Comparator != "" &&
		req.pageMeta.Comparator != reader.EqualKey &&
		req.pageMeta.Comparator != reader.LowerThanKey &&
		req.pageMeta.Comparator != reader.LowerThanEqualKey &&
		req.pageMeta.Comparator != reader.GreaterThanKey &&
		req.pageMeta.Comparator != reader.GreaterThanEqualKey {
		return errInvalidQueryParams
	}

	return nil
}
