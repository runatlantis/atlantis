// Copyright 2016 Peter Goetz
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package watch_test

import (
	"go/build"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/petergtz/pegomock/pegomock/testutil"
	"github.com/petergtz/pegomock/pegomock/watch"
)

var (
	joinPath = filepath.Join
)

func TestWatchCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Watch Suite")
}

var _ = Describe("NewMockFileUpdater", func() {
	var (
		packageDir, subPackageDir, vendorPackageDir string
		origWorkingDir                              string
	)

	BeforeEach(func() {
		packageDir = joinPath(build.Default.GOPATH, "src", "pegomocktest")
		Expect(os.MkdirAll(packageDir, 0755)).To(Succeed())
		subPackageDir = joinPath(packageDir, "subpackage")
		Expect(os.MkdirAll(subPackageDir, 0755)).To(Succeed())
		vendorPackageDir = joinPath(packageDir, "vendor", "github.com", "petergtz", "vendored_package")
		Expect(os.MkdirAll(vendorPackageDir, 0755)).To(Succeed())

		var e error
		origWorkingDir, e = os.Getwd()
		Expect(e).NotTo(HaveOccurred())
		os.Chdir(packageDir)

		WriteFile(joinPath(packageDir, "mydisplay.go"),
			"package pegomocktest; type MyDisplay interface {  Show() }")
		WriteFile(joinPath(subPackageDir, "subdisplay.go"),
			"package subpackage; type SubDisplay interface {  ShowMe() }")
		WriteFile(joinPath(vendorPackageDir, "iface.go"),
			`package vendored_package; type Interface interface{ Foobar() }`)
		WriteFile(joinPath(packageDir, "vendordisplay.go"), `package pegomocktest
			import ( "github.com/petergtz/vendored_package" )
			type VendorDisplay interface { Show(something vendored_package.Interface) }`)

	})

	AfterEach(func() {
		Expect(os.RemoveAll(packageDir)).To(Succeed())
		os.Chdir(origWorkingDir)
	})

	Context("after populating interfaces_to_mock with an actual interface", func() {
		It(`Eventually creates a file mock_mydisplay_test.go starting with "package pegomocktest_test"`, func() {
			WriteFile(joinPath(packageDir, "interfaces_to_mock"), "MyDisplay")

			watch.NewMockFileUpdater([]string{packageDir}, false).Update()

			Eventually(joinPath(packageDir, "mock_mydisplay_test.go"), "3s").Should(SatisfyAll(
				BeAnExistingFile(),
				BeAFileContainingSubString("package pegomocktest_test")))
		})

		Context("and overriding the output filepath", func() {
			It(`Eventually creates a file foo.go starting with "package pegomocktest_test"`, func() {
				WriteFile(joinPath(packageDir, "interfaces_to_mock"), "-o foo.go MyDisplay")

				watch.NewMockFileUpdater([]string{packageDir}, false).Update()

				Eventually(joinPath(packageDir, "foo.go"), "3s").Should(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

		Context("and overriding the package name", func() {
			It(`Eventually creates a file starting with "package the_overriden_test_package"`, func() {
				WriteFile(joinPath(packageDir, "interfaces_to_mock"), "--package the_overriden_test_package MyDisplay")

				watch.NewMockFileUpdater([]string{packageDir}, false).Update()

				Eventually(joinPath(packageDir, "mock_mydisplay_test.go"), "3s").Should(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package the_overriden_test_package")))
			})
		})

		Context(`and specifying the vendor path`, func() {

			It(`Eventually creates a file containing the import ( vendored_package "github.com/petergtz/vendored_package" )'`, func() {
				WriteFile(joinPath(packageDir, "interfaces_to_mock"), "VendorDisplay")

				watch.NewMockFileUpdater([]string{packageDir}, false).Update()

				Expect(joinPath(packageDir, "mock_vendordisplay_test.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString(`vendored_package "github.com/petergtz/vendored_package"`)))
			})
		})

		Context("in multiple packages and providing those packages to watch", func() {
			It(`Eventually creates correct files in respective directories`, func() {
				os.Chdir("..")
				WriteFile(joinPath(packageDir, "interfaces_to_mock"), "MyDisplay")
				WriteFile(joinPath(subPackageDir, "interfaces_to_mock"), "SubDisplay")

				watch.NewMockFileUpdater([]string{"pegomocktest", "pegomocktest/subpackage"}, false).Update()

				Eventually(joinPath(packageDir, "mock_mydisplay_test.go"), "3s").Should(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
				Eventually(joinPath(subPackageDir, "mock_subdisplay_test.go"), "3s").Should(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package subpackage_test")))
			})
		})

		Context("in one package, but providing multiple packages to create mocks from", func() {
			It(`Eventually creates correct files in respective directories`, func() {
				os.Chdir("..")
				WriteFile(joinPath(packageDir, "interfaces_to_mock"), "MyDisplay\npegomocktest/subpackage SubDisplay")

				watch.NewMockFileUpdater([]string{"pegomocktest", "pegomocktest/subpackage"}, false).Update()

				Eventually(joinPath(packageDir, "mock_mydisplay_test.go"), "3s").Should(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
				Eventually(joinPath(packageDir, "mock_subdisplay_test.go"), "3s").Should(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

		Context("in multiple packages and watching --recursive", func() {
			It(`Eventually creates correct files in respective directories`, func() {
				WriteFile(joinPath(packageDir, "interfaces_to_mock"), "MyDisplay")
				WriteFile(joinPath(subPackageDir, "interfaces_to_mock"), "SubDisplay")

				watch.NewMockFileUpdater([]string{packageDir}, true).Update()

				Eventually(joinPath(packageDir, "mock_mydisplay_test.go"), "3s").Should(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
				Eventually(joinPath(subPackageDir, "mock_subdisplay_test.go"), "3s").Should(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package subpackage_test")))
			})
		})

	})

	Context("after populating interfaces_to_mock with a Go file", func() {
		It(`Eventually creates a file mock_mydisplay_test.go starting with "package pegomocktest_test"`, func() {
			WriteFile(joinPath(packageDir, "interfaces_to_mock"), "mydisplay.go")

			watch.NewMockFileUpdater([]string{packageDir}, false).Update()

			Eventually(joinPath(packageDir, "mock_mydisplay_test.go"), "3s").Should(SatisfyAll(
				BeAnExistingFile(),
				BeAFileContainingSubString("package pegomocktest_test")))
		})
	})

})
