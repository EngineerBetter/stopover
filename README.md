# stopover

Emits a YAML file listing versions of all resources for a given Concourse build.

## Testing

You'll need a Bearer token for the ATC you want to test against. You can
find one in your `~/.flyrc` file.

```
$ CONCOURSE_BEARER_TOKEN=foo.bar.baz ginkgo
```