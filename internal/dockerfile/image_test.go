package dockerfile

import "testing"

func TestSplitImage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ref, repo, reference string
		wantErr              bool
	}{
		{ref: "alpine", repo: "library/alpine", reference: "latest"},
		{ref: "alpine:3.20", repo: "library/alpine", reference: "3.20"},
		{ref: "jameswoolfenden/ghat", repo: "jameswoolfenden/ghat", reference: "latest"},
		{ref: "jameswoolfenden/ghat:v0.1.23", repo: "jameswoolfenden/ghat", reference: "v0.1.23"},
		{ref: "alpine@sha256:abc", repo: "library/alpine", reference: "sha256:abc"},
		{ref: "jameswoolfenden/ghat:v0.1.23@sha256:abc", repo: "jameswoolfenden/ghat", reference: "sha256:abc"},
		{ref: "docker.io/library/alpine:3.20", repo: "library/alpine", reference: "3.20"},
		{ref: "localhost:5000/foo", wantErr: true},
		{ref: "ghcr.io/x/y:1", wantErr: true},
		{ref: "", wantErr: true},
	}

	for _, tt := range tests {
		repo, reference, err := splitImage(tt.ref)
		if (err != nil) != tt.wantErr {
			t.Errorf("splitImage(%q) error = %v, wantErr %v", tt.ref, err, tt.wantErr)

			continue
		}

		if repo != tt.repo || reference != tt.reference {
			t.Errorf("splitImage(%q) = %q, %q; want %q, %q", tt.ref, repo, reference, tt.repo, tt.reference)
		}
	}
}
