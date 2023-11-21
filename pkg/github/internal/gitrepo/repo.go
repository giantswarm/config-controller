package gitrepo

import (
	"context"
	"github.com/giantswarm/config-controller/internal/shared"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"

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
	ccrRepository := owner + "/" + name + ".git"

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

		url = "ssh://git@ssh.github.com:443/" + ccrRepository
	} else {
		auth = &http.BasicAuth{
			Username: "can-be-anything-but-not-empty",
			Password: r.gitHubToken,
		}

		url = "https://github.com/" + ccrRepository
	}

	// Clone root config repository
	fs := memfs.New()
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

	// Clone shared config repository
	sharedRepository := owner + "/" + r.sharedConfigRepository.Name + ".git"

	sharedRepositoryPath := r.sharedConfigRepository.Name

	err = fs.MkdirAll(sharedRepositoryPath, os.FileMode(int(0777)))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if !r.sharedConfigRepository.IsEmpty() {
		auth, err = ssh.NewPublicKeys(
			"git",
			[]byte(r.sharedConfigRepository.Key),
			r.sharedConfigRepository.Password,
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		url = "ssh://git@ssh.github.com:443/" + sharedRepository
	} else {
		auth = &http.BasicAuth{
			Username: "can-be-anything-but-not-empty",
			Password: r.gitHubToken,
		}

		url = "https://github.com/" + sharedRepository
	}

	sharedRepositoryCloneFs, err := fs.Chroot(sharedRepositoryPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	_, err = git.CloneContext(ctx, memory.NewStorage(), sharedRepositoryCloneFs, &git.CloneOptions{
		Auth:          auth,
		URL:           url,
		ReferenceName: plumbing.ReferenceName(r.sharedConfigRepository.Ref),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Move shared defaults to root
	err = fs.Rename("./"+sharedRepositoryPath+"/configs/default", "./default")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Move shared includes to root
	err = fs.Rename("./"+sharedRepositoryPath+"/configs/include", "./include")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = r.RemoveAll(fs, sharedRepositoryPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Initialize and return store
	store := &Store{
		fs: fs,
	}

	return store, err
}

func (r *Repo) RemoveAll(fs billy.Filesystem, path string) error {
	fileInfo, err := fs.Lstat(path)
	if fileInfo.IsDir() {
		fileInfos, err := fs.ReadDir(path)
		if err != nil {
			return microerror.Mask(err)
		}
		for _, fileInfo := range fileInfos {
			err = r.RemoveAll(fs, fs.Join(path, fileInfo.Name()))
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
