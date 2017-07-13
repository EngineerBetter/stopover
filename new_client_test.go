package main_test

import (
	. "github.com/EngineerBetter/stopover"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("NewClient", func() {
	var url = "https://arthropods.dpsas.io"
	var bearerToken = os.Getenv("CONCOURSE_BEARER_TOKEN")
	var ignoreTls = true
	var teamName = "cf-ops"

	It("returns a client using the specified URL", func() {
		client := NewClient(url, bearerToken, ignoreTls)
		Ω(client).ShouldNot(BeNil())
		Ω(client.URL()).Should(Equal(url))
	})

	It("returns a client which can talk to the given server", func() {
		client := NewClient(url, bearerToken, ignoreTls)
		info, err := client.GetInfo()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(info.Version).ShouldNot(BeEmpty())
		Ω(info.WorkerVersion).ShouldNot(BeEmpty())
	})

	It("returns a client which can make authenticated API requests", func() {
		client := NewClient(url, bearerToken, ignoreTls)
		pipelines, err := client.Team(teamName).ListPipelines()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(pipelines).ShouldNot(BeEmpty())
	})
})
