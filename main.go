package main

import (
	"fmt"
	"github.com/concourse/atc"
	"github.com/concourse/go-concourse/concourse"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) != 6 || os.Getenv("CONCOURSE_BEARER_TOKEN") == "" {
		printUsageAndExit(1)
	}

	bearerToken := os.Getenv("CONCOURSE_BEARER_TOKEN")
	url := os.Args[1]
	team := os.Args[2]
	pipeline := os.Args[3]
	job := os.Args[4]
	build := os.Args[5]

	client := NewClient(url, bearerToken, true)
	resourceVersions, err := GetResourceVersions(client, team, pipeline, job, build)
	exitIfErr(err)
	yaml, err := GenerateYaml(resourceVersions)
	exitIfErr(err)
	fmt.Print(string(yaml))
}

func NewClient(url, bearerToken string, ignoreTls bool) concourse.Client {
	// Initialise the default client before modifying its Transport in place
	// Panic occurs if this isn't done
	_, _ = http.DefaultClient.Get("http://127.0.0.1")

	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig.InsecureSkipVerify = ignoreTls

	oAuthToken := &oauth2.Token{
		AccessToken: bearerToken,
		TokenType:   "Bearer",
	}

	transport := &oauth2.Transport{
		Source: oauth2.StaticTokenSource(oAuthToken),
		Base:   tr,
	}

	httpClient := &http.Client{Transport: transport}

	return concourse.NewClient(url, httpClient)
}

func GetResourceVersions(client concourse.Client, teamName, pipelineName, jobName, buildName string) (map[string]atc.Version, error) {
	team := client.Team(teamName)

	build, found, err := team.JobBuild(pipelineName, jobName, buildName)

	if !found || err != nil {
		return nil, errors.New("could not get build for job")
	}

	globalID := build.ID
	buildInputsOutputs, found, err := client.BuildResources(globalID)

	if !found || err != nil {
		return nil, errors.New("could not get resources for build with global ID " + string(globalID))
	}

	resourceVersions := make(map[string]atc.Version)
	for _, input := range buildInputsOutputs.Inputs {
		key := "resource_version_" + input.Resource
		resourceVersions[key] = input.Version
	}

	return resourceVersions, nil
}

func GenerateYaml(resourceVersions map[string]atc.Version) ([]byte, error) {
	return yaml.Marshal(resourceVersions)
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printUsageAndExit(status int) {
	usage := `** Error: arguments not found
Usage:
$ export CONCOURSE_BEARER_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJj....."
$ stopover https://ci.server.tld my-team my-pipeline my-job job-build-id`
	fmt.Fprintln(os.Stderr, usage)
	os.Exit(status)
}
