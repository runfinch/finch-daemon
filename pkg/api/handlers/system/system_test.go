package system

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestSystem is the entry point for unit tests in the system package
func TestSystem(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests - System APIs Handler")
}
