package stevedore

import (
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func TestLabel(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			if got := Label(tt.args.result); got != tt.want {
				t.Errorf("Label() = %v, want %v", got, tt.want)
			}
		})
	}
}
