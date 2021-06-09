package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	hoverfly "github.com/SpectoLabs/hoverfly/core"
	v2 "github.com/SpectoLabs/hoverfly/core/handlers/v2"
	"github.com/SpectoLabs/hoverfly/core/modes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Stopover", func() {
	simulation_path := "./fixtures/http_conversation.json"

	var binPath string
	var session *gexec.Session
	var args []string
	var bearerTokenFromEnv = os.Getenv("ATC_BEARER_TOKEN")
	var bearerTokenEnvVar string
	var port string
	var hfly *hoverfly.Hoverfly
	var recording bool

	BeforeSuite(func() {
		var err error
		binPath, err = gexec.Build("github.com/EngineerBetter/stopover")
		Ω(err).ShouldNot(HaveOccurred())

		_, recording = os.LookupEnv("ATC_BEARER_TOKEN")

		config := hoverfly.InitSettings()
		config.TLSVerification = false

		hfly = hoverfly.NewHoverflyWithConfiguration(config)

		if recording {
			config.SetMode(modes.Capture)
		} else {
			config.SetMode(modes.Simulate)

			bytes, err := ioutil.ReadFile(simulation_path)
			Ω(err).ShouldNot(HaveOccurred())

			simulation := v2.SimulationViewV4{}
			err = json.Unmarshal(bytes, &simulation)
			Ω(err).ShouldNot(HaveOccurred())

			err = hfly.PutSimulation(simulation)
			Ω(err).ShouldNot(HaveOccurred())
		}

		err = hfly.StartProxy()
		Ω(err).ShouldNot(HaveOccurred())
		port = hfly.Cfg.ProxyPort
	})

	AfterSuite(func() {
		if recording {
			simulation, err := hfly.GetSimulation()
			Ω(err).ShouldNot(HaveOccurred())

			bytes, err := json.Marshal(simulation)
			Ω(err).ShouldNot(HaveOccurred())

			err = ioutil.WriteFile(simulation_path, bytes, 0644)
			Ω(err).ShouldNot(HaveOccurred())
		}

		hfly.StopProxy()
	})

	BeforeEach(func() {
		args = []string{}
		if recording {
			bearerTokenEnvVar = "ATC_BEARER_TOKEN=" + bearerTokenFromEnv
		} else {
			bearerTokenEnvVar = "ATC_BEARER_TOKEN=dummy-value"
		}

	})

	JustBeforeEach(func() {
		command := exec.Command(binPath, args...)
		command.Env = append(command.Env, bearerTokenEnvVar, "HTTP_PROXY=http://localhost:"+port, "HTTPS_PROXY=http://localhost:"+port)
		var err error
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
	})

	Context("on the happy path", func() {
		var expected string

		BeforeEach(func() {
			expectedBytes, err := ioutil.ReadFile("./fixtures/expected_output.yml")
			Ω(err).ShouldNot(HaveOccurred())
			expected = string(expectedBytes)

			args = []string{"https://ci.engineerbetter.com", "main", "control-tower", "minor", "1"}
		})

		It("outputs a YAML file of resource versions", func() {
			Eventually(session).Should(Say(expected))
			Ω(session).Should(gexec.Exit(0))
		})
	})

	var usage = regexp.QuoteMeta(`** Error: arguments not found
Usage:
$ export ATC_BEARER_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJj....."
$ stopover https://ci.server.tld my-team my-pipeline my-job job-build-id`)

	Context("when no arguments are provided", func() {
		It("exits 1 and prints usage", func() {
			Eventually(session).Should(gexec.Exit(1))
			Ω(session.Err).Should(Say(usage))
		})
	})

	Context("when the envvar ATC_BEARER_TOKEN is not set", func() {
		BeforeEach(func() {
			bearerTokenEnvVar = ""
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
			Ω(session.Err).Should(Say("error getting build for job \\[Get not-valid-url/api/v1/teams/team/pipelines/pipeline/jobs/job/builds/1: unsupported protocol scheme \"\"\\]"))
		})
	})
})
