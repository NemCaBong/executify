package domain

import (
	"encoding/json"
	"strconv"
	"strings"
)

// ParseSampleValues parses the multi-line sample text for ONE kind (all input
// fields, or all output fields) into a value per field, keyed by line_index.
//
// Each field occupies exactly one line: the line at its line_index. The value
// is parsed according to data_type (see ParseLineValue). Nested types
// (int[][], string[], ...) are JSON-encoded on that single line, so a field's
// size can vary freely without affecting the line layout of other fields.
//
// A field whose line_index is past the end of the sample text yields nil
// (sample data may not cover every declared field).
func ParseSampleValues(fields []ProblemIOField, text string) map[int]interface{} {
	result := make(map[int]interface{}, len(fields))
	lines := splitTextLines(text)
	for _, f := range fields {
		var raw string
		if f.LineIndex >= 0 && f.LineIndex < len(lines) {
			raw = lines[f.LineIndex]
		}
		result[f.LineIndex] = ParseLineValue(raw, f.DataType)
	}
	return result
}

// ParseLineValue turns one raw line into a typed value, driven by data_type so
// the data never gets to decide its own shape:
//
//   - string scalar            -> the line taken literally
//   - int/float/bool scalar    -> the token parsed as that type
//   - any array (e.g. int[][], string[]) -> JSON-decoded ("[[1,2],[3,4]]", `["a","b]c"]`)
//   - numeric 1D array         -> also accepts whitespace-separated form ("9 2 7 11 15") as a fallback
func ParseLineValue(raw, dataType string) interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	baseDataType, dimensions := splitDataType(dataType)

	switch {
	case dimensions == 0 && isStringBase(baseDataType):
		// String scalar
		return raw
	case dimensions == 0:
		return parseScalar(raw, baseDataType)
	default:
		// Arrays: JSON is the reliable, unambiguous encoding.
		var v interface{}
		if err := json.Unmarshal([]byte(raw), &v); err == nil {
			return v
		}
		// Fallback for the space-separated numeric 1D files.
		if dimensions == 1 && !isStringBase(baseDataType) {
			return parseLine1D(raw, baseDataType)
		}
		return raw
	}
}

// parseLine1D splits a single line on whitespace and parses each token.
func parseLine1D(line, base string) []interface{} {
	tokens := strings.Fields(line)
	out := make([]interface{}, len(tokens))
	for i, t := range tokens {
		out[i] = parseScalar(t, base)
	}
	return out
}

// splitDataType peels trailing "[]" suffixes off a data_type, returning the
// scalar base type and the array dimension count ("int[][]" -> "int", 2).
func splitDataType(dataType string) (base string, dims int) {
	base = strings.TrimSpace(dataType)
	for strings.HasSuffix(base, "[]") {
		dims++
		base = strings.TrimSpace(strings.TrimSuffix(base, "[]"))
	}
	return base, dims
}

func isStringBase(base string) bool {
	switch strings.ToLower(base) {
	case "string", "str", "char":
		return true
	}
	return false
}

// parseScalar parses a single token according to its base type, falling back
// to the raw string for string/char types or any failed numeric parse.
func parseScalar(value, base string) interface{} {
	switch strings.ToLower(base) {
	case "int", "int32", "int64", "long":
		if n, err := strconv.ParseInt(value, 10, 64); err == nil {
			return n
		}
	case "float", "double", "float64":
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	case "bool", "boolean":
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return value
}

// splitTextLines normalizes line endings and drops a single trailing newline so
// a terminating "\n" does not produce a spurious empty final line.
func splitTextLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
