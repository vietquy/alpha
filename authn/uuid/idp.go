// Package uuid provides a UUID identity provider.
package uuid

import (
	"github.com/gofrs/uuid"
	"github.com/vietquy/alpha/authn"
	"github.com/vietquy/alpha/errors"
)

// errGeneratingID indicates error in generating UUID
var errGeneratingID = errors.New("failed to generate uuid")

var _ authn.IdentityProvider = (*uuidIdentityProvider)(nil)

type uuidIdentityProvider struct{}

// New instantiates a UUID identity provider.
func New() authn.IdentityProvider {
	return &uuidIdentityProvider{}
}

func (idp *uuidIdentityProvider) ID() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", errors.Wrap(errGeneratingID, err)
	}

	return id.String(), nil
}
