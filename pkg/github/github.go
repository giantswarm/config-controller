package github

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/pkg/github/internal/gitrepo"
	"github.com/giantswarm/config-controller/pkg/github/internal/graphql"
)

type Config struct {
	Token string
}

type GitHub struct {
	graphQLClient *graphql.Client
	repo          *gitrepo.Repo
}

func New(config Config) (*GitHub, error) {
	if config.Token == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Token must not be empty", config)
	}

	var err error

	var graphQLClient *graphql.Client
	{
		c := graphql.Config{
			Headers: map[string]string{
				"Authorization": "bearer " + config.Token,
			},
			URL: "https://api.github.com/graphql",
		}
		graphQLClient, err = graphql.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var repo *gitrepo.Repo
	{
		c := gitrepo.Config{
			GitHubToken: config.Token,
		}

		repo, err = gitrepo.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	g := &GitHub{
		graphQLClient: graphQLClient,
		repo:          repo,
	}

	return g, nil
}

func (g *GitHub) GetFilesByBranch(ctx context.Context, owner, name, branch string) (Store, error) {
	url := "https://github.com/" + owner + "/" + name + ".git"
	store, err := g.repo.ShallowCloneBranch(ctx, url, branch)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return store, nil
}
