package stevedore_test

import (
	stevedore "stevedore/src"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func TestLabel(t *testing.T) {
	t.Parallel()

	type args struct {
		result *parser.Result
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := stevedore.Label(tt.args.result); got != tt.want {
				t.Errorf("Label() = %v, want %v", got, tt.want)
			}
		})
	}
}
