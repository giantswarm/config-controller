package generator

import (
	"context"
	"github.com/giantswarm/config-controller/internal/shared"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/config-controller/internal/generator/github"
	"github.com/giantswarm/config-controller/internal/ssh"
	"github.com/giantswarm/config-controller/pkg/decrypt"
	"github.com/giantswarm/config-controller/pkg/generator"
	"github.com/giantswarm/config-controller/pkg/xstrings"
)

type Config struct {
	Log         micrologger.Logger
	VaultClient *vaultapi.Client

	SharedConfigRepository  shared.ConfigRepository
	ConfigRepoSSHCredential ssh.Credential
	GitHubToken             string
	RepositoryName          string
	RepositoryRef           string
	Installation            string
	Verbose                 bool
}

type Service struct {
	log              micrologger.Logger
	decryptTraverser generator.DecryptTraverser
	gitHub           *github.GitHub

	repositoryName string
	repositoryRef  string
	installation   string
	verbose        bool
}

func New(config Config) (*Service, error) {
	if config.VaultClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VaultClient must not be empty", config)
	}

	if config.GitHubToken == "" && config.ConfigRepoSSHCredential.IsEmpty() {
		return nil, microerror.Maskf(invalidConfigError, "%T.GitHubToken or %T.ConfigRepoSSHCredential must not be empty", config, config)
	}
	if config.RepositoryName == "" {
		// If repository name is not specified, fall back to original behaviour of using `giantswarm/config`
		config.RepositoryName = "config"
	}
	if config.RepositoryRef == "" {
		// If repository ref is not specified, fall back to using the main branch
		config.RepositoryName = "main"
	}
	if config.Installation == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Installation must not be empty", config)
	}

	var err error

	var decrypter *decrypt.VaultDecrypter
	{
		c := decrypt.VaultDecrypterConfig{
			VaultClient: config.VaultClient,
		}

		decrypter, err = decrypt.NewVaultDecrypter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var decryptTraverser *decrypt.YAMLTraverser
	{
		c := decrypt.YAMLTraverserConfig{
			Decrypter: decrypter,
		}

		decryptTraverser, err = decrypt.NewYAMLTraverser(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

	}

	var gitHub *github.GitHub
	{
		c := github.Config{
			SharedConfigRepository:  config.SharedConfigRepository,
			ConfigRepoSSHCredential: config.ConfigRepoSSHCredential,
			Token:                   config.GitHubToken,
		}

		gitHub, err = github.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		log:              config.Log,
		decryptTraverser: decryptTraverser,
		gitHub:           gitHub,

		repositoryName: config.RepositoryName,
		repositoryRef:  config.RepositoryRef,
		installation:   config.Installation,
		verbose:        config.Verbose,
	}

	return s, nil
}

type GenerateInput struct {
	// App for which the configuration is generated.
	App string

	// Name of the generated ConfigMap and Secret.
	Name string
	// Namespace of the generated ConfigMap and Secret.
	Namespace string

	// ExtraAnnotations are additional annotations to be set on the
	// generated ConfigMap and Secret. By default,
	// "config.giantswarm.io/version" annotation is set.
	ExtraAnnotations map[string]string
	// ExtraLabels are additional labels to be set on the generated
	// ConfigMap and Secret.
	ExtraLabels map[string]string
}

func (s *Service) Generate(ctx context.Context, in GenerateInput) (configmap *corev1.ConfigMap, secret *corev1.Secret, err error) {
	const (
		owner = "giantswarm"
	)

	var store github.Store

	store, err = s.gitHub.GetFilesByBranch(ctx, owner, s.repositoryName, s.repositoryRef)
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	var gen *generator.Generator
	{
		c := generator.Config{
			Fs:               store,
			DecryptTraverser: s.decryptTraverser,

			Installation: s.installation,
			Verbose:      s.verbose,
		}

		gen, err = generator.New(c)
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
	}

	annotations := xstrings.CopyMap(in.ExtraAnnotations)

	meta := metav1.ObjectMeta{
		Name:      in.Name,
		Namespace: in.Namespace,

		Annotations: annotations,
		Labels:      in.ExtraLabels,
	}

	configMap, secret, err := gen.GenerateConfig(ctx, in.App, meta)
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	return configMap, secret, nil
}
