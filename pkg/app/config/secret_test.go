package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecret_Load(t *testing.T) {
	const envKey = "TEST_SECRET_VAR"
	old := os.Getenv(envKey)
	defer func() { _ = os.Setenv(envKey, old) }()

	tests := []struct {
		name    string
		secret  Secret
		setEnv  string
		want    string
		wantErr bool
	}{
		{
			name:    "env not set",
			secret:  Secret{Env: envKey},
			setEnv:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "env set",
			secret:  Secret{Env: envKey},
			setEnv:  "my-secret-value",
			want:    "my-secret-value",
			wantErr: false,
		},
		{
			name:    "secret config not set",
			secret:  Secret{},
			setEnv:  "",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv != "" {
				_ = os.Setenv(envKey, tt.setEnv)
			} else {
				_ = os.Unsetenv(envKey)
			}
			got, err := tt.secret.Load()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
