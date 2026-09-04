package file

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		contentType string
		size        int64
		wantErr     error
	}{
		{"valid conf", "zscaler.conf", "text/plain", 100, nil},
		{"valid txt", "config.txt", "", 100, nil},
		{"valid json", "rules.json", "application/json", 100, nil},
		{"valid xml octetstream", "rules.xml", "application/octet-stream", 100, nil},
		{"empty file", "empty.conf", "text/plain", 0, ErrEmpty},
		{"bad extension", "malware.exe", "application/octet-stream", 100, ErrUnsupported},
		{"bad content type", "data.conf", "image/png", 100, ErrUnsupported},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(tc.filename, tc.contentType, tc.size)
			if err != tc.wantErr {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	cases := map[string]string{
		"simple.conf":     "simple.conf",
		"../etc/passwd":   "passwd",
		"":                "upload",
		"/":               "upload",
		".":               "upload",
	}
	for in, want := range cases {
		got := sanitizeFilename(in)
		if got != want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestShardedKey(t *testing.T) {
	if k := ShardedKey("9a3c1234"); k != "9a/9a3c1234" {
		t.Fatalf("unexpected key: %s", k)
	}
}
