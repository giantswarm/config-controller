package values

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/config-controller/pkg/generator/key"
	controllerkey "github.com/giantswarm/config-controller/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	app, err := controllerkey.ToAppCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	configVersion, ok := app.GetAnnotations()[key.ConfigVersion]
	if !ok {
		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("App CR %q is missing %q annotation", app.Name, key.ConfigVersion))
		return nil
	}

	appAndVersion := fmt.Sprintf("App %#q, config version %#q", app.Spec.Name, configVersion)
	r.logger.LogCtx(ctx, "level", "debug", "message", "deleting "+appAndVersion)

	deleteOpts := client.MatchingLabels{
		key.ConfigVersion: configVersion,
		key.AppLabel:      app.Name,
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "deleting configmap for "+appAndVersion)
	err = r.k8sClient.CtrlClient().DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace(app.Namespace), deleteOpts)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", "deleted configmap for "+appAndVersion)

	r.logger.LogCtx(ctx, "level", "debug", "message", "deleting secret for "+appAndVersion)
	err = r.k8sClient.CtrlClient().DeleteAllOf(ctx, &corev1.Secret{}, client.InNamespace(app.Namespace), deleteOpts)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", "deleted secret for "+appAndVersion)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("clearing %s configmap and secret details", appAndVersion))
	app.Spec.Config = v1alpha1.AppSpecConfig{}
	err = r.k8sClient.CtrlClient().Update(ctx, &app)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cleared %s configmap and secret details", appAndVersion))

	r.logger.LogCtx(ctx, "level", "debug", "message", "deleted "+appAndVersion)

	return nil
}
