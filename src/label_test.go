package stevedore_test

import (
	"os/user"
	"reflect"
	"testing"

	stevedore "github.com/jameswoolfenden/stevedore/src"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func TestLabel(t *testing.T) {
	t.Parallel()

	type args struct {
		result *parser.Result
		file   *string
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
			if got := stevedore.Label(tt.args.result, tt.args.file); got != tt.want {
				t.Errorf("Label() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeLabel(t *testing.T) {
	type args struct {
		child   *parser.Node
		layer   int64
		myUser  *user.User
		endLine int
		file    *string
	}
	tests := []struct {
		name string
		args args
		want *parser.Node
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stevedore.MakeLabel(tt.args.child, tt.args.layer, tt.args.myUser, tt.args.endLine, tt.args.file); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}
