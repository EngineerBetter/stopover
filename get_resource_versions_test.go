package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/atc"
	"github.com/concourse/go-concourse/concourse"
	"github.com/concourse/go-concourse/concourse/concoursefakes"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var _ = Describe("GetResourceVersions", func() {

	var teamName = "cf-ops"
	var pipelineName = "ant"
	var jobName = "opsman-apply-changes"
	var buildName = "1"

	var expectedStruct map[string]atc.Version
	var client *concoursefakes.FakeClient

	BeforeEach(func() {
		expectedStruct = map[string]atc.Version{}
		expectedBytes, err := ioutil.ReadFile("expected_output.yml")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(expectedBytes).ShouldNot(BeEmpty())
		err = yaml.Unmarshal(expectedBytes, expectedStruct)
		Ω(err).ShouldNot(HaveOccurred())

		fakeTeam := new(concoursefakes.FakeTeam)
		fakeTeam.JobBuildStub = func(pipeline, job, build string) (atc.Build, bool, error) {
			if pipeline == "ant" && job == "opsman-apply-changes" && build == "1" {
				return atc.Build{ID: 2098}, true, nil
			}

			return atc.Build{}, false, nil
		}

		wrongTeam := new(concoursefakes.FakeTeam)
		wrongTeam.JobBuildStub = func(pipeline, job, build string) (atc.Build, bool, error) {
			return atc.Build{}, false, nil
		}

		client = new(concoursefakes.FakeClient)
		client.TeamStub = func(teamName string) concourse.Team {
			if teamName == "cf-ops" {
				return fakeTeam
			}

			return wrongTeam
		}

		client.BuildResourcesStub = func(buildID int) (atc.BuildInputsOutputs, bool, error) {
			if buildID == 2098 {
				return atc.BuildInputsOutputs{
					Inputs: []atc.PublicBuildInput{
						{
							Resource: "config-repo",
							Version: atc.Version{
								"ref": "14a1260f1fe39088fcde02fedb8de30da435dd61",
							},
						},
						{
							Resource: "ert-config",
							Version: atc.Version{
								"ref": "d3fa445682b68c6869fa9866ef69b31de0ec25ce",
							},
						},
						{
							Resource: "om-cli",
							Version: atc.Version{
								"version_id": "Q.6uemV1Y5FPnTo5sAQbwHINdMPnEoaP",
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
