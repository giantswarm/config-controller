package values

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	app, err := ToAppCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	configVersion, ok := app.GetAnnotations()[key.ConfigVersionAnnotation]
	if !ok {
		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("App CR %q is missing %q annotation", app.Name, key.ConfigVersionAnnotation))
		return nil
	}

	configmap, secret, err := r.GenerateConfig(ctx)

	return nil
}
