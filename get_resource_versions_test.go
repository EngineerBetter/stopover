package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/atc"
	"github.com/concourse/go-concourse/concourse"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var _ = Describe("GetResourceVersions", func() {

	var bearerToken = os.Getenv("CONCOURSE_BEARER_TOKEN")
	var teamName = "cf-ops"
	var pipelineName = "ant"
	var jobName = "opsman-apply-changes"
	var buildName = "1"

	var expectedStruct map[string]atc.Version
	var client concourse.Client

	BeforeEach(func() {
		expectedStruct = map[string]atc.Version{}
		expectedBytes, err := ioutil.ReadFile("expected_output.yml")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(expectedBytes).ShouldNot(BeEmpty())
		err = yaml.Unmarshal(expectedBytes, expectedStruct)
		Ω(err).ShouldNot(HaveOccurred())

		client = NewClient("https://arthropods.dpsas.io", bearerToken, true)
	})

	It("returns the expected stuff", func() {
		resourceVersions, err := GetResourceVersions(client, teamName, pipelineName, jobName, buildName)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(resourceVersions).Should(Equal(expectedStruct))
	})

	Context("when the team does not exist", func() {
		It("returns an error", func() {
			resourceVersions, err := GetResourceVersions(client, "does-not-exist", pipelineName, jobName, buildName)
			Ω(err).Should(HaveOccurred())
			Ω(resourceVersions).Should(BeNil())
		})
	})

	Context("when the pipeline does not exist", func() {
		It("returns an error", func() {
			resourceVersions, err := GetResourceVersions(client, teamName, "does-not-exist", jobName, buildName)
			Ω(err).Should(HaveOccurred())
			Ω(resourceVersions).Should(BeNil())
		})
	})

	Context("when the job does not exist", func() {
		It("returns an error", func() {
			resourceVersions, err := GetResourceVersions(client, teamName, pipelineName, "does-not-exist", buildName)
			Ω(err).Should(HaveOccurred())
			Ω(resourceVersions).Should(BeNil())
		})
	})

	Context("when the build does not exist", func() {
		It("returns an error", func() {
			resourceVersions, err := GetResourceVersions(client, teamName, pipelineName, jobName, "does-not-exist")
			Ω(err).Should(HaveOccurred())
			Ω(resourceVersions).Should(BeNil())
		})
	})

})
