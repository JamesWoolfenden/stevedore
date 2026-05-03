package stevedore_test

import (
	"os"
	"os/user"
	"reflect"
	"strings"
	"testing"

	stevedore "github.com/jameswoolfenden/stevedore/src"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestDockerfile_Label(t *testing.T) {
	t.Parallel()

	type fields struct {
		Parsed *parser.Result
		Path   string
		Author string
	}

	want :=
		`FROM jameswoolfenden/ghat
WORKDIR /app
COPY . .
RUN yarn install --production
CMD ["node", "src/index.js"]
EXPOSE 3000
LABEL layer.0.author="James Woolfenden"`

	wantShort :=
		`FROM jameswoolfenden/ghat
WORKDIR /app
COPY . .
RUN yarn install --production
CMD ["node", "src/index.js"]
EXPOSE 3000`

	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{name: "empty", fields: fields{Parsed: nil, Path: "../examples/labelled/Dockerfile", Author: "James Woolfenden"}, wantErr: true},
		{"Pass", fields{Parsed: nil, Path: "../examples/basic/Dockerfile", Author: "James Woolfenden"}, want, false},
		{"Pass short", fields{Parsed: nil, Path: "../examples/basic/Dockerfile", Author: "James Woolfenden"}, wantShort, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, _ := os.Open(tt.fields.Path)
			tt.fields.Parsed, _ = parser.Parse(data)

			result := &stevedore.Dockerfile{
				Parsed: tt.fields.Parsed,
				Path:   tt.fields.Path,
			}
			got, err := result.Label(tt.fields.Author)
			if (err != nil) != tt.wantErr {
				t.Errorf("Label() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(got, tt.want, false)

			if !strings.Contains(got, tt.want) {
				temp := dmp.DiffPrettyText(diffs)
				t.Errorf("failed %s", temp)
			}
		})
	}
}

func TestDockerfile_ParseFile(t *testing.T) {
	t.Parallel()

	type fields struct {
		Parsed *parser.Result
		Path   string
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"Pass", fields{nil, "../examples/basic/Dockerfile"}, false},
		{"Not a file", fields{nil, "../examples/basic/"}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

func TestDockerfile_GetDockerLabels(t *testing.T) {
	t.Parallel()
	type fields struct {
		Parsed *parser.Result
		Path   string
		Image  string
	}

	var pass map[string]interface{}

	tests := []struct {
		name    string
		fields  fields
		want    map[string]interface{}
		wantErr bool
	}{
		{"Pass", fields{nil, "", "jameswoolfenden/ghat"}, pass, false},
		// Note: Newer Docker manifests don't have v1Compatibility history, so these return nil
		{"Fail", fields{nil, "", "jameswoolfenden/guff"}, nil, false},
		{"library", fields{nil, "", "alpine"}, nil, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
			// Allow nil or empty map as equivalent for modern manifests without history
			if tt.want == nil && len(got) == 0 {
				return // This is expected for modern Docker manifests
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDockerLabels() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeLabel(t *testing.T) {
	t.Parallel()
	//myUser, _ := user.Current()
	//file := "../examples/basic/Dockerfile"
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
		//{"Pass", args{nil, 0, myUser, 100, &file}, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := stevedore.MakeLabel(tt.args.child, tt.args.layer, tt.args.myUser, tt.args.endLine, tt.args.file); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}
