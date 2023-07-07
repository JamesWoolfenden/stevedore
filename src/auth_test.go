package stevedore_test

import (
	stevedore "stevedore/src"
	"testing"
)

func TestGetAuthToken(t *testing.T) {
	t.Parallel()

	type args struct {
		from string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"this", args{"jameswoolfenden/stevedore"}, "anything", false},
		{"rubbish", args{"jameswoolfenden/notarepo"}, "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := stevedore.GetAuthToken(tt.args.from)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAuthToken() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != "" {
				if tt.want != "" {
					t.Errorf("GetAuthToken() got = %v, want %v", got, tt.want)
				}
			}

		})
	}
}
