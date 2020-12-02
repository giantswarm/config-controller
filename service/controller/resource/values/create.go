package values

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	controllerkey "github.com/giantswarm/config-controller/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	app, err := controllerkey.ToAppCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	configVersion, ok := app.GetAnnotations()[annotation.ConfigVersion]
	if !ok {
		r.logger.Debugf(ctx, "App CR %q is missing %q annotation", app.Name, annotation.ConfigVersion)
		r.logger.Debugf(ctx, "cancelling resource")
		return nil
	}

	r.logger.Debugf(ctx, "generating app %#q config version %#q", app.Spec.Name, configVersion)
	configmap, secret, err := r.generateConfig(ctx, r.installation, app.Namespace, app.Spec.Name, configVersion)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.Debugf(ctx, "generated app %#q config version %#q", app.Spec.Name, configVersion)

	r.logger.Debugf(ctx, "creating configmap %s/%s", configmap.Namespace, configmap.Name)
	err = r.k8sClient.CtrlClient().Create(ctx, configmap)
	if apierrors.IsAlreadyExists(err) {
		r.logger.Debugf(ctx, "configmap %s/%s already exists", configmap.Namespace, configmap.Name)
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.Debugf(ctx, "created configmap %s/%s", configmap.Namespace, configmap.Name)

	r.logger.Debugf(ctx, "creating secret %s/%s", secret.Namespace, secret.Name)
	err = r.k8sClient.CtrlClient().Create(ctx, secret)
	if apierrors.IsAlreadyExists(err) {
		r.logger.Debugf(ctx, "secret %s/%s already exists", secret.Namespace, secret.Name)
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.Debugf(ctx, "created secret %s/%s", secret.Namespace, secret.Name)

	r.logger.Debugf(ctx, "updating App CR %#q with configmap and secret details", app.Name)
	app.Spec.Config.ConfigMap = v1alpha1.AppSpecConfigConfigMap{
		Namespace: configmap.Namespace,
		Name:      configmap.Name,
	}
	app.Spec.Config.Secret = v1alpha1.AppSpecConfigSecret{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}
	err = r.k8sClient.CtrlClient().Update(ctx, &app)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.Debugf(ctx, "updated App CR %#q with configmap and secret details", app.Name)

	return nil
}
