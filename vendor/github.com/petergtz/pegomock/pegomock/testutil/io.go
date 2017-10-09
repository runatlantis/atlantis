package testutil

import (
	"io/ioutil"

	. "github.com/onsi/gomega"
)

func WriteFile(filepath string, content string) {
	Expect(ioutil.WriteFile(filepath, []byte(content), 0644)).To(Succeed())
}
