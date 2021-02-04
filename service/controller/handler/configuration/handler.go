package configuration

import (
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	vaultapi "github.com/hashicorp/vault/api"

	"github.com/giantswarm/config-controller/internal/configversion"
	"github.com/giantswarm/config-controller/internal/generator"
	"github.com/giantswarm/config-controller/pkg/k8sresource"
)

const (
	Name = "configuration"
)

type Config struct {
	K8sClient   k8sclient.Interface
	Logger      micrologger.Logger
	VaultClient *vaultapi.Client

	GitHubToken  string
	Installation string
}

type Handler struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	configVersion *configversion.Service
	generator     *generator.Service
	resource      *k8sresource.Service
}

func New(config Config) (*Handler, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.VaultClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VaultClient must not be empty", config)
	}

	if config.GitHubToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GitHubToken must not be empty", config)
	}
	if config.Installation == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Installation must not be empty", config)
	}

	var err error

	var configVersion *configversion.Service
	{
		c := configversion.Config{
			K8sClient: config.K8sClient,
		}

		configVersion, err = configversion.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var gen *generator.Service
	{
		c := generator.Config{
			VaultClient: config.VaultClient,

			GitHubToken:  config.GitHubToken,
			Installation: config.Installation,
		}

		gen, err = generator.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resource *k8sresource.Service
	{
		c := k8sresource.Config{
			Client: config.K8sClient,
			Logger: config.Logger,
		}

		resource, err = k8sresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

	}

	h := &Handler{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		configVersion: configVersion,
		generator:     gen,
		resource:      resource,
	}

	return h, nil
}

func (h *Handler) Name() string {
	return Name
}
