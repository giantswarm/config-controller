package generate

import (
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	flagApp          = "app"
	flagDir          = "dir"
	flagInstallation = "installation"
	flagVersion      = "version"
)

type flag struct {
	App          string
	Dir          string
	Installation string
	Version      string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.App, flagApp, "", `Name of an application to generate the config for (e.g. "kvm-operator").`)
	cmd.Flags().StringVar(&f.Dir, flagDir, "", "Directory containing configuration. If empty, contents of giantswarm/config will be used.")
	cmd.Flags().StringVar(&f.Installation, flagInstallation, "", `Installation codename (e.g. "gauss").`)
	cmd.Flags().StringVar(&f.Version, flagVersion, "", `Version of config to use for generation (e.g. "v2.3.19").`)
}

func (f *flag) Validate() error {
	if f.App == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagApp)
	}
	if f.Installation == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagInstallation)
	}
	if f.Version == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", flagVersion)
	}

	return nil
}
