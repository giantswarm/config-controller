package lint

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	flagFilterFunctions  = "filter-functions"
	flagGithubToken      = "github-token"
	flagMaxMessages      = "max-messages"
	flagNoDescriptions   = "no-descriptions"
	flagNoFuncNames      = "no-function-names"
	flagOnlyErrors       = "only-errors"
	flagRepositoryName   = "repository-name"
	flagRepositoryRef    = "repository-ref"
	flagSkipFieldsRegexp = "skip-fields-regexp"

	envConfigControllerGithubToken = "CONFIG_CONTROLLER_GITHUB_TOKEN" //nolint:gosec
)

type flag struct {
	FilterFunctions  []string
	GitHubToken      string
	MaxMessages      int
	NoDescriptions   bool
	NoFuncNames      bool
	OnlyErrors       bool
	RepositoryName   string
	RepositoryRef    string
	SkipFieldsRegexp string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&f.FilterFunctions, flagFilterFunctions, []string{}, `Enables filtering linter functions by supplying a list of patterns to match, (e.g. "Lint.*,LintUnusedConfigValues").`)
	cmd.Flags().StringVar(&f.GitHubToken, flagGithubToken, "", fmt.Sprintf(`GitHub token to use for "opsctl create vaultconfig" calls. Defaults to the value of %s env var.`, envConfigControllerGithubToken))
	cmd.Flags().IntVar(&f.MaxMessages, flagMaxMessages, 50, "Max number of linter messages to display. Unlimited output if set to 0. Defaults to 50.")
	cmd.Flags().BoolVar(&f.NoDescriptions, flagNoDescriptions, false, "Disables output of message descriptions.")
	cmd.Flags().BoolVar(&f.NoFuncNames, flagNoFuncNames, false, "Disables output of linter function names.")
	cmd.Flags().BoolVar(&f.OnlyErrors, flagOnlyErrors, false, "Enables linter to output only errors, omitting suggestions.")
	cmd.Flags().StringVar(&f.RepositoryName, flagRepositoryName, "config", `Repository name where configs are stored under the giantswarm organization, defaults to "config".`)
	cmd.Flags().StringVar(&f.RepositoryRef, flagRepositoryRef, "main", `Repository branch to use, defaults to "main"`)
	cmd.Flags().StringVar(&f.SkipFieldsRegexp, flagSkipFieldsRegexp, "", "List of regexp matchers to match field paths, which don't require validation.")
}

func (f *flag) Validate() error {
	if f.GitHubToken == "" {
		f.GitHubToken = os.Getenv(envConfigControllerGithubToken)
	}
	if f.GitHubToken == "" {
		return microerror.Maskf(invalidFlagError, "--%s or $%s must not be empty", flagGithubToken, envConfigControllerGithubToken)
	}

	res := strings.Split(f.SkipFieldsRegexp, ",")
	if res[0] == "" {
		return nil
	}

	for _, re := range res {
		_, err := regexp.Compile(re)
		if err != nil {
			return microerror.Maskf(invalidFlagError, "%#q must be a valid regex string", re)
		}
	}

	return nil
}
