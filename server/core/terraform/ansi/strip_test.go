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
			str:  "\n\x1b[32m+\x1b[0m create\n\x1b[0m\x1b[1mPlan:\x1b[0m 3 to add, 0 to change, 0 to destroy.\n",
			want: "\n+ create\nPlan: 3 to add, 0 to change, 0 to destroy.\n",
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
