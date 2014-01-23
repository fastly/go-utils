package common_test

import (
	"github.com/fastly/go-utils/common"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestEscapedLines(t *testing.T) {
	tests := []struct {
		Input  []string
		Expect []string
	}{
		{[]string{}, []string{}},
		{[]string{"One\nTwo\nThree", "Four\nTwo\nThree", "Five\nOne\nTwo"},
			[]string{"Two"}},
	}

	for _, test := range tests {
		got := common.EscapedLines(test.Input)
		if !reflect.DeepEqual(test.Expect, got) {
			t.Errorf("return slice != expect, return: %v, expect: %v", got, test.Expect)
		}
	}
}

func BenchmarkEscapedLines(b *testing.B) {
	b.StopTimer()
	bytes, err := ioutil.ReadFile("commonlines_test_file.txt")
	if err != nil {
		b.Fatalf("unable to read commonlines file, err: %v", err)
	}

	// json pre-escaped the newlines, unescape them
	input := strings.Split(string(bytes), "\n")
	for i := range input {
		input[i] = strings.Replace(input[i], "\\n", "\n", -1)
	}
	output := common.EscapedLines(input)
	b.Logf("output of commonlines file: %v, len: %v", output, len(output))

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		common.EscapedLines(input)
	}
}
