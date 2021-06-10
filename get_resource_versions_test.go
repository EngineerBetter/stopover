package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/concourse/concourse/go-concourse/concourse/concoursefakes"
	yaml "gopkg.in/yaml.v2"
)

var _ = Describe("GetResourceVersions", func() {

	var teamName = "main"
	var pipelineName = "control-tower"
	var jobName = "minor"
	var buildName = "1"

	var expectedStruct map[string]atc.Version
	var client *concoursefakes.FakeClient

	BeforeEach(func() {
		expectedStruct = map[string]atc.Version{}
		expectedBytes, err := ioutil.ReadFile("./fixtures/expected_output.yml")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(expectedBytes).ShouldNot(BeEmpty())
		err = yaml.Unmarshal(expectedBytes, expectedStruct)
		Ω(err).ShouldNot(HaveOccurred())

		fakeTeam := new(concoursefakes.FakeTeam)
		fakeTeam.JobBuildStub = func(pipeline atc.PipelineRef, job, build string) (atc.Build, bool, error) {
			if pipeline.Name == "control-tower" && job == "minor" && build == "1" {
				return atc.Build{ID: 2098}, true, nil
			}

			return atc.Build{}, false, nil
		}

		wrongTeam := new(concoursefakes.FakeTeam)
		wrongTeam.JobBuildStub = func(pipeline atc.PipelineRef, job, build string) (atc.Build, bool, error) {
			return atc.Build{}, false, nil
		}

		client = new(concoursefakes.FakeClient)
		client.TeamStub = func(teamName string) concourse.Team {
			if teamName == "main" {
				return fakeTeam
			}

			return wrongTeam
		}

		client.BuildResourcesStub = func(buildID int) (atc.BuildInputsOutputs, bool, error) {
			if buildID == 2098 {
				return atc.BuildInputsOutputs{
					Inputs: []atc.PublicBuildInput{
						{
							Name: "control-tower-ops",
							Version: atc.Version{
								"commit": "407f8ab92a7258cbae32d1ad987b64f8d18a9a3a",
								"ref":    "0.0.8",
							},
						},
						{
							Name: "pcf-ops",
							Version: atc.Version{
								"digest": "sha256:8a4f9f1647080c224f015cc655146fda7329baa8c7b279b597dee114a69ff97a",
							},
						},
						{
							Name: "version",
							Version: atc.Version{
								"number": "0.2.0",
							},
						},
						{
							Name: "control-tower",
							Version: atc.Version{
								"ref": "244a2df8b612d8e9b560ba73023d7673b5d4d007",
							},
						},
					},
				}, true, nil
			}

			return atc.BuildInputsOutputs{}, false, nil
		}
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
