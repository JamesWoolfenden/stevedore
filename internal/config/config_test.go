package config

import "testing"

func TestValidateDockerfilePath(t *testing.T) {
	t.Parallel()

	for path, wantErr := range map[string]bool{
		"":                  true,
		"Dockerfile":        false,
		"../Dockerfile":     false,
		"a/../b/Dockerfile": false,
		"..":                false,
	} {
		if err := ValidateDockerfilePath(path); (err != nil) != wantErr {
			t.Errorf("ValidateDockerfilePath(%q) error = %v, wantErr %v", path, err, wantErr)
		}
	}
}

func TestSetupLoggingRejectsBadLevel(t *testing.T) {
	if err := (&Config{LogLevel: "loud"}).SetupLogging(); err == nil {
		t.Error("expected an error for an unknown log level")
	}

	if err := NewConfig().SetupLogging(); err != nil {
		t.Errorf("default config: %v", err)
	}
}
