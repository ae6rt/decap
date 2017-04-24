package v1

import (
	"context"
	"errors"
)

var (
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
)

type KitVersion struct {
	BuildDate   string
	BuildCommit string
}

type Service interface {
	GetVersion(ctx context.Context) (KitVersion, error)
}

type DefaultService struct {
}

func (t DefaultService) GetVersion(ctx context.Context) (KitVersion, error) {
	return KitVersion{BuildDate: "now", BuildCommit: "abcde"}, nil
}

func NewService() Service {
	return DefaultService{}
}

type getVersionRequest struct {
}

type getVersionResponse struct {
	Version KitVersion `json:"version,omitempty"`
	Err     error      `json:"err,omitempty"`
}
