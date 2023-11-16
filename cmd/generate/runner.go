package generate

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ghodss/yaml"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"github.com/giantswarm/config-controller/internal/generator"
	"github.com/giantswarm/config-controller/internal/meta"
	"github.com/giantswarm/config-controller/internal/ssh"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	stdout io.Writer
	stderr io.Writer
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	var vaultClient *vaultapi.Client
	{
		vaultClient, err = createVaultClientUsingOpsctl(ctx, r.flag.GitHubToken, r.flag.SSHUser, r.flag.Installation)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	configRepoSshKey := ""
	if r.flag.ConfigRepoSSHPemPath != "" {
		configRepoSshKey, err = r.readSSHPem(r.flag.ConfigRepoSSHPemPath)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	defaultConfigRepoSshKey := ""
	if r.flag.ConfigRepoSSHPemPath != "" {
		configRepoSshKey, err = r.readSSHPem(r.flag.DefaultConfigRepoSSHPemPath)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	includeConfigRepoSshKey := ""
	if r.flag.ConfigRepoSSHPemPath != "" {
		configRepoSshKey, err = r.readSSHPem(r.flag.IncludeConfigRepoSSHPemPath)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var gen *generator.Service
	{
		c := generator.Config{
			VaultClient: vaultClient,

			DefaultConfigRepoSSHCredential: ssh.Credential{
				Key:      defaultConfigRepoSshKey,
				Password: r.flag.DefaultConfigRepoSSHPemPassword,
			},
			IncludeConfigRepoSSHCredential: ssh.Credential{
				Key:      includeConfigRepoSshKey,
				Password: r.flag.IncludeConfigRepoSSHPemPassword,
			},
			ConfigRepoSSHCredential: ssh.Credential{
				Key:      configRepoSshKey,
				Password: r.flag.ConfigRepoSSHPemPassword,
			},
			GitHubToken:    r.flag.GitHubToken,
			RepositoryName: r.flag.RepositoryName,
			RepositoryRef:  r.flag.RepositoryRef,
			Installation:   r.flag.Installation,
			Verbose:        r.flag.Verbose,
		}

		gen, err = generator.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	in := generator.GenerateInput{
		App: r.flag.App,

		Name:      r.flag.Name,
		Namespace: r.flag.Namespace,

		ExtraAnnotations: map[string]string{
			meta.Annotation.XAppInfo.Key():        meta.Annotation.XAppInfo.Val("<unknown>", r.flag.App, "<unknown>"),
			meta.Annotation.XCreator.Key():        meta.Annotation.XCreator.Default(),
			meta.Annotation.XInstallation.Key():   r.flag.Installation,
			meta.Annotation.XProjectVersion.Key(): meta.Annotation.XProjectVersion.Val(false),
		},
		ExtraLabels: nil,
	}

	configmap, secret, err := gen.Generate(ctx, in)
	if err != nil {
		return microerror.Mask(err)
	}

	if r.flag.Raw {
		fmt.Println("---")
		fmt.Printf(string(configmap.Data["configmap-values.yaml"]) + "\n")
		fmt.Println("---")
		fmt.Printf(string(secret.Data["secret-values.yaml"]) + "\n")
		return nil
	}

	fmt.Println("---")
	out, err := yaml.Marshal(configmap)
	if err != nil {
		return microerror.Mask(err)
	}
	fmt.Printf(string(out) + "\n")

	fmt.Println("---")
	out, err = yaml.Marshal(secret)
	if err != nil {
		return microerror.Mask(err)
	}
	fmt.Printf(string(out) + "\n")

	return nil
}

func (r *runner) readSSHPem(path string) (string, error) {
	keyByte, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(keyByte), nil
}
