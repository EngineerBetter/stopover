package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/concourse/atc"
	flyrc "github.com/concourse/fly/rc"
	"github.com/concourse/go-concourse/concourse"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var (
	target  = flag.String("target", "", "fly target")
	jobFlag = flag.String("job", "", "PIELINE/JOB")
	build   = flag.String("build", "", "build number")
)

func main() {
	flag.Parse()

	if *target == "" {
		printUsageAndExit(1)
	}
	t, err := flyrc.LoadTarget(flyrc.TargetName(*target), false)
	exitIfErr(err)

	splitJob := strings.SplitN(*jobFlag, "/", 2)
	if len(splitJob) != 2 {
		printUsageAndExit(1)
	}
	pipeline := splitJob[0]
	job := splitJob[1]
	resourceVersions, err := GetResourceVersions(t.Client(), t.Team(), pipeline, job, *build)
	exitIfErr(err)
	yaml, err := GenerateYaml(resourceVersions)
	exitIfErr(err)
	fmt.Print(string(yaml))
}

func GetResourceVersions(client concourse.Client, team concourse.Team, pipelineName, jobName, buildName string) (map[string]atc.Version, error) {
	build, found, err := team.JobBuild(pipelineName, jobName, buildName)

	if !found || err != nil {
		return nil, errors.New("could not get build for job")
	}

	globalID := build.ID
	buildInputsOutputs, found, err := client.BuildResources(globalID)

	if !found || err != nil {
		return nil, errors.New("could not get resources for build with global ID " + strconv.Itoa(globalID))
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
	usage := `Usage:
$ fly login -t foo -c https://ci.server.tld
$ stopover -target foo -job my-pipeline/my-job -build job-build-id`
	fmt.Fprintln(os.Stderr, usage)
	os.Exit(status)
}
