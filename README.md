# goodmod

`goodmod` is a tool that help Go developers that want to perform bulk replace in their `go.mod` files.

Example use cases are:
* A project is tracking specific tag for multiple go modules (eg. `kubernetes-1.16.2`)
* A project is tracking specific branch for multiple go modules (eg. `master`)
* A project is tracking specific commit for single go module

To resolve branches and tag, this tool use Github and as a falls back it will `git clone`.
I encourage to set `GITHUB_TOKEN` environment variable to your personal Github token to speed this tool up.
If you don't use the token, it will still try to use Github API, but you might see errors about being rate limited.

### Installation

The easiest way to get `goodmod` is to grab the binaries from the [release](https://github.com/mfojtik/goodmod/releases) page.
If you want to build this command yourself, you can clone this repository and run the `GOPATH=~/go make` command.

Alternatively, you can use:

```shell script
go install github.com/mfojtik/goodmod
```

#### `replace`

To replace all modules, except klog and utils, that starts with `k8s.io/` prefix to point to a commit under `kubernetes-1.16.2` tag, use:
```
$ goodmod replace --tag=kubernetes-1.16.2 --paths=k8s.io/* --excludes=k8s.io/klog,k8s.io/utils
```

To replace `github.com/openshift/library-go` module to point to a HEAD commit in `master` branch, use:
```
$ goodmod replace --branch=master --paths=github.com/openshift/library-go
```

To replace all modules with `github.com/openshift/` prefix to point to a commit under `master` branch, use:
```
$ goodmod replace --branch=master --paths=github.com/openshift/*
```

**Note**: By default, this command **not** directly modify the `go.mod` file, but it will output a series of `go mod edit -replace` commands
you can copy&paste to terminal, or you can pipe to `xargs`.

```shell script
$ goodmod replace --tag=kubernetes-1.16.2 --paths=k8s.io/
go mod edit -replace k8s.io/api=k8s.io/api@"v0.0.0-20191016110408-35e52d86657a"
go mod edit -replace k8s.io/apiextensions-apiserver=k8s.io/apiextensions-apiserver@"v0.0.0-20191016113550-5357c4baaf65"
go mod edit -replace k8s.io/apimachinery=k8s.io/apimachinery@"v0.0.0-20191004115801-a2eda9f80ab8"
go mod edit -replace k8s.io/apiserver=k8s.io/apiserver@"v0.0.0-20191016112112-5190913f932d"
go mod edit -replace k8s.io/client-go=k8s.io/client-go@"v0.0.0-20191016111102-bec269661e48"
...
```

If you want `goodmod replace` directly modify the `go.mod` file, you can pass the `--apply` flag.

#### `go-helpers.yaml`

In case you want to track what branches and tags you are following in your package, you can use the `go-helpers.yaml` file.
That file can include multiple rules to apply on the `go.mod` file. Check the [examples/](https://github.com/mfojtik/goodmod/tree/master/examples).

#### License

`goodmod` is licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/).
