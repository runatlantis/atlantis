// Copyright 2015 Peter Goetz
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

package modelgen_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/petergtz/pegomock/model"
	"github.com/petergtz/pegomock/modelgen/gomock"
	"github.com/petergtz/pegomock/modelgen/loader"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

func TestDSL(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "modelgen Suite")
}

type alphabetically []*model.Method

func (a alphabetically) Len() int           { return len(a) }
func (a alphabetically) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a alphabetically) Less(i, j int) bool { return a[i].Name < a[j].Name }

var _ = Describe("modelgen/loader", func() {
	It("generates an equivalent model as gomock/reflect does", func() {
		pkgFromReflect, e := gomock.Reflect("github.com/petergtz/pegomock/test_interface", []string{"Display"})
		Expect(e).NotTo(HaveOccurred())
		sort.Sort(alphabetically(pkgFromReflect.Interfaces[0].Methods))

		pkgFromLoader, e := loader.GenerateModel("github.com/petergtz/pegomock/test_interface", "Display")
		Expect(e).NotTo(HaveOccurred())
		sort.Sort(alphabetically(pkgFromLoader.Interfaces[0].Methods))

		Expect(pkgFromLoader.Name).To(Equal(pkgFromReflect.Name))
		Expect(pkgFromLoader.Interfaces).To(HaveLen(1))
		Expect(pkgFromLoader.Interfaces[0].Name).To(Equal("Display"))

		for i := range pkgFromReflect.Interfaces[0].Methods {
			expectMethodsEqual(pkgFromLoader.Interfaces[0].Methods[i], pkgFromReflect.Interfaces[0].Methods[i])
		}
	})

	It("generates a model with the basic properties", func() {
		pkg, e := loader.GenerateModel("github.com/petergtz/pegomock/modelgen/test_data/default_test_interface", "Display")
		Expect(e).NotTo(HaveOccurred())

		Expect(pkg.Name).To(Equal("test_interface"))
		Expect(pkg.Interfaces).To(HaveLen(1))
		Expect(pkg.Interfaces[0].Name).To(Equal("Display"))

		Expect(pkg.Interfaces[0].Methods).To(ContainElement(
			&model.Method{
				Name: "Show",
				In: []*model.Parameter{
					&model.Parameter{
						Name: "_param0",
						Type: model.PredeclaredType("string"),
					},
				},
			},
		))

		// TODO add more test cases
	})
})

func expectMethodsEqual(actual, expected *model.Method) {
	Expect(actual.Name).To(Equal(expected.Name))
	expectParamsEqual(actual.Name, actual.In, expected.In)
	expectParamsEqual(actual.Name, actual.Out, expected.Out)
}

func expectParamsEqual(methodName string, actual, expected []*model.Parameter) {
	for i := range expected {
		if actual[i].Name != expected[i].Name {
			fmt.Printf("Note: In method %v, param names differ \"%v\" != \"%v\"\n", methodName, actual[i].Name, expected[i].Name)
		}
		Expect(actual[i].Type).To(Equal(expected[i].Type))
	}
}
