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
	SharedConfigRepository shared.ConfigRepository
	client                 *github.GitHub
	repoCache              *cache.Repository
	tagCache               *cache.Tag
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
		SharedConfigRepository: c.SharedConfigRepository,
		client:                 client,
		repoCache:              cache.NewRepository(),
		tagCache:               cache.NewTag(),
	}
	return gh, nil
}

func (gh *GitHub) AssembleConfigRepository(ctx context.Context, owner, name, branch string) (github.Store, error) {
	key := gh.repoCache.Key(owner, name, branch, gh.SharedConfigRepository.Name, gh.SharedConfigRepository.Ref)
	store, cached := gh.repoCache.Get(ctx, key)
	if cached {
		return store, nil
	}

	store, err := gh.client.AssembleConfigRepository(ctx, owner, name, branch)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	gh.repoCache.Set(ctx, key, store)
	return store, nil
}
