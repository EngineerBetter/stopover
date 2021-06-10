package main_test

import (
	. "github.com/FidelityInternational/stopover"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/concourse/atc"
	"gopkg.in/yaml.v2"
)

var _ = Describe("GenerateYaml", func() {
	It("generates yaml that can be interpreted", func() {

		resourceVersions := map[string]atc.Version{
			"resource1": {"ref": "sha1"},
			"resource2": {"ref": "sha1", "thing": "version-foo"},
		}

		yamlBytes, err := GenerateYaml(resourceVersions)
		Ω(err).ShouldNot(HaveOccurred())

		actual := map[string]atc.Version{}
		err = yaml.Unmarshal(yamlBytes, actual)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(actual).Should(Equal(resourceVersions))
	})
})
