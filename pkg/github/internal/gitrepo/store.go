package gitrepo

import (
	"context"
	"io/ioutil"

	"github.com/giantswarm/microerror"
	"github.com/go-git/go-billy/v5"
)

type Store struct {
	fs billy.Filesystem
}

func (s *Store) GetContent(ctx context.Context, path string) ([]byte, error) {
	stat, err := s.fs.Stat(path)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if stat.IsDir() {

	}

	f, err := s.fs.Open(path)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return bs, nil
}
