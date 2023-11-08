package github

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/pkg/github/internal/gitrepo"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	if gitrepo.IsNotFound(err) {
		return true
	}

	return microerror.Cause(err) == notFoundError
}
