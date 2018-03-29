package main_test

import (
	"github.com/concourse/fly/rc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/SpectoLabs/hoverfly/core"
	"github.com/SpectoLabs/hoverfly/core/handlers/v2"
	"github.com/SpectoLabs/hoverfly/core/modes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Stopover", func() {
	simulation_path := "./fixtures/http_conversation.json"

	var binPath string
	var session *gexec.Session
	var args []string
	var bearerTokenEnvVar string
	var port string
	var hfly *hoverfly.Hoverfly

	BeforeSuite(func() {
		var err error
		binPath, err = gexec.Build("github.com/EngineerBetter/stopover")
		Ω(err).ShouldNot(HaveOccurred())

		config := hoverfly.InitSettings()
		config.TLSVerification = false

		hfly = hoverfly.NewHoverflyWithConfiguration(config)

		config.SetMode(modes.Simulate)

		bytes, err := ioutil.ReadFile(simulation_path)
		Ω(err).ShouldNot(HaveOccurred())

		simulation := v2.SimulationViewV4{}
		err = json.Unmarshal(bytes, &simulation)
		Ω(err).ShouldNot(HaveOccurred())

		err = hfly.PutSimulation(simulation)
		Ω(err).ShouldNot(HaveOccurred())

		err = hfly.StartProxy()
		Ω(err).ShouldNot(HaveOccurred())
		port = hfly.Cfg.ProxyPort
		err = rc.SaveTarget(rc.TargetName("stopover_test"), "https://arthropods.dpsas.io", true, "cf-ops", nil, "")
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterSuite(func() {
		err := rc.DeleteTarget(rc.TargetName("stopover_test"))
		Ω(err).ShouldNot(HaveOccurred())

		hfly.StopProxy()
	})

	JustBeforeEach(func() {
		command := exec.Command(binPath, args...)
		command.Env = append(os.Environ(), bearerTokenEnvVar, "HTTP_PROXY=http://localhost:"+port, "HTTPS_PROXY=https://localhost:"+port)
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

			args = []string{"-target", "stopover_test", "-job", "ant/opsman-apply-changes", "-build", "1"}
		})

		It("outputs a YAML file of resource versions", func() {
			Eventually(session).Should(gexec.Exit(0))
			Ω(session).Should(Say(expected))
		})
	})

	Context("when no arguments are provided", func() {
		BeforeEach(func() {
			args = nil
		})
		It("exits 1 and prints usage", func() {
			Eventually(session).Should(gexec.Exit(1))
			Ω(session.Err).Should(Say("Usage"))
		})
	})

	Context("when the target is not present", func() {
		BeforeEach(func() {
			args = []string{"-target", "stopover_test_not_present", "-job", "ant/opsman-apply-changes", "-build", "1"}
		})

		It("exits 1 and prints usage", func() {
			Eventually(session).Should(gexec.Exit(1))
			Ω(session.Err).Should(Say("unknown target"))
		})
	})

	Context("when an argument is missing", func() {
		BeforeEach(func() {
			args = []string{"-target", "stopover_test_not_present", "-job", "ant/opsman-apply-changes"}
		})

		It("exits 1", func() {
			Eventually(session).Should(gexec.Exit(1))
		})
	})
})
