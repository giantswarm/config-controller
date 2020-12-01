package values

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	controllerkey "github.com/giantswarm/config-controller/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	app, err := controllerkey.ToAppCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	configVersion, ok := app.GetAnnotations()[annotation.ConfigMajorVersion]
	if !ok {
		r.logger.Debugf(ctx, "App CR %q is missing %q annotation", app.Name, annotation.ConfigMajorVersion)
		r.logger.Debugf(ctx, "cancelling resource")
		return nil
	}

	r.logger.Debugf(ctx, "deleting App %#q, config version %#q", app.Spec.Name, configVersion)

	if app.Spec.Config.ConfigMap.Name != "" {
		r.logger.Debugf(ctx, "deleting configmap for App %#q, config version %#q", app.Spec.Name, configVersion)
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      app.Spec.Config.ConfigMap.Name,
				Namespace: app.Spec.Config.ConfigMap.Namespace,
			},
		}
		err = r.k8sClient.CtrlClient().Delete(ctx, cm)
		if client.IgnoreNotFound(err) != nil {
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "deleted configmap for App %#q, config version %#q", app.Spec.Name, configVersion)
	}

	if app.Spec.Config.Secret.Name != "" {
		r.logger.Debugf(ctx, "deleting secret for App %#q, config version %#q", app.Spec.Name, configVersion)
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      app.Spec.Config.ConfigMap.Name,
				Namespace: app.Spec.Config.ConfigMap.Namespace,
			},
		}
		err = r.k8sClient.CtrlClient().Delete(ctx, secret)
		if client.IgnoreNotFound(err) != nil {
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "deleted secret for App %#q, config version %#q", app.Spec.Name, configVersion)
	}

	r.logger.Debugf(ctx, "clearing App %#q, config version %#q configmap and secret details", app.Spec.Name, configVersion)
	app.Spec.Config = v1alpha1.AppSpecConfig{}
	err = r.k8sClient.CtrlClient().Update(ctx, &app)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.Debugf(ctx, "cleared App %#q, config version %#q configmap and secret details", app.Spec.Name, configVersion)

	r.logger.Debugf(ctx, "deleted App %#q, config version %#q", app.Spec.Name, configVersion)

	return nil
}
