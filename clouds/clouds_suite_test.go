package clouds

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClouds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clouds Suite")
}
