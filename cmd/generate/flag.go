package generate

import (
	"fmt"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	flagApp           = "app"
	flagBranch        = "branch"
	flagConfigVersion = "config-version"
	flagGithubToken   = "github-token"
	flagInstallation  = "installation"
	flagNamespace     = "namespace"

	flagLocalGenerator = "local-generator"

	envConfigControllerGithubToken = "CONFIG_CONTROLLER_GITHUB_TOKEN" //nolint:gosec
)

type flag struct {
	App            string
	Branch         string
	ConfigVersion  string
	GitHubToken    string
	Installation   string
	LocalGenerator bool
	Namespace      string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.App, flagApp, "", `Name of an application to generate the config for (e.g. "kvm-operator").`)
	cmd.Flags().StringVar(&f.Branch, flagBranch, "", "Branch of giantswarm/config used to generate configuraton.")
	cmd.Flags().StringVar(&f.ConfigVersion, flagConfigVersion, "", `Major part of the configuration version to use for generation (e.g. "v2").`)
	cmd.Flags().StringVar(&f.Installation, flagInstallation, "", `Installation codename (e.g. "gauss").`)
	cmd.Flags().StringVar(&f.Namespace, flagNamespace, "giantswarm", `Namespace to generate cm/secret for (defaults to "giantswarm").`)
	cmd.Flags().StringVar(&f.GitHubToken, flagGithubToken, "", fmt.Sprintf(`GitHub token to use for "opsctl create vaultconfig" calls. Defaults to the value of %s env var.`, envConfigControllerGithubToken))
	cmd.Flags().BoolVar(&f.LocalGenerator, flagLocalGenerator, false, `Use local filesystem as source of configuration.`)
}

func (f *flag) Validate() error {
	if f.App == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagApp)
	}
	if f.ConfigVersion != "" && f.Branch != "" {
		return microerror.Maskf(invalidFlagError, "--%s can not be used with --%s", flagConfigVersion, flagBranch)
	}
	if f.ConfigVersion == "" && f.Branch == "" && !f.LocalGenerator {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagConfigVersion)
	}
	if (f.ConfigVersion != "" || f.Branch != "") && f.LocalGenerator {
		return microerror.Maskf(invalidFlagError, "--%s and --%s can not be used with --%s", flagConfigVersion, flagBranch, flagLocalGenerator)
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

	return nil
}
