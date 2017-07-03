package app

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	GetVersionEndpoint endpoint.Endpoint
}

func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		GetVersionEndpoint: MakeGetVersionEndpoint(s),
	}
}

func MakeGetVersionEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		p, e := s.GetVersion(ctx)
		return getVersionResponse{Version: p, Err: e}, nil
	}
}
