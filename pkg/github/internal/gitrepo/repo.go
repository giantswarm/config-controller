package gitrepo

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"

	intssh "github.com/giantswarm/config-controller/internal/ssh"
)

type Config struct {
	GitHubSSHCredential intssh.Credential
	GitHubToken         string
}

type Repo struct {
	gitHubSSHCredential intssh.Credential
	gitHubToken         string
}

func New(config Config) (*Repo, error) {
	r := &Repo{
		gitHubSSHCredential: config.GitHubSSHCredential,
		gitHubToken:         config.GitHubToken,
	}

	return r, nil
}

func (r *Repo) ShallowCloneBranch(ctx context.Context, repository, branch string) (*Store, error) {
	return r.ShallowClone(ctx, repository, plumbing.NewBranchReferenceName(branch))
}

func (r *Repo) ShallowClone(ctx context.Context, repository string, ref plumbing.ReferenceName) (*Store, error) {
	var (
		auth transport.AuthMethod
		err  error
		url  string
	)

	if !r.gitHubSSHCredential.IsEmpty() {
		auth, err = ssh.NewPublicKeys(
			"git",
			[]byte(r.gitHubSSHCredential.Key),
			r.gitHubSSHCredential.Password,
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		url = "ssh://git@ssh.github.com:443/" + repository
	} else {
		auth = &http.BasicAuth{
			Username: "can-be-anything-but-not-empty",
			Password: r.gitHubToken,
		}

		url = "https://github.com/" + repository
	}

	fs := memfs.New()
	gitRepo, err := git.CloneContext(ctx, memory.NewStorage(), fs, &git.CloneOptions{
		Auth:          auth,
		URL:           url,
		ReferenceName: ref,
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	wt, err := gitRepo.Worktree()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	submodules, err := wt.Submodules()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = submodules.Update(&git.SubmoduleUpdateOptions{
		Init:              true,
		Depth:             0,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	store := &Store{
		fs: fs,
	}

	return store, err
}
