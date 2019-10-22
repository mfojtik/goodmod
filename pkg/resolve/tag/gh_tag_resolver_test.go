package tag

import (
	"context"
	"testing"
)

func TestGithubTagResolver(t *testing.T) {
	r := &GithubTagResolver{}
	c, err := r.Resolve(context.TODO(), "https://github.com/kubernetes/apiserver", "kubernetes-1.16.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.SHA) == 0 {
		t.Errorf("unable to get SHA")
	}
}
