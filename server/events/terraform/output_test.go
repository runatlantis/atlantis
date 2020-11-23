package terraform

import (
	"reflect"
	"testing"
)

func TestFileOutput_List(t *testing.T) {
}

func TestNewOutput(t *testing.T) {
	type args struct {
		outputCmdDir string
	}
	tests := []struct {
		name string
		args args
		want Output
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewOutput(tt.args.outputCmdDir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
