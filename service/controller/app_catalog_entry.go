package controller

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v4/pkg/controller"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"
	"k8s.io/apimachinery/pkg/runtime"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/giantswarm/config-controller/pkg/label"
	"github.com/giantswarm/config-controller/pkg/project"
	"github.com/giantswarm/config-controller/service/controller/handler/configversion"
)

type AppCatalogEntryConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	GitHubToken  string
	Installation string
	UniqueApp    bool
	VaultClient  *vaultapi.Client
}

type AppCatalogEntry struct {
	*controller.Controller
}

func NewAppCatalogEntry(config AppCatalogEntryConfig) (*AppCatalogEntry, error) {
	var err error

	resources, err := newAppCatalogEntryResources(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() runtime.Object {
				// This should become AppCatalogEntry in the
				// future. See:
				//
				//	https://github.com/giantswarm/giantswarm/issues/14832
				//
				return new(v1alpha1.App)
			},
			Resources: resources,
			Selector:  label.AppVersionSelector(config.UniqueApp),

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/config-controller-app-controller.
			Name: project.Name() + "-app-catalog-entry-controller",
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &AppCatalogEntry{
		Controller: operatorkitController,
	}

	return c, nil
}

func newAppCatalogEntryResources(config AppCatalogEntryConfig) ([]resource.Interface, error) {
	var err error

	var configVersionResource resource.Interface
	{
		c := configversion.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		configVersionResource, err = configversion.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		configVersionResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}
