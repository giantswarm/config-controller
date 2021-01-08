package values

import (
	"context"
	"reflect"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/config-controller/service/controller/key"
)

func (h *Handler) EnsureCreated(ctx context.Context, obj interface{}) error {
	app, err := key.ToAppCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	annotations := app.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	configVersion, ok := annotations[annotation.ConfigVersion]
	if !ok {
		h.logger.Debugf(ctx, "App CR is missing %q annotation", annotation.ConfigVersion)
		h.logger.Debugf(ctx, "cancelling handler")
		return nil
	}

	if configVersion == "0.0.0" {
		h.logger.Debugf(ctx, "App CR has config version %#q", configVersion)
		h.logger.Debugf(ctx, "cancelling handler")
	}

	h.logger.Debugf(ctx, "generating app config version %#q", configVersion)
	configmap, secret, err := h.generateConfig(ctx, h.installation, app.Namespace, app.Spec.Name, configVersion)
	if err != nil {
		return microerror.Mask(err)
	}
	h.logger.Debugf(ctx, "generated app config version %#q", configVersion)

	h.logger.Debugf(ctx, "ensuring configmap %s/%s", configmap.Namespace, configmap.Name)
	err = h.k8sClient.CtrlClient().Create(ctx, configmap)
	if apierrors.IsAlreadyExists(err) {
		err = h.k8sClient.CtrlClient().Update(ctx, configmap)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if err != nil {
		return microerror.Mask(err)
	}
	h.logger.Debugf(ctx, "ensured configmap %s/%s", configmap.Namespace, configmap.Name)

	h.logger.Debugf(ctx, "ensuring secret %s/%s", secret.Namespace, secret.Name)
	err = h.k8sClient.CtrlClient().Create(ctx, secret)
	if apierrors.IsAlreadyExists(err) {
		err = h.k8sClient.CtrlClient().Update(ctx, secret)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if err != nil {
		return microerror.Mask(err)
	}
	h.logger.Debugf(ctx, "ensured secret %s/%s", secret.Namespace, secret.Name)

	configmapReference := v1alpha1.AppSpecConfigConfigMap{
		Namespace: configmap.Namespace,
		Name:      configmap.Name,
	}
	secretReference := v1alpha1.AppSpecConfigSecret{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}
	if !reflect.DeepEqual(app.Spec.Config.ConfigMap, configmapReference) || !reflect.DeepEqual(app.Spec.Config.Secret, secretReference) {
		h.logger.Debugf(ctx, "updating App CR with configmap and secret details")
		app.SetAnnotations(key.RemoveAnnotation(annotations, key.PauseAnnotation))
		app.Spec.Config.ConfigMap = configmapReference
		app.Spec.Config.Secret = secretReference
		err = h.k8sClient.CtrlClient().Update(ctx, &app)
		if err != nil {
			return microerror.Mask(err)
		}
		h.logger.Debugf(ctx, "updated App CR with configmap and secret details")
	}

	return nil
}
