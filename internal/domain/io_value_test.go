package domain

import (
	"encoding/json"
	"testing"
)

func jsonOf(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestParseSampleValues(t *testing.T) {
	tests := []struct {
		name   string
		fields []ProblemIOField
		text   string
		want   map[int]string // line_index -> JSON of value
	}{
		{
			name: "space-separated 1D array is no longer coerced (JSON required)",
			fields: []ProblemIOField{
				{Kind: IOKindInput, LineIndex: 0, KeyName: "nums", DataType: "int[]"},
			},
			text: "9 2 7 11 15",
			want: map[int]string{0: `"9 2 7 11 15"`},
		},
		{
			name: "JSON 1D int array",
			fields: []ProblemIOField{
				{Kind: IOKindInput, LineIndex: 0, KeyName: "nums", DataType: "int[]"},
			},
			text: "[2,7,11,15]",
			want: map[int]string{0: `[2,7,11,15]`},
		},
		{
			name: "scalar int and 1D array on separate lines",
			fields: []ProblemIOField{
				{Kind: IOKindInput, LineIndex: 0, KeyName: "target", DataType: "int"},
				{Kind: IOKindInput, LineIndex: 1, KeyName: "nums", DataType: "int[]"},
			},
			text: "6\n[2,7,11,15]",
			want: map[int]string{0: `6`, 1: `[2,7,11,15]`},
		},
		{
			name: "int[][] on a single JSON line (variable size, line layout unaffected)",
			fields: []ProblemIOField{
				{Kind: IOKindInput, LineIndex: 0, KeyName: "matrix", DataType: "int[][]"},
				{Kind: IOKindInput, LineIndex: 1, KeyName: "target", DataType: "int"},
			},
			text: "[[1,2,3],[4,5,6]]\n7",
			want: map[int]string{0: `[[1,2,3],[4,5,6]]`, 1: `7`},
		},
		{
			name: "string scalar containing brackets stays literal",
			fields: []ProblemIOField{
				{Kind: IOKindInput, LineIndex: 0, KeyName: "s", DataType: "string"},
			},
			text: "a[b]c",
			want: map[int]string{0: `"a[b]c"`},
		},
		{
			name: "string scalar that looks like JSON is NOT decoded",
			fields: []ProblemIOField{
				{Kind: IOKindInput, LineIndex: 0, KeyName: "s", DataType: "string"},
			},
			text: "[1,2]",
			want: map[int]string{0: `"[1,2]"`},
		},
		{
			name: "string scalar keeps spaces",
			fields: []ProblemIOField{
				{Kind: IOKindInput, LineIndex: 0, KeyName: "s", DataType: "string"},
			},
			text: "hello world",
			want: map[int]string{0: `"hello world"`},
		},
		{
			name: "string[] with brackets inside quoted elements",
			fields: []ProblemIOField{
				{Kind: IOKindInput, LineIndex: 0, KeyName: "words", DataType: "string[]"},
			},
			text: `["a","b[c]","d]e"]`,
			want: map[int]string{0: `["a","b[c]","d]e"]`},
		},
		{
			name: "missing sample line yields null",
			fields: []ProblemIOField{
				{Kind: IOKindInput, LineIndex: 0, KeyName: "a", DataType: "int"},
				{Kind: IOKindInput, LineIndex: 1, KeyName: "b", DataType: "int"},
			},
			text: "5",
			want: map[int]string{0: `5`, 1: `null`},
		},
		{
			name: "output 1D indices (JSON)",
			fields: []ProblemIOField{
				{Kind: IOKindOutput, LineIndex: 0, KeyName: "result", DataType: "int[]"},
			},
			text: "[0,1]",
			want: map[int]string{0: `[0,1]`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSampleValues(tt.fields, tt.text)
			for idx, wantJSON := range tt.want {
				if gotJSON := jsonOf(got[idx]); gotJSON != wantJSON {
					t.Errorf("line %d: got %s, want %s", idx, gotJSON, wantJSON)
				}
			}
		})
	}
}

func intPtr(n int) *int { return &n }

func TestCompareOutput(t *testing.T) {
	intArr := []ProblemIOField{{Kind: IOKindOutput, LineIndex: 0, KeyName: "result", DataType: "int[]"}}
	floatScalar := []ProblemIOField{{Kind: IOKindOutput, LineIndex: 0, KeyName: "ans", DataType: "float"}}
	floatArr := []ProblemIOField{{Kind: IOKindOutput, LineIndex: 0, KeyName: "ans", DataType: "float[]"}}

	tests := []struct {
		name       string
		fields     []ProblemIOField
		expected   string
		actual     string
		precision  *int
		wantEquals bool
	}{
		{"int array exact match", intArr, "[0,1]", "[0,1]", nil, true},
		{"int array space-separated no longer matches JSON", intArr, "[0,1]", "0 1", nil, false},
		{"int array order matters", intArr, "[0,1]", "[1,0]", nil, false},
		{"int array length differs", intArr, "[0,1]", "[0,1,2]", nil, false},
		{"float within precision", floatScalar, "3.141592", "3.141593", intPtr(5), true},
		{"float outside precision", floatScalar, "3.14159", "3.14260", intPtr(5), false},
		{"float exact required when no precision", floatScalar, "3.14", "3.140001", nil, false},
		{"float array within precision", floatArr, "[1.0,2.0]", "[1.000001,1.999999]", intPtr(4), true},
		{"no schema falls back to trimmed string", nil, "0 1\n", "0 1", nil, true},
		{"no schema string mismatch", nil, "0 1", "1 0", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareOutput(tt.fields, tt.expected, tt.actual, tt.precision); got != tt.wantEquals {
				t.Errorf("CompareOutput(%q, %q) = %v, want %v", tt.expected, tt.actual, got, tt.wantEquals)
			}
		})
	}
}

func TestParseLineValue(t *testing.T) {
	tests := []struct {
		raw      string
		dataType string
		want     string
	}{
		{"5", "int", `5`},
		{"3.14", "float", `3.14`},
		{"true", "bool", `true`},
		{"hello", "string", `"hello"`},
		{"1 2 3", "int[]", `"1 2 3"`}, // space-separated no longer coerced; stays raw
		{"[1,2,3]", "int[]", `[1,2,3]`},
		{"[[1,2],[3,4]]", "int[][]", `[[1,2],[3,4]]`},
		{"", "int", `null`},
		{"[1,2]", "string", `"[1,2]"`}, // string never decoded as array
	}
	for _, tt := range tests {
		if got := jsonOf(ParseLineValue(tt.raw, tt.dataType)); got != tt.want {
			t.Errorf("ParseLineValue(%q, %q) = %s, want %s", tt.raw, tt.dataType, got, tt.want)
		}
	}
}
