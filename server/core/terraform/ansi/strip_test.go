package ansi

import "testing"

func TestStrip(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "strip ansi",
			str: `
[32m+[0m create
[0m[1mPlan:[0m 3 to add, 0 to change, 0 to destroy.
`,
			want: `
+ create
Plan: 3 to add, 0 to change, 0 to destroy.
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Strip(tt.str); got != tt.want {
				t.Errorf("Strip() = %v, want %v", got, tt.want)
			}
		})
	}
}
