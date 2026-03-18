package swd

import (
	"reflect"
	"testing"
)

var (
	testWords = []SensitiveWord{
		{Word: "笨猪", Type: "default"},
		{Word: "笨蛋", Type: "default"},
		{Word: "傻瓜", Type: "default"},
	}
	swdInstance = NewSwd(testWords)
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *SensitiveWordCheckResult
		wantErr bool
	}{
		{
			name:  "test1",
			input: "你这个笨猪",
			want: &SensitiveWordCheckResult{
				Contains:     true,
				FilteredText: "你这个**",
				Matches: []SensitiveWordMatch{
					{Word: "笨猪", Category: "default", StartPos: 3, EndPos: 5},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := swdInstance.Check(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Swd.Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Swd.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkCheck(b *testing.B) {
	input := "你这个笨猪"
	for i := 0; i < b.N; i++ {
		swdInstance.Check(input)
	}
}
