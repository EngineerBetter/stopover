# stopover

Emits a YAML file listing versions of all resources for a given
Concourse build.

## Usage

```
$ export ATC_BEARER_TOKEN=foo
$ stopover https://ci.domain.com team-name pipeline job build-number
```

## Using `stopover` to pin Resource Versions

You can automatically pin Concourse pipelines to specific versions of resources by using a file created by`stopover`.

If you `put` this to a resource, then _other_ pipelines can be triggered by changes to this version file. These pipelines can then call `fly set-pipeline` on themselves or others, providing a `load-vars-from` flag pointing to the resource versions file. If the pipelines being set have their `version` blocks parameterised, then this will result in the pipeline _only_ being able to use those specified versions.

### Getting the versions file

First, you'll need to emit a `stopover` versions YAML file:

* Create a job that runs `stopover`
* Provide that job a valid ATC bearer token (maybe by using [fly-github-auth-task](https://github.com/EngineerBetter/fly-github-auth-task))
* Provide that job metadata about the current build
* Provide `get`s to that job, for each resource you want to snapshot. Remember to used `passed` if required.
* Invoke `stopover`, passing in all the above
* `put` the resultant file somewhere

An example version-snapshotting job looks like this:

```
jobs:
- name: snapshot-versions
  plan:
  - get : some-repo
    passed: [jobs, that, should, have, passed]
    trigger: true
  - get: resource-we-want-to-track
    passed: [some-test]
  - get: p-another-resource-we-want-to-track
    passed: [another-test]
  - get: bearer-token
  - get: metadata
  - task: generate-versions
    file: config-repo/tasks/generate-versions.yml
    params:
      ATC_URL: ((atc_url))
  - put: resource-versions
    params:
      file: resources/versions.yml
      acl: private

resources:
- name: metadata
  type: build-metadata
```

The corresponding task is thus:

```
export ATC_BEARER_TOKEN=$(<bearer-token/bearer-token)
export PIPELINE_NAME=$(<metadata/build-pipeline-name)
stopover ${ATC_URL} team-name ${PIPELINE_NAME} snapshot-versions $(cat metadata/build-name) > resources/versions.yml
```

### Using the versions file

An example job that calls `fly set-pipeline` on itself:

```
- name: set-pipeline
  plan:
  - get: some-repo
    trigger: true
    version: ((resource_version_some-repo))
  - get: upstream-resource-versions
    trigger: true
  - get: bearer-token
  - get: metadata
  - task: fly-set-pipeline
    config:
      inputs:
      - name: bearer-token
      - name: some-repo
      - name: metadata
      - name: upstream-resource-versions
      run:
        path: /bin/bash
        args:
        - -euc
        - |
          export PIPELINE_NAME=$(<metadata/build-pipeline-name)

          # Write bearer token to ~/.flyrc, ommitted for brevity

          fly -t target set-pipeline \
              "$@" \
              --pipeline ${PIPELINE_NAME} \
              --config some-repo/pipeline.yml \
              --load-vars-from some-repo/vars/${PIPELINE_NAME}.yml \
              --load-vars-from upstream-resource-versions/versions.yml \
              --load-vars-from secrets.yml
      outputs:
      - name: resources
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: engineerbetter/pcf-ops
```

## Testing

To test using saved HTTP requests/responses:

```
$ ginkgo
```

We use Hoverfly as a library, which can be set to record or replay. If
you need to re-record the HTTP conversation, specify a valid bearer token:

```
$ ATC_BEARER_TOKEN=foo ginkgo
```

You can find the bearer token for the ATC you want to test against in
your `~/.flyrc` file (assuming you're logged in via `fly`).

