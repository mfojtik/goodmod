# gomod-helpers

This repository contain various helpers that should help developers working with `go mod`.

#### Usage

* Bulk replacing/pinning to specific version/tag/commit:

```bash
gomod-helpers replace --tag=kubernetes-1.16.1 --paths=k8s.io/ --exclude=k8s.io/gengo
# Setting version "v0.0.0-2019102162552-e32f40effea1" for repository "k8s.io/component-base" ...
# Setting version "v0.0.0-2019102162552-33a2f62922c7" for repository "k8s.io/code-generator" ...
# Setting version "v0.0.0-2019102162552-e15c9dfa956a" for repository "k8s.io/apimachinery" ...
# Setting version "v0.0.0-2019102162552-813d2abdde12" for repository "k8s.io/apiserver" ...
# Setting version "v0.0.0-2019102162552-ef11380c2891" for repository "k8s.io/client-go" ...
# Setting version "v0.0.0-2019102162552-aec909fb44be" for repository "k8s.io/kube-aggregator" ...
# Setting version "v0.0.0-2019102162552-d348b21713e8" for repository "k8s.io/api" ...
go mod edit -replace k8s.io/api=k8s.io/api@"v0.0.0-2019102162552-d348b21713e8"
go mod edit -replace k8s.io/apimachinery=k8s.io/apimachinery@"v0.0.0-2019102162552-e15c9dfa956a"
go mod edit -replace k8s.io/apiserver=k8s.io/apiserver@"v0.0.0-2019102162552-813d2abdde12"
go mod edit -replace k8s.io/client-go=k8s.io/client-go@"v0.0.0-2019102162552-ef11380c2891"
go mod edit -replace k8s.io/code-generator=k8s.io/code-generator@"v0.0.0-2019102162552-33a2f62922c7"
go mod edit -replace k8s.io/component-base=k8s.io/component-base@"v0.0.0-2019102162552-e32f40effea1"
go mod edit -replace k8s.io/kube-aggregator=k8s.io/kube-aggregator@"v0.0.0-2019102162552-aec909fb44be"
```
