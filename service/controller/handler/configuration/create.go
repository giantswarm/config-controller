package configuration

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/service/controller/key"
)

func (h *Handler) EnsureCreated(ctx context.Context, obj interface{}) error {
	config, err := key.ToConfigCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	panic("not implemented")
}
