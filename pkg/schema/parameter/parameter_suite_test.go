package parameter_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestParameter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parameter Suite")
}
