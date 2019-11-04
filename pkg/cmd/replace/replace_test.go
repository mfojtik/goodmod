package replace

import (
	"testing"
)

func TestPathGlob(t *testing.T) {
	tests := []struct {
		path     string
		rulePath string
		match    bool
	}{
		{
			rulePath: "k8s.io/api",
			path:     "k8s.io/apiserver",
			match:    false,
		},
		{
			rulePath: "k8s.io/api*",
			path:     "k8s.io/apiserver",
			match:    true,
		},
		{
			rulePath: "k8s.io/*",
			path:     "k8s.io/apiserver",
			match:    true,
		},
		{
			rulePath: "github.com/openshift",
			path:     "github.com/openshift/library-go",
			match:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			if match := matchRulePath(test.path, test.rulePath); match != test.match {
				t.Errorf("expected match=%t, got %t ", test.match, match)
			}
		})
	}
}
