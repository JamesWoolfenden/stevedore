package stevedore_test

import (
	"os/user"
	"reflect"
	"testing"

	stevedore "github.com/jameswoolfenden/stevedore/src"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func TestDockerfile_Label(t *testing.T) {
	type fields struct {
		Parsed *parser.Result
		Path   string
	}

	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"empty", fields{nil, "../examples/labelled/Dockerfile"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &stevedore.Dockerfile{
				Parsed: tt.fields.Parsed,
				Path:   tt.fields.Path,
			}
			got, err := result.Label()
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

func TestDockerfile_ParseFile(t *testing.T) {
	type fields struct {
		Parsed *parser.Result
		Path   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &stevedore.Dockerfile{
				Parsed: tt.fields.Parsed,
				Path:   tt.fields.Path,
			}
			if err := result.ParseFile(); (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
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

func TestDockerfile_GetDockerLabels(t *testing.T) {
	type fields struct {
		Parsed *parser.Result
		Path   string
		Image  string
	}

	var pass map[string]interface{}

	empty := make(map[string]interface{})

	tests := []struct {
		name    string
		fields  fields
		want    map[string]interface{}
		wantErr bool
	}{
		{"Pass", fields{nil, "", "jameswoolfenden/ghat"}, pass, false},
		{"Fail", fields{nil, "", "jameswoolfenden/guff"}, nil, true},
		{"library", fields{nil, "", "alpine"}, empty, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &stevedore.Dockerfile{
				Parsed: tt.fields.Parsed,
				Path:   tt.fields.Path,
				Image:  tt.fields.Image,
			}
			got, err := result.GetDockerLabels()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDockerLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDockerLabels() got = %v, want %v", got, tt.want)
			}
		})
	}
}
