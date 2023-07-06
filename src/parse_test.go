package stevedore

import (
	"reflect"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func TestGetDockerLabels(t *testing.T) {
	type args struct {
		from string
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
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDockerLabels(tt.args.from)
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
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetParentLabels(tt.args.from, tt.args.version, tt.args.token)
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
	type args struct {
		file *string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Parse(tt.args.file); (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseAll(t *testing.T) {
	type args struct {
		file      *string
		directory string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseAll(tt.args.file, tt.args.directory); (err != nil) != tt.wantErr {
				t.Errorf("ParseAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFile(tt.args.file)
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
