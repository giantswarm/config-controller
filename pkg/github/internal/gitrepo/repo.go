package gitrepo

import (
	"context"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/giantswarm/config-controller/internal/shared"

	intssh "github.com/giantswarm/config-controller/internal/ssh"
)

type Config struct {
	SharedConfigRepository shared.ConfigRepository
	GitHubSSHCredential    intssh.Credential
	GitHubToken            string
}

type Repo struct {
	sharedConfigRepository shared.ConfigRepository
	gitHubSSHCredential    intssh.Credential
	gitHubToken            string
}

func New(config Config) (*Repo, error) {
	r := &Repo{
		sharedConfigRepository: config.SharedConfigRepository,
		gitHubSSHCredential:    config.GitHubSSHCredential,
		gitHubToken:            config.GitHubToken,
	}

	return r, nil
}

func (r *Repo) AssembleConfigRepository(ctx context.Context, owner, name, branch string) (*Store, error) {
	return r.ShallowAssembleConfigRepository(ctx, owner, name, branch)
}

func (r *Repo) ShallowAssembleConfigRepository(ctx context.Context, owner, name, branch string) (*Store, error) {
	// Initialize file system
	fs := memfs.New()

	// Clone root config repository
	url, auth, err := r.createUrlAndAuthMethod(owner, name, r.gitHubToken, r.gitHubSSHCredential.Key, r.gitHubSSHCredential.Password)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	_, err = git.CloneContext(ctx, memory.NewStorage(), fs, &git.CloneOptions{
		Auth:          auth,
		URL:           url,
		ReferenceName: plumbing.ReferenceName(branch),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Assemble shared configs for split config setups
	if r.isSplitSetup(owner, name) {
		err = r.assembleWithSharedConfigs(ctx, fs, owner)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// Initialize and return store
	store := &Store{
		fs: fs,
	}

	return store, err
}

// If the referenced config repository is giantswarm/config then we use the original
// repository setup logic. In all other cases the repository is assumed to be a split
// config setup, and we assemble it from the customer and the shared configs.
//
// This intentionally does not consider if the SSH deploy keys are set up or not so
// that we can use both methods and can be possibly cleaned up when everything uses
// the split setup to enforce the usage of deploy keys and deprecate the usage of a token.
func (r *Repo) isSplitSetup(owner, configRepositoryName string) bool {
	return !(owner == "giantswarm" && configRepositoryName == "config")
}

func (r *Repo) createUrlAndAuthMethod(owner, repositoryName, token, key, password string) (string, transport.AuthMethod, error) {
	repository := owner + "/" + repositoryName + ".git"

	var (
		auth transport.AuthMethod
		err  error
		url  string
	)

	if !(key == "" && password == "") {
		auth, err = ssh.NewPublicKeys(
			"git",
			[]byte(key),
			password,
		)
		if err != nil {
			return "", nil, microerror.Mask(err)
		}

		url = "ssh://git@ssh.github.com:443/" + repository
	} else {
		auth = &http.BasicAuth{
			Username: "can-be-anything-but-not-empty",
			Password: token,
		}

		url = "https://github.com/" + repository
	}

	return url, auth, nil
}

func (r *Repo) assembleWithSharedConfigs(ctx context.Context, fs billy.Filesystem, owner string) error {
	var (
		auth transport.AuthMethod
		err  error
		url  string
	)

	// Clone shared config repository
	sharedRepositoryPath := r.sharedConfigRepository.Name

	err = fs.MkdirAll(sharedRepositoryPath, os.FileMode(int(0777)))
	if err != nil {
		return microerror.Mask(err)
	}

	url, auth, err = r.createUrlAndAuthMethod(owner, r.sharedConfigRepository.Name, r.gitHubToken, r.sharedConfigRepository.Key, r.sharedConfigRepository.Password)
	if err != nil {
		return microerror.Mask(err)
	}

	sharedRepositoryCloneFs, err := fs.Chroot(sharedRepositoryPath)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = git.CloneContext(ctx, memory.NewStorage(), sharedRepositoryCloneFs, &git.CloneOptions{
		Auth:          auth,
		URL:           url,
		ReferenceName: plumbing.ReferenceName(r.sharedConfigRepository.Ref),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	// Move shared defaults to root
	err = fs.Rename("./"+sharedRepositoryPath+"/default", "./default")
	if err != nil {
		return microerror.Mask(err)
	}

	// Move shared includes to root
	err = fs.Rename("./"+sharedRepositoryPath+"/include", "./include")
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.removeAll(fs, sharedRepositoryPath)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Repo) removeAll(fs billy.Filesystem, path string) error {
	fileInfo, err := fs.Lstat(path)
	if err != nil {
		return microerror.Mask(err)
	}

	if fileInfo.IsDir() {
		fileInfos, err := fs.ReadDir(path)
		if err != nil {
			return microerror.Mask(err)
		}
		for _, fileInfo := range fileInfos {
			err = r.removeAll(fs, fs.Join(path, fileInfo.Name()))
			if err != nil {
				return microerror.Mask(err)
			}
		}
		err = fs.Remove(path)
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		err = fs.Remove(path)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
