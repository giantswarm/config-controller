package values

import (
	"context"
	"fmt"

	"github.com/giantswarm/config-controller/service/controller/key"
	"github.com/giantswarm/microerror"
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

	// TODO Check if configVersion is branch or tag

	return nil
}
