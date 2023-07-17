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

	file := "../examples/labelled/Dockerfile"

	var empty *parser.Result

	type args struct {
		result *parser.Result
		file   *string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{empty, &file}, "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := stevedore.Label(tt.args.result, tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Label() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Label() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeLabel(t *testing.T) {
	t.Parallel()

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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := stevedore.MakeLabel(tt.args.child, tt.args.layer, tt.args.myUser, tt.args.endLine, tt.args.file); !reflect.DeepEqual(got, tt.want) { //nolint:lll
				t.Errorf("MakeLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}
