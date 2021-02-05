package configuration

import (
	"context"
	"crypto/sha1" // nolint:gosec
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

			// TODO extra labels and annotations
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
		err = h.resource.EnsureCreated(ctx, meta.Annotation.XObjectHash.Key(), configmap)
		if err != nil {
			return microerror.Mask(err)
		}

		err = h.resource.EnsureCreated(ctx, meta.Annotation.XObjectHash.Key(), secret)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Cleanup orphaned ConfigMap and Secret in case previous loop failed in between.
	{
		config, err = h.cleanupOrphanedConfig(ctx, config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Update Config CR status.
	{
		h.logger.Debugf(ctx, "updating Config status")

		current := &v1alpha1.Config{}

		err := h.k8sClient.CtrlClient().Get(ctx, k8sresource.ObjectKey(config), current)
		if err != nil {
			return microerror.Mask(err)
		}

		desiredStatus := *current.Status.DeepCopy()
		desiredStatus.App = v1alpha1.ConfigStatusApp(config.Spec.App)
		desiredStatus.Config.ConfigMapRef.Name = configmap.Name
		desiredStatus.Config.ConfigMapRef.Namespace = configmap.Namespace
		desiredStatus.Config.SecretRef.Name = secret.Name
		desiredStatus.Config.SecretRef.Namespace = secret.Namespace
		desiredStatus.Version = configVersion

		if reflect.DeepEqual(current.Status, desiredStatus) {
			h.logger.Debugf(ctx, "Config status already up to date")
		} else {
			current.Status = desiredStatus
			err := h.k8sClient.CtrlClient().Update(ctx, current)
			if err != nil {
				return microerror.Mask(err)
			}

			h.logger.Debugf(ctx, "updated Config status")
		}
	}

	// Cleanup orphaned ConfigMap and Secret.
	{
		config, err = h.cleanupOrphanedConfig(ctx, config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (h *Handler) cleanupOrphanedConfig(ctx context.Context, config *v1alpha1.Config) (*v1alpha1.Config, error) {
	// Get the most recent Config.
	err := h.k8sClient.CtrlClient().Get(ctx, k8sresource.ObjectKey(config), config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var needsCleanup bool
	var previousConfig v1alpha1.ConfigStatusConfig
	{
		c, ok, err := meta.Annotation.XPreviousConfig.Get(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// Annotation is there and the value is equal to the current
		// .status.config so nothing to do here. Return early.
		if ok && reflect.DeepEqual(config.Status.Config, c) {
			return config, nil
		}

		// If annotation exists but it isn't equal to .status.config
		// trigger the cleanup.
		if ok {
			needsCleanup = true
			previousConfig = c
		}
	}

	// Delete the old ConfigMap and Secret need to be deleted if they are
	// still there.
	if needsCleanup {
		configmap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      previousConfig.ConfigMapRef.Name,
				Namespace: previousConfig.ConfigMapRef.Namespace,
			},
		}

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      previousConfig.SecretRef.Name,
				Namespace: previousConfig.SecretRef.Namespace,
			},
		}

		h.logger.Debugf(ctx, "cleaning up orphaned ConfigMap %#q", k8sresource.ObjectKey(configmap))

		err = h.resource.EnsureDeleted(ctx, configmap)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		h.logger.Debugf(ctx, "cleaned up orphaned ConfigMap %#q", k8sresource.ObjectKey(configmap))

		h.logger.Debugf(ctx, "cleaning up orphaned Secret %#q", k8sresource.ObjectKey(secret))

		err = h.resource.EnsureDeleted(ctx, secret)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		h.logger.Debugf(ctx, "cleaned up orphaned Secret %#q", k8sresource.ObjectKey(secret))
	}

	// Now the ConfigMap and the Secret referenced by the annotation (if it
	// exists) are deleted. Update/set the annotation to the current status
	// value.
	{
		h.logger.Debugf(ctx, "updating %#q annotation", meta.Annotation.XPreviousConfig.Key())

		err = meta.Annotation.XPreviousConfig.Set(config, config.Status.Config)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		err = h.k8sClient.CtrlClient().Update(ctx, config)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		h.logger.Debugf(ctx, "updated %#q annotation", meta.Annotation.XPreviousConfig.Key())
	}

	// Try again. If the annotation and the .spec.config value are equal it
	// will return early with an up to date object.
	c, err := h.cleanupOrphanedConfig(ctx, config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return c, nil
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
