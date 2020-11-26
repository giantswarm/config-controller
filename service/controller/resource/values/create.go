package values

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/config-controller/pkg/generator/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	app, err := ToAppCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	configVersion, ok := app.GetAnnotations()[key.ConfigVersion]
	if !ok {
		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("App CR %q is missing %q annotation", app.Name, key.ConfigVersion))
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("generating app %#q config version %#q", app.Spec.Name, configVersion))
	configmap, secret, err := r.GenerateConfig(ctx, key.Owner, r.installation, app.Namespace, app.Spec.Name, configVersion)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("generated app %#q config version %#q", app.Spec.Name, configVersion))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating configmap %s/%s", configmap.Namespace, configmap.Name))
	err = r.k8sClient.CtrlClient().Create(context.Background(), configmap)
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("configmap %s/%s already exists", configmap.Namespace, configmap.Name))
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created configmap %s/%s", configmap.Namespace, configmap.Name))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating secret %s/%s", secret.Namespace, secret.Name))
	err = r.k8sClient.CtrlClient().Create(context.Background(), secret)
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("secret %s/%s already exists", secret.Namespace, secret.Name))
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created secret %s/%s", secret.Namespace, secret.Name))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating App CR %#q with configmap and secret details", app.Name))
	app.Spec.Config.ConfigMap = v1alpha1.AppSpecConfigConfigMap{
		Namespace: configmap.Namespace,
		Name:      configmap.Name,
	}
	app.Spec.Config.Secret = v1alpha1.AppSpecConfigSecret{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}
	err = r.k8sClient.CtrlClient().Update(context.Background(), secret)
	if err == nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated App CR %#q with configmap and secret details", app.Name))

	return nil
}
