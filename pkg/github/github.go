package github

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/internal/ssh"
	"github.com/giantswarm/config-controller/pkg/github/internal/gitrepo"
)

type Config struct {
	SSHCredential ssh.Credential
	Token         string
}

type GitHub struct {
	repo *gitrepo.Repo
}

func New(config Config) (*GitHub, error) {
	if config.Token == "" || config.SSHCredential.IsEmpty() {
		return nil, microerror.Maskf(invalidConfigError, "%T.Token or %T.SSHCredential must not be empty", config, config)
	}

	var err error
	var repo *gitrepo.Repo
	{
		c := gitrepo.Config{
			GitHubSSHCredential: config.SSHCredential,
			GitHubToken:         config.Token,
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

func (g *GitHub) GetFilesByBranch(ctx context.Context, owner, name, branch string) (Store, error) {
	store, err := g.repo.ShallowCloneBranch(ctx, owner+"/"+name+".git", branch)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return store, nil

	/*if !g.sshCredential.IsEmpty() {
		url := "ssh://git@ssh.github.com:443/" + owner + "/" + name + ".git"
		store, err = g.repo.ShallowCloneBranchHTTPs(ctx, url, branch)
	} else {
		url := "https://github.com/" + owner + "/" + name + ".git"
		store, err = g.repo.ShallowCloneBranch(ctx, url, branch)
	}

	if err != nil {
		return nil, microerror.Mask(err)
	}

	return store, nil*/
}
