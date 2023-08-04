package stevedore_test

import (
	"testing"

	stevedore "github.com/jameswoolfenden/stevedore/src"
)

func TestParser_Parse(t *testing.T) {
	type fields struct {
		File      *string
		Output    string
		Directory string
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
			content := &stevedore.Parser{
				File:      tt.fields.File,
				Output:    tt.fields.Output,
				Directory: tt.fields.Directory,
			}
			if err := content.Parse(); (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParser_ParseAll(t *testing.T) {
	type fields struct {
		File      *string
		Output    string
		Directory string
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
			content := &stevedore.Parser{
				File:      tt.fields.File,
				Output:    tt.fields.Output,
				Directory: tt.fields.Directory,
			}
			if err := content.ParseAll(); (err != nil) != tt.wantErr {
				t.Errorf("ParseAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
