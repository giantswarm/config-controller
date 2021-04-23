// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"

	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/config-controller/flag"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper
}

type Service struct {
	Version *version.Service
}

// New creates a new configured service object.
func New(config Config) (*Service, error) {
	return &Service{}, nil
}

func (s *Service) Boot(ctx context.Context) {
}
