package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
)

var _ = Describe("Stopover", func() {
	var binPath string
	var session *gexec.Session
	var args []string
	var bearerTokenFromEnv = os.Getenv("CONCOURSE_BEARER_TOKEN")
	var bearerToken string

	BeforeSuite(func() {
		var err error
		binPath, err = gexec.Build("github.com/EngineerBetter/stopover")
		Ω(err).ShouldNot(HaveOccurred())
	})

	BeforeEach(func() {
		args = []string{}
		bearerToken = "CONCOURSE_BEARER_TOKEN=" + bearerTokenFromEnv
	})

	JustBeforeEach(func() {
		command := exec.Command(binPath, args...)
		command.Env = append(command.Env, bearerToken)
		var err error
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
	})

	Context("on the happy path", func() {
		var expected string

		BeforeEach(func() {
			expectedBytes, err := ioutil.ReadFile("expected_output.yml")
			Ω(err).ShouldNot(HaveOccurred())
			expected = string(expectedBytes)

			args = []string{"https://arthropods.dpsas.io", "cf-ops", "ant", "opsman-apply-changes", "1"}
		})

		It("outputs a YAML file of resource versions", func() {
			Eventually(session).Should(gexec.Exit(0))
			Ω(session).Should(Say(expected))
		})
	})

	var usage = regexp.QuoteMeta(`** Error: arguments not found
Usage:
$ export CONCOURSE_BEARER_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJj....."
$ stopover https://ci.server.tld my-team my-pipeline my-job job-build-id`)

	Context("when no arguments are provided", func() {
		It("exits 1 and prints usage", func() {
			Eventually(session).Should(gexec.Exit(1))
			Ω(session.Err).Should(Say(usage))
		})
	})

	Context("when the envvar CONCOURSE_BEARER_TOKEN is not set", func() {
		BeforeEach(func() {
			bearerToken = ""
		})

		It("exits 1 and prints usage", func() {
			Eventually(session).Should(gexec.Exit(1))
			Ω(session.Err).Should(Say(usage))
		})
	})

	Context("when an argument is missing", func() {
		BeforeEach(func() {
			args = []string{"https://arthropods.dpsas.io", "pipeline", "job", "1"}
		})

		It("exits 1", func() {
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when the URL is invalid", func() {
		BeforeEach(func() {
			args = []string{"not-valid-url", "team", "pipeline", "job", "1"}
		})

		It("exits 1 with a useful error message", func() {
			Eventually(session).Should(gexec.Exit(1))
			Ω(session.Err).Should(Say("could not get build for job"))
		})
	})
})
