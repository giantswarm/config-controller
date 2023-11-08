package generate

import (
	"fmt"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	flagApp           = "app"
	flagConfigVersion = "config-version"
	flagGithubToken   = "github-token"
	repositoryName    = "repository-name"
	flagInstallation  = "installation"
	flagName          = "name"
	flagNamespace     = "namespace"
	flagRaw           = "raw"
	flagSSHUser       = "ssh-user"
	flagVerbose       = "verbose"

	envConfigControllerGithubToken = "CONFIG_CONTROLLER_GITHUB_TOKEN" //nolint:gosec
)

type flag struct {
	App            string
	ConfigVersion  string
	GitHubToken    string
	RepositoryName string
	Installation   string
	Name           string
	Namespace      string
	Raw            bool
	SSHUser        string
	Verbose        bool
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.App, flagApp, "", `Name of an application to generate the config for (e.g. "kvm-operator").`)
	cmd.Flags().StringVar(&f.ConfigVersion, flagConfigVersion, "", `Configuration version. Can be a major version range in format "2.x.x" or a branch name.`)
	cmd.Flags().StringVar(&f.GitHubToken, flagGithubToken, "", fmt.Sprintf(`GitHub token to use for "opsctl create vaultconfig" calls. Defaults to the value of %s env var.`, envConfigControllerGithubToken))
	cmd.Flags().StringVar(&f.RepositoryName, repositoryName, "config", `Repository name where configs are stored under the giantswarm organization, defaults to "config".`)
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
	if f.ConfigVersion == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagConfigVersion)
	}
	if f.GitHubToken == "" {
		f.GitHubToken = os.Getenv(envConfigControllerGithubToken)
	}
	if f.GitHubToken == "" {
		return microerror.Maskf(invalidFlagError, "--%s or $%s must not be empty", flagGithubToken, envConfigControllerGithubToken)
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
