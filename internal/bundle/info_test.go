package bundle

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseInfo(t *testing.T) {
	wantTime, _ := time.Parse(time.RFC3339, "2023-01-04T22:53:22Z")

	tests := map[string]struct {
		InFile  string
		Want    Info
		WantErr string
	}{
		"v1": {
			InFile: "elastic-agent-versionyaml",
			Want: Info{
				BuildTime: wantTime,
				Commit:    "b79a5db77b5d6ffab9855234f8371d9e53978a24",
				ID:        "4fb4d506-cd63-4231-bce7-f0907cd418e2",
				Snapshot:  true,
				Version:   "8.6.0",
			},
		},
		"v2": {
			InFile: "version.txt",
			Want: Info{
				BuildTime: wantTime,
				Commit:    "b79a5db77b5d6ffab9855234f8371d9e53978a24",
				Snapshot:  true,
				Version:   "8.6.0",
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			testData, err := os.ReadFile(filepath.Join("testdata", tc.InFile))
			require.NoError(t, err)
			testReader := bytes.NewReader(testData)

			got, err := ParseInfo(testReader)
			if tc.WantErr != "" {
				require.ErrorContains(t, err, tc.WantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.Want, got)
			}

		})
	}
}
