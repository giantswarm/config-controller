package configuration

import (
	"context"
	"crypto/sha1" // nolint:gosec
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/config-controller/internal/generator"
	"github.com/giantswarm/config-controller/internal/meta"
	"github.com/giantswarm/config-controller/pkg/k8sresource"
	"github.com/giantswarm/config-controller/service/controller/key"
)

func (h *Handler) EnsureCreated(ctx context.Context, obj interface{}) error {
	config, err := key.ToConfigCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var configVersion string
	{
		cav := config.Spec.App.Catalog + "/" + config.Spec.App.Name + "@" + config.Spec.App.Version

		h.logger.Debugf(ctx, "getting config version for App %#q", cav)

		configVersion, err = h.configVersion.Get(ctx, config.Spec.App)
		if err != nil {
			return microerror.Mask(err)
		}

		h.logger.Debugf(ctx, "got config version for App %#q = %#q", cav, configVersion)
	}

	var configmap *corev1.ConfigMap
	var secret *corev1.Secret
	{
		name, err := genStableObjectName(config)
		if err != nil {
			return microerror.Mask(err)
		}

		namespace := config.Namespace

		nn := namespace + "/" + name

		h.logger.Debugf(ctx, "generating %#q ConfigMap and Secret for config version %#q", nn, configVersion)

		configmap, secret, err = h.generator.Generate(ctx, generator.GenerateInput{
			App:           config.Spec.App.Name,
			ConfigVersion: configVersion,

			Name:      name,
			Namespace: namespace,

			ExtraAnnotations: map[string]string{},
			ExtraLabels:      map[string]string{},
		})
		if err != nil {
			return microerror.Mask(err)
		}

		h.logger.Debugf(ctx, "generated %#q ConfigMap and Secret for config version %#q", nn, configVersion)
	}

	// Ensure ConfigMap and Secret.
	{
		err = h.resource.EnsureCreated(ctx, meta.Annotation.XObjectHash(), configmap)
		if err != nil {
			return microerror.Mask(err)
		}

		err = h.resource.EnsureCreated(ctx, meta.Annotation.XObjectHash(), secret)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Update Config CR status.
	{
		var current *v1alpha1.Config

		modifyFunc := func() error {
			current.Status.App = v1alpha1.ConfigStatusApp(config.Spec.App)
			current.Status.Config.ConfigMapRef.Name = configmap.Name
			current.Status.Config.ConfigMapRef.Namespace = configmap.Namespace
			current.Status.Config.SecretRef.Name = secret.Name
			current.Status.Config.SecretRef.Namespace = secret.Namespace
			current.Status.Version = configVersion

			return nil
		}

		err = h.resource.ModifyStatus(ctx, k8sresource.ObjectKey(config), current, modifyFunc, nil)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func genStableObjectName(config *v1alpha1.Config) (string, error) {
	h, err := hash(config.Spec.App)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return setSuffixMax63(config.Name, h), nil
}

func hash(v interface{}) (string, error) {
	bs, err := json.Marshal(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	sum := sha1.Sum(bs) // nolint:gosec
	return fmt.Sprintf("%x", sum)[:10], nil
}

func setSuffixMax63(s string, suffix string) string {
	maxLen := 63

	if len(s)+len(suffix)+1 <= maxLen {
		return s + "-" + suffix
	}

	return s[:maxLen-len(suffix)-1] + "-" + suffix
}
