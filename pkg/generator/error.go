package generator

import (
	"github.com/giantswarm/config-controller/pkg/github"
	"github.com/giantswarm/microerror"
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

var emptyValueError = &microerror.Error{
	Kind: "emptyValueError",
}

// IsEmptyValue asserts emptyValueError.
func IsEmptyValue(err error) bool {
	return microerror.Cause(err) == emptyValueError
}
