package gomock_test

import (
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/petergtz/pegomock/modelgen/gomock"
)

func TestGomock(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "gomock Suite")
}

var _ = Describe("reflect", func() {
	It("can generate mocks for interfaces taken from vendored packages", func() {
		_, e := gomock.Reflect("github.com/petergtz/vendored_package", []string{"Interface"})
		Expect(e).NotTo(HaveOccurred())
	})
})
