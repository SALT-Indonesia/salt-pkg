package common_test

import (
	"examples/logmanager/pkg/common"
	"fmt"
	"testing"
	"time"
)

func TestDefaultValueEnv(t *testing.T) {
	type args struct {
		key        string
		defaultVal string
	}

	fakeWord := fmt.Sprintf("%v", time.Now().Unix())

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "OK",
			args: args{
				key:        "TEST",
				defaultVal: fakeWord + "1",
			},
			want: fakeWord + "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.DefaultValueEnv(tt.args.key, tt.args.defaultVal); got != tt.want {
				t.Errorf("DefaultValueEnv1() = %v, want %v", got, tt.want)
			}
		})
	}
}
