package stevedore_test

import (
	"reflect"
	stevedore "stevedore/src"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func TestGetDockerLabels(t *testing.T) {
	t.Parallel()

	type args struct {
		from string
	}

	pass := map[string]interface{}{
		//	"layer.0.author": "JamesWoolfenden",
	}

	empty := make(map[string]interface{})

	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{"Pass", args{"jameswoolfenden/ghat"}, pass, false},
		{"Fail", args{"jameswoolfenden/guff"}, nil, true},
		{"library", args{"alpine"}, empty, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := stevedore.GetDockerLabels(tt.args.from)
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

func TestGetParentLabels(t *testing.T) {
	t.Parallel()

	type args struct {
		from    string
		version string
		token   string
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := stevedore.GetParentLabels(tt.args.from, tt.args.version, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetParentLabels() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetParentLabels() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	type args struct {
		file   *string
		output string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := stevedore.Parse(tt.args.file, tt.args.output); (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseAll(t *testing.T) {
	t.Parallel()

	type args struct {
		file      *string
		directory string
		output    string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := stevedore.ParseAll(tt.args.file, tt.args.directory, tt.args.output); (err != nil) != tt.wantErr {
				t.Errorf("ParseAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	t.Parallel()

	type args struct {
		file string
	}

	tests := []struct {
		name    string
		args    args
		want    *parser.Result
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := stevedore.ParseFile(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
