package generator

import (
	"context"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/config-controller/internal/generator/github"
	"github.com/giantswarm/config-controller/pkg/decrypt"
	"github.com/giantswarm/config-controller/pkg/generator"
)

type Config struct {
	Log         micrologger.Logger
	VaultClient *vaultapi.Client

	Installation string
	Verbose      bool
}

type Service struct {
	log              micrologger.Logger
	decryptTraverser generator.DecryptTraverser

	installation string
	verbose      bool
}

func New(config Config) (*Service, error) {
	if config.VaultClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VaultClient must not be empty", config)
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

	s := &Service{
		log:              config.Log,
		decryptTraverser: decryptTraverser,

		installation: config.Installation,
		verbose:      config.Verbose,
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
	// generated ConfigMap and Secret. By default
	// "config.giantswarm.io/version" annotation is set.
	ExtraAnnotations map[string]string
	// ExtraLabels are additional labels to be set on the generated
	// ConfigMap and Secret.
	ExtraLabels map[string]string
}

type fsStore struct{}

func (*fsStore) ReadFile(path string) ([]byte, error) {
	bs, err := os.ReadFile(path)
	return bs, microerror.Mask(err)
}

func (*fsStore) ReadDir(dirname string) (infos []os.FileInfo, err error) {
	dirs, err := os.ReadDir(dirname)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, d := range dirs {
		i, err := d.Info()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		infos = append(infos, i)
	}

	return infos, nil
}

func (s *Service) Generate(ctx context.Context, in GenerateInput) (configmap *corev1.ConfigMap, secret *corev1.Secret, err error) {
	const (
		owner = "giantswarm"
		repo  = "config"
	)

	var store github.Store
	store = &fsStore{}

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

	meta := metav1.ObjectMeta{
		Name:      in.Name,
		Namespace: in.Namespace,

		Annotations: in.ExtraAnnotations,
		Labels:      in.ExtraLabels,
	}

	configMap, secret, err := gen.GenerateConfig(ctx, in.App, meta)
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	return configMap, secret, nil
}
