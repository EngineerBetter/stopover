package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestStopover(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stopover Suite")
}
