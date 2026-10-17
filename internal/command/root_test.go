package command

import "testing"

func TestVersionString(t *testing.T) {
	tests := []struct {
		name string
		app  App
		want string
	}{
		{
			name: "release build without commit",
			app:  App{Version: "v1.2.3", Commit: "none", Date: "unknown"},
			want: "v1.2.3",
		},
		{
			name: "release build with commit",
			app:  App{Version: "v1.2.3", Commit: "abc1234", Date: "2026-01-01T00:00:00Z"},
			want: "v1.2.3 (abc1234, 2026-01-01T00:00:00Z)",
		},
		{
			name: "development build",
			app:  App{Version: "dev", Commit: "", Date: ""},
			want: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.app.versionString(); got != tt.want {
				t.Fatalf("versionString() = %q, want %q", got, tt.want)
			}
		})
	}
}
