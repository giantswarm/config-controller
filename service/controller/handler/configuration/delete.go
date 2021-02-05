package configuration

import (
	"context"
	"reflect"

	"github.com/giantswarm/config-controller/internal/meta"
	"github.com/giantswarm/config-controller/pkg/k8sresource"
	"github.com/giantswarm/config-controller/service/controller/key"
	"github.com/giantswarm/microerror"
)

func (h *Handler) EnsureDeleted(ctx context.Context, obj interface{}) error {
	config, err := key.ToConfigCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var toDelete []k8sresource.Object
	{
		configMap := configMapMeta(config.Status.Config)
		secret := secretMeta(config.Status.Config)

		h.logger.Debugf(ctx, "found ConfigMap %#q to delete", k8sresource.ObjectKey(configMap))
		h.logger.Debugf(ctx, "found Secret %#q to delete", k8sresource.ObjectKey(secret))

		toDelete = append(toDelete, configMap, secret)

		previousConfig, ok, err := meta.Annotation.XPreviousConfig.Get(config)
		if err != nil {
			return microerror.Mask(err)
		}
		if ok && !reflect.DeepEqual(config.Status.Config, previousConfig) {
			orphanedConfigMap := configMapMeta(previousConfig)
			orphanedSecret := secretMeta(previousConfig)

			h.logger.Debugf(ctx, "found orphaned ConfigMap %#q to delete", k8sresource.ObjectKey(orphanedConfigMap))
			h.logger.Debugf(ctx, "found orphaned Secret %#q to delete", k8sresource.ObjectKey(orphanedSecret))
		}
	}

	if len(toDelete) == 0 {
		h.logger.Debugf(ctx, "cancelling handler due to no found objects to cleanup")
	}

	// Cleanup.
	for _, o := range toDelete {
		err := h.resource.EnsureDeleted(ctx, o)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil

}
