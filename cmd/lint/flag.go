package lint

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
	flagMaxErrors     = "max-errors"

	envConfigControllerGithubToken = "CONFIG_CONTROLLER_GITHUB_TOKEN" //nolint:gosec
)

type flag struct {
	App           string
	Branch        string
	ConfigVersion string
	GitHubToken   string
	Installation  string
	MaxErrors     int
}

func (f *flag) Init(cmd *cobra.Command) {
	// TODO: flags are optional if you wish to narrow down linter's scope
	cmd.Flags().StringVar(&f.App, flagApp, "", `Name of an application to generate the config for (e.g. "kvm-operator").`)
	cmd.Flags().StringVar(&f.Branch, flagBranch, "", "Branch of giantswarm/config used to generate configuraton.")
	cmd.Flags().StringVar(&f.ConfigVersion, flagConfigVersion, "", `Major part of the configuration version to use for generation (e.g. "v2").`)
	cmd.Flags().StringVar(&f.Installation, flagInstallation, "", `Installation codename (e.g. "gauss").`)
	cmd.Flags().StringVar(&f.GitHubToken, flagGithubToken, "", fmt.Sprintf(`GitHub token to use for "opsctl create vaultconfig" calls. Defaults to the value of %s env var.`, envConfigControllerGithubToken))
	cmd.Flags().IntVar(&f.MaxErrors, flagMaxErrors, 50, "Max number of linter errors to display. Unlimited output if set to 0. Defaults to 50.")
}

func (f *flag) Validate() error {
	if f.ConfigVersion == "" && f.Branch == "" {
		f.Branch = "main"
	}
	if f.GitHubToken == "" {
		f.GitHubToken = os.Getenv(envConfigControllerGithubToken)
	}
	if f.GitHubToken == "" {
		return microerror.Maskf(invalidFlagError, "--%s or $%s must not be empty", flagGithubToken, envConfigControllerGithubToken)
	}

	return nil
}