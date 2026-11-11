package consts

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConsts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Consts Suite")
}
