package controller

import (
	"fmt"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v8/pkg/controller"
	"github.com/giantswarm/operatorkit/v8/pkg/resource"
	"github.com/giantswarm/operatorkit/v8/pkg/resource/wrapper/metricsresource"
	vaultapi "github.com/hashicorp/vault/api"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/config-controller/internal/shared"

	"github.com/giantswarm/config-controller/api/v1alpha1"
	"github.com/giantswarm/config-controller/internal/meta"
	"github.com/giantswarm/config-controller/pkg/project"
	"github.com/giantswarm/config-controller/service/controller/handler/configuration"

	"github.com/giantswarm/config-controller/internal/ssh"
)

type ConfigConfig struct {
	K8sClient   k8sclient.Interface
	Logger      micrologger.Logger
	VaultClient *vaultapi.Client

	SharedConfigRepository  shared.ConfigRepository
	ConfigRepoSSHCredential ssh.Credential
	GitHubToken             string
	RepositoryName          string
	RepositoryRef           string
	Installation            string
	UniqueApp               bool
}

type Config struct {
	*controller.Controller
}

func NewConfig(config ConfigConfig) (*Config, error) {
	var err error

	resources, err := newConfigHandlers(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		selector, err := meta.Label.Version.Selector(config.UniqueApp)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() client.Object {
				return new(v1alpha1.Config)
			},
			Resources: resources,
			Selector:  selector,

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

func newConfigHandlers(config ConfigConfig) ([]resource.Interface, error) {
	var err error

	var configurationHandler resource.Interface
	{
		c := configuration.Config{
			Logger: config.Logger,

			K8sClient:   config.K8sClient,
			VaultClient: config.VaultClient,

			SharedConfigRepository:  config.SharedConfigRepository,
			ConfigRepoSSHCredential: config.ConfigRepoSSHCredential,
			GitHubToken:             config.GitHubToken,
			RepositoryName:          config.RepositoryName,
			RepositoryRef:           config.RepositoryRef,
			Installation:            config.Installation,
			UniqueApp:               config.UniqueApp,
		}

		fmt.Println("Lvl 2 CR Token: " + c.GitHubToken)
		fmt.Println("Lvl 2 CR SSH Key: " + c.ConfigRepoSSHCredential.Key)
		fmt.Println("Lvl 2 CR SSH Pw: " + c.ConfigRepoSSHCredential.Password)
		fmt.Println("Lvl 2 SCR SSH Name: " + c.SharedConfigRepository.Name)
		fmt.Println("Lvl 2 SCR SSH Ref: " + c.SharedConfigRepository.Ref)
		fmt.Println("Lvl 2 SCR SSH Key: " + c.SharedConfigRepository.Key)
		fmt.Println("Lvl 2 SCR SSH Pw: " + c.SharedConfigRepository.Password)

		configurationHandler, err = configuration.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	handlers := []resource.Interface{
		configurationHandler,
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
