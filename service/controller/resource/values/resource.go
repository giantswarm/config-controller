package values

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/config-controller/pkg/decrypt"
	"github.com/giantswarm/config-controller/pkg/generator"
	"github.com/giantswarm/config-controller/pkg/github"
)

const (
	Name       = "values"
	ConfigRepo = "config"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	VaultClient  *vaultapi.Client
	GitHubToken  string
	Installation string
}

type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	decryptTraverser *decrypt.YAMLTraverser
	gitHubToken      string
	installation     string
}

func New(config Config) (*Resource, error) {
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

	var decryptTraverser *decrypt.YAMLTraverser
	{
		c := decrypt.VaultDecrypterConfig{
			VaultClient: config.VaultClient,
		}

		decrypter, err := decrypt.NewVaultDecrypter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		decryptTraverser, err = decrypt.NewYAMLTraverser(
			decrypt.YAMLTraverserConfig{
				Decrypter: decrypter,
			},
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	r := &Resource{
		k8sClient:        config.K8sClient,
		logger:           config.Logger,
		decryptTraverser: decryptTraverser,
		gitHubToken:      config.GitHubToken,
		installation:     config.Installation,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) GenerateConfig(ctx context.Context, owner, installation, namespace, app, configVersion string) (configmap *corev1.ConfigMap, secret *corev1.Secret, err error) {
	var store generator.Filesystem
	var ref string
	{
		gh, err := github.New(github.Config{
			Token: r.gitHubToken,
		})
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}

		if configVersion == "" {
			return nil, nil, microerror.Maskf(executionFailedError, "configVersion must be defined")
		}

		isTag, err := gh.ResolvesToTag(ctx, owner, app, configVersion)
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}

		if isTag {
			tag, err := gh.GetLatestTag(ctx, owner, ConfigRepo, configVersion)
			if err != nil {
				return nil, nil, microerror.Mask(err)
			}

			store, err = gh.GetFilesByTag(ctx, owner, ConfigRepo, tag)
			if err != nil {
				return nil, nil, microerror.Mask(err)
			}
			ref = tag
		} else {
			store, err = gh.GetFilesByBranch(ctx, owner, ConfigRepo, configVersion)
			if err != nil {
				return nil, nil, microerror.Mask(err)
			}
			ref = configVersion
		}
	}

	gen, err := generator.New(&generator.Config{
		Fs:               store,
		DecryptTraverser: r.decryptTraverser,
	})
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	configmap, secret, err = gen.GenerateConfig(ctx, installation, app, namespace, ref)
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	return configmap, secret, nil
}

func ToAppCR(v interface{}) (v1alpha1.App, error) {
	if v == nil {
		return v1alpha1.App{}, microerror.Maskf(wrongTypeError, "expected non-nil, got %#v", v)
	}

	p, ok := v.(*v1alpha1.App)
	if !ok {
		return v1alpha1.App{}, microerror.Maskf(wrongTypeError, "expected %T, got %T", p, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
