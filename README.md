# stopover

Emits a YAML file listing versions of all resources for a given
Concourse build.

## Usage

```
$ export ATC_BEARER_TOKEN=foo
$ stopover https://ci.domain.com team-name pipeline job build-number
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

