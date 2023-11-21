package github

import (
	"context"
	"github.com/giantswarm/config-controller/internal/shared"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/internal/generator/github/cache"
	"github.com/giantswarm/config-controller/internal/ssh"
	"github.com/giantswarm/config-controller/pkg/github"
)

type Config struct {
	SharedConfigRepository  shared.ConfigRepository
	ConfigRepoSSHCredential ssh.Credential
	Token                   string
}

type GitHub struct {
	client    *github.GitHub
	repoCache *cache.Repository
	tagCache  *cache.Tag
}

func New(c Config) (*GitHub, error) {
	client, err := github.New(github.Config{
		SharedConfigRepository: c.SharedConfigRepository,
		SSHCredential:          c.ConfigRepoSSHCredential,
		Token:                  c.Token,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	gh := &GitHub{
		client:    client,
		repoCache: cache.NewRepository(),
		tagCache:  cache.NewTag(),
	}
	return gh, nil
}

func (gh *GitHub) GetFilesByBranch(ctx context.Context, owner, name, branch string) (github.Store, error) {
	key := gh.repoCache.Key(owner, name, branch)
	store, cached := gh.repoCache.Get(ctx, key)
	if cached {
		return store, nil
	}

	store, err := gh.client.GetFilesByBranch(ctx, owner, name, branch)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	gh.repoCache.Set(ctx, key, store)
	return store, nil
}
