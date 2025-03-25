// Package testutil holds code that may be useful for any of the e2e subpackages (including e2e itself).
// It is useful to avoid import loops between the various e2e test pacakges.
package util

import (
	"github.com/runfinch/common-tests/option"
)

// NewOpt is a helper to make it easier for functions to accept wrapped option creators.
type NewOpt func(subject []string, modifiers ...option.Modifier) (*option.Option, error)

// WrappedOption allows injection of new prefixed option creator function into tests.
// This is useful for scenarios where CLI commands must be run in an environment which is
// not the same as the system running the tests, like inside a SSH shell.
func WrappedOption(prefix []string, wModifiers ...option.Modifier) NewOpt {
	return func(subject []string, modifiers ...option.Modifier) (*option.Option, error) {
		prefix = append(prefix, subject...)
		return option.New(prefix, append(wModifiers, modifiers...)...)
	}
}
