package turbo_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTurbo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Turbo Suite")
}
