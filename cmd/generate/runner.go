package generate

import (
	"context"
	"fmt"
	"io"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/config-controller/pkg/decrypter"
	"github.com/giantswarm/config-controller/pkg/generator"
	"github.com/giantswarm/config-controller/pkg/github"
)

const (
	owner = "giantswarm"
	repo  = "config"
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
	gh, err := github.New(github.Config{
		Token: r.flag.GitHubToken,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	var store generator.Filesystem
	if r.flag.ConfigVersion != "" {
		tag, err := gh.GetLatestTag(ctx, owner, repo, r.flag.ConfigVersion)
		if err != nil {
			return microerror.Mask(err)
		}

		store, err = gh.GetFilesByTag(ctx, owner, repo, tag)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if r.flag.Branch != "" {
		store, err = gh.GetFilesByBranch(ctx, owner, repo, r.flag.Branch)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	gen, err := generator.New(&generator.Config{
		Fs: store,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	configmap, secrets, err := gen.GenerateConfig(r.flag.Installation, r.flag.App)
	if err != nil {
		return microerror.Mask(err)
	}

	// TODO KUBA: clean up below

	fmt.Fprintf(r.stdout, "Creating vault client using opsctl\n")

	vaultClient, err := createVaultClientUsingOpsctl(ctx, r.flag.GitHubToken, r.flag.Installation)
	if err != nil {
		return microerror.Mask(err)
	}

	var dec *decrypter.Decrypter
	{
		c := decrypter.Config{
			VaultClient: vaultClient,
		}

		dec, err = decrypter.New(c)
		if err != nil {
			return microerror.Mask(err)
		}

	}

	if len(args) != 1 {
		fmt.Fprintf(r.stderr, "Error: Expected the first argument to encrypted blob")
	}

	decrypted, err := dec.Decrypt(ctx, []byte(args[0]))
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Fprintf(r.stdout, "Decrypted: %s\n", decrypted)

	return nil
}
