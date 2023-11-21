package github

import (
	"context"
	"github.com/giantswarm/config-controller/internal/shared"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/internal/ssh"
	"github.com/giantswarm/config-controller/pkg/github/internal/gitrepo"
)

type Config struct {
	SharedConfigRepository shared.ConfigRepository
	SSHCredential          ssh.Credential
	Token                  string
}

type GitHub struct {
	repo *gitrepo.Repo
}

func New(config Config) (*GitHub, error) {
	if config.Token == "" && config.SSHCredential.IsEmpty() {
		return nil, microerror.Maskf(invalidConfigError, "%T.Token or %T.ConfigRepoSSHCredential must not be empty", config, config)
	}

	var err error
	var repo *gitrepo.Repo
	{
		c := gitrepo.Config{
			SharedConfigRepository: config.SharedConfigRepository,
			GitHubSSHCredential:    config.SSHCredential,
			GitHubToken:            config.Token,
		}

		repo, err = gitrepo.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	g := &GitHub{
		repo: repo,
	}

	return g, nil
}

func (g *GitHub) AssembleConfigRepository(ctx context.Context, owner, name, branch string) (Store, error) {
	store, err := g.repo.AssembleConfigRepository(ctx, owner, name, branch)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return store, nil
}
