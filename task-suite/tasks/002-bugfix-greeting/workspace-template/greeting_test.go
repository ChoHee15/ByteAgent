package greeting

import "testing"

func TestFormatGreeting(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "trims surrounding spaces",
			in:   "  Alice  ",
			want: "Hello, Alice!",
		},
		{
			name: "falls back for empty name",
			in:   "   ",
			want: "Hello, stranger!",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FormatGreeting(tc.in)
			if got != tc.want {
				t.Fatalf("FormatGreeting(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
