package values

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("App CR %q is missing %q annotation", app.Name, key.ConfigVersion))
		r.logger.LogCtx(ctx, "level", "debug", "message", "cancelling resource")
		return nil
	}

	appAndVersion := fmt.Sprintf("App %#q, config version %#q", app.Spec.Name, configVersion)
	r.logger.LogCtx(ctx, "level", "debug", "message", "deleting "+appAndVersion)

	if app.Spec.Config.ConfigMap.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting configmap for "+appAndVersion)
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      app.Spec.Config.ConfigMap.Name,
				Namespace: app.Spec.Config.ConfigMap.Namespace,
			},
		}
		err = r.k8sClient.CtrlClient().Delete(ctx, cm)
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted configmap for "+appAndVersion)
	}

	if app.Spec.Config.Secret.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting secret for "+appAndVersion)
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      app.Spec.Config.ConfigMap.Name,
				Namespace: app.Spec.Config.ConfigMap.Namespace,
			},
		}
		err = r.k8sClient.CtrlClient().Delete(ctx, secret)
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted secret for "+appAndVersion)
	}

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
