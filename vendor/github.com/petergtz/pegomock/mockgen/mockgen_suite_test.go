package mockgen_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMockgen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mockgen Suite")
}
