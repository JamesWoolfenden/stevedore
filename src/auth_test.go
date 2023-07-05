package stevedore

import "testing"

func TestGetAuthToken(t *testing.T) {
	type args struct {
		from string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAuthToken(tt.args.from)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAuthToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetAuthToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}
