package controller

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v4/pkg/controller"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"
	"k8s.io/apimachinery/pkg/runtime"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/giantswarm/config-controller/internal/meta"
	"github.com/giantswarm/config-controller/pkg/project"
	"github.com/giantswarm/config-controller/service/controller/handler/configuration"
)

type ConfigConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	GitHubToken  string
	UniqueConfig bool
	VaultClient  *vaultapi.Client
}

type Config struct {
	*controller.Controller
}

func NewConfig(config ConfigConfig) (*Config, error) {
	var err error

	resources, err := newConfigResources(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.Config)
			},
			Resources: resources,
			Selector:  meta.Label.Version.Selector(config.UniqueConfig),

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/config-controller-config-controller.
			Name: project.Name() + "-config-controller",
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Config{
		Controller: operatorkitController,
	}

	return c, nil
}

func newConfigResources(config ConfigConfig) ([]resource.Interface, error) {
	var err error

	var configurationHandler resource.Interface
	{
		c := configuration.Config{
			K8sClient:   config.K8sClient,
			Logger:      config.Logger,
			VaultClient: config.VaultClient,

			GitHubToken: config.GitHubToken,
		}

		configurationHandler, err = configuration.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	handlers := []resource.Interface{
		configurationHandler,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		handlers, err = retryresource.Wrap(handlers, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		handlers, err = metricsresource.Wrap(handlers, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return handlers, nil
}
