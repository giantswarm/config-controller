package generate

import (
	"fmt"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	flagApp                             = "app"
	flagDefaultConfigRepoSSHPemPath     = "default-config-repo-ssh-pem-path"
	flagDefaultConfigRepoSSHPemPassword = "default-config-repo-ssh-pem-password" // #nosec G101
	flagIncludeConfigRepoSSHPemPath     = "include-config-repo-ssh-pem-path"
	flagIncludeConfigRepoSSHPemPassword = "include-config-repo-ssh-pem-password" // #nosec G101
	flagConfigRepoSSHPemPath            = "config-repo-ssh-pem-path"
	flagConfigRepoSSHPemPassword        = "config-repo-ssh-pem-password" // #nosec G101
	flagGithubToken                     = "github-token"
	flagInstallation                    = "installation"
	flagName                            = "name"
	flagNamespace                       = "namespace"
	flagRaw                             = "raw"
	flagRepositoryName                  = "repository-name"
	flagRepositoryRef                   = "repository-ref"
	flagSSHUser                         = "ssh-user"
	flagVerbose                         = "verbose"

	envConfigControllerGithubToken = "CONFIG_CONTROLLER_GITHUB_TOKEN" //nolint:gosec
)

type flag struct {
	App                             string
	DefaultConfigRepoSSHPemPath     string
	DefaultConfigRepoSSHPemPassword string
	IncludeConfigRepoSSHPemPath     string
	IncludeConfigRepoSSHPemPassword string
	ConfigRepoSSHPemPath            string
	ConfigRepoSSHPemPassword        string
	GitHubToken                     string
	RepositoryName                  string
	RepositoryRef                   string
	Installation                    string
	Name                            string
	Namespace                       string
	Raw                             bool
	SSHUser                         string
	Verbose                         bool
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.App, flagApp, "", `Name of an application to generate the config for (e.g. "kvm-operator").`)
	cmd.Flags().StringVar(&f.DefaultConfigRepoSSHPemPath, flagDefaultConfigRepoSSHPemPath, "", `Path to the SSH private key file to use for downloading the default configuration submodule.`)
	cmd.Flags().StringVar(&f.DefaultConfigRepoSSHPemPassword, flagDefaultConfigRepoSSHPemPassword, "", `Passphrase to the default config submodule SSH private key.`)
	cmd.Flags().StringVar(&f.IncludeConfigRepoSSHPemPath, flagIncludeConfigRepoSSHPemPath, "", `Path to the SSH private key file to use for downloading the include configuration submodule.`)
	cmd.Flags().StringVar(&f.IncludeConfigRepoSSHPemPassword, flagIncludeConfigRepoSSHPemPassword, "", `Passphrase to the include config submodule SSH private key.`)
	cmd.Flags().StringVar(&f.ConfigRepoSSHPemPath, flagConfigRepoSSHPemPath, "", `Path to the SSH private key file to use for downloading the configuration repository.`)
	cmd.Flags().StringVar(&f.ConfigRepoSSHPemPassword, flagConfigRepoSSHPemPassword, "", `Passphrase to the config repo SSH private key.`)
	cmd.Flags().StringVar(&f.GitHubToken, flagGithubToken, "", fmt.Sprintf(`GitHub token to use for "opsctl create vaultconfig" calls. Defaults to the value of %s env var.`, envConfigControllerGithubToken))
	cmd.Flags().StringVar(&f.RepositoryName, flagRepositoryName, "config", `Repository name where configs are stored under the giantswarm organization, defaults to "config".`)
	cmd.Flags().StringVar(&f.RepositoryRef, flagRepositoryRef, "main", `Repository branch to use, defaults to "main"`)
	cmd.Flags().StringVar(&f.Installation, flagInstallation, "", `Installation codename (e.g. "gauss").`)
	cmd.Flags().StringVar(&f.Name, flagName, "giantswarm", `Name of the generated ConfigMap/Secret.`)
	cmd.Flags().StringVar(&f.Namespace, flagNamespace, "giantswarm", `Namespace of the generated ConfigMap/Secret.`)
	cmd.Flags().BoolVar(&f.Raw, flagRaw, false, `Forces generator to output YAML instead of ConfigMap & Secret.`)
	cmd.Flags().StringVar(&f.SSHUser, flagSSHUser, "", `User to be passed to opsctl.`)
	cmd.Flags().BoolVar(&f.Verbose, flagVerbose, false, `Enables generator to output consecutive generation stages.`)
}

func (f *flag) Validate() error {
	if f.App == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagApp)
	}
	if f.GitHubToken == "" {
		f.GitHubToken = os.Getenv(envConfigControllerGithubToken)
	}
	if f.GitHubToken == "" && f.ConfigRepoSSHPemPath == "" {
		return microerror.Maskf(
			invalidFlagError,
			"--%s or $%s must not be empty when SSH credentials are not provided for the config repository either.",
			flagGithubToken, envConfigControllerGithubToken)
	}
	if f.Installation == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagInstallation)
	}
	if f.Name == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagName)
	}
	if f.Namespace == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagNamespace)
	}

	return nil
}
