# gomod-helpers

`gomod-helpers` is a helper tool to perform bulk updates of multiple go modules at once, especially when bumping
the module versions. This helper will help to bump the version using branch, tag or specific commit.

To resolve branches and tag, this helper will use Github and if that fails, it falls back to git clone (slow).
Github API is rate limited, so I encourage you to set `GITHUB_TOKEN` environment variable to bump the allowed requests limit.

### Installation

The easiest way to get `gomod-helpers` is to grab the binaries from the [release](https://github.com/mfojtik/goodmod/releases) page.
If you want to build this command yourself, you can clone this repository and run the `GOPATH=~/go make` command.

#### Usage

To replace all modules that starts with `k8s.io/` prefix to point to a commit under `kubernetes-1.16.2` tag, use:
```
$ gomod-helpers replace --tag=kubernetes-1.16.2 --paths=k8s.io/
```

To replace `github.com/openshift/library-go` module to point to a commit under `master` branch, use:
```
$ gomod-helpers replace --branch=master --paths=github.com/openshift/library-go
```

To replace all modules with `github.com/openshift/` prefix to point to a commit under `master` branch, use:
```
$ gomod-helpers replace --branch=master --paths=github.com/openshift/
```

**Note**: This command will **not** directly modify the `go.mod` file, but it will output a series of `go mod edit` commands
you can copy&paste to terminal, or you can pipe to `xargs`.

```cgo
$ gomod-helpers replace --tag=kubernetes-1.16.2 --paths=k8s.io/
go mod edit -replace k8s.io/api=k8s.io/api@"v0.0.0-20191016110408-35e52d86657a"
go mod edit -replace k8s.io/apiextensions-apiserver=k8s.io/apiextensions-apiserver@"v0.0.0-20191016113550-5357c4baaf65"
go mod edit -replace k8s.io/apimachinery=k8s.io/apimachinery@"v0.0.0-20191004115801-a2eda9f80ab8"
go mod edit -replace k8s.io/apiserver=k8s.io/apiserver@"v0.0.0-20191016112112-5190913f932d"
go mod edit -replace k8s.io/client-go=k8s.io/client-go@"v0.0.0-20191016111102-bec269661e48"
...
```

If you want `gomod-helpers` directly modify the `go.mod` file, you can pass the `--apply` flag.

#### `go-helpers.yaml`

In case you want to track what branches and tags you are following in your package, you can use the `go-helpers.yaml` file.
That file can include multiple rules to apply on the `go.mod` file. Check the [examples/](https://github.com/mfojtik/goodmod/tree/master/examples).
