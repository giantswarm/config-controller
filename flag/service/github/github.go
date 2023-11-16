package github

import "github.com/giantswarm/config-controller/flag/service/github/ssh"

type GitHub struct {
	RepositoryName string
	RepositoryRef  string
	SSH            ssh.SSH
	Token          string
}
