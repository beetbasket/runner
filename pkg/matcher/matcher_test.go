package matcher

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatcher(t *testing.T) {
	for _, tt := range []struct {
		name    string
		prefix  string
		input   []string
		output  []string
		special []string
	}{
		{
			name:    "one line no nl",
			prefix:  "aaa",
			input:   newA("foo bar baz"),
			output:  newA("foo bar baz"),
			special: newA(""),
		},
		{
			name:    "one line nl",
			prefix:  "aaa",
			input:   newA("foo bar baz\nnext line"),
			output:  newA("foo bar baz\nnext line"),
			special: newA(""),
		},
		{
			name:    "one line no nl broken",
			prefix:  "aaa",
			input:   newA("f", "o", "o bar", " baz"),
			output:  newA("f", "o", "o bar", " baz"),
			special: newA("", "", "", ""),
		},
		{
			name:    "one line nl broken",
			prefix:  "aaa",
			input:   newA("foo b", "ar ba", "z\nne", "x", "t line"),
			output:  newA("foo b", "ar ba", "z\nne", "x", "t line"),
			special: newA("", "", "", "", ""),
		},
		{
			name:    "with special",
			prefix:  "aaa",
			input:   newA("foo bar baz\naaa foo bar baz\n"),
			output:  newA("foo bar baz\n"),
			special: newA("foo bar baz\n"),
		},
		{
			name:    "with special broken",
			prefix:  "aaa",
			input:   newA("fo", "o bar ", "baz\naaa foo bar baz\n"),
			output:  newA("fo", "o bar ", "baz\n"),
			special: newA("", "", "foo bar baz\n"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.prefix)
			var out, spec []string
			for _, b := range tt.input {
				_, _ = m.Write([]byte(b))
				out = append(out, string(m.ReadOut()))
				spec = append(spec, string(m.ReadSpecial()))
			}

			assert.Equal(t, tt.output, out)
			assert.Equal(t, tt.special, spec)
		})
	}
}

func newA[B Bytes](d ...B) []string {
	var a []string
	for _, d := range d {
		a = append(a, string(d))
	}
	return a
}
