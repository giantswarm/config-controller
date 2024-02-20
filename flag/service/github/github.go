package github

import "github.com/giantswarm/config-controller/flag/service/github/ssh"

type GitHub struct {
	RepositoryName         string
	RepositoryRef          string
	SSH                    ssh.SSH
	Token                  string
	SharedConfigRepository SharedConfigRepository
}

type SharedConfigRepository struct {
	Name     string
	Ref      string
	Key      string
	Password string
}
