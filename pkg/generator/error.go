package generator

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/pkg/github"
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
	if github.IsNotFound(err) {
		return true
	}

	return microerror.Cause(err) == notFoundError
}

var emptySecretValueError = &microerror.Error{
	Kind: "emptySecretValueError",
}

// IsEmptySecretValue asserts emptySecretValueError.
func IsEmptySecretValue(err error) bool {
	return microerror.Cause(err) == emptySecretValueError
}
