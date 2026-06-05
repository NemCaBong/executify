package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ParseSampleValues parses the multi-line sample text for ONE kind (all input
// fields, or all output fields) into a value per field, keyed by line_index.
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
//   - primitive data types     -> the token parsed as that type
//   - any array (e.g. int[][], string[]) -> JSON-decoded ("[[1,2],[3,4]]", `["a","b]c"]`)
func ParseLineValue(raw, dataType string) interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	baseDataType, dimensions := splitDataType(dataType)

	switch {
	case dimensions == 0:
		return parsePrimitiveData(raw, baseDataType)
	default:
		// format as json
		var v interface{}
		if err := json.Unmarshal([]byte(raw), &v); err == nil {
			return v
		}
		// no fallback plan
		return raw
	}
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

// parsePrimitiveData parses a single token according to its base type.
func parsePrimitiveData(value, base string) interface{} {
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
	case "string", "str", "char":
		return value
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

// CompareOutput reports whether a program's actual output matches the expected
// output for ONE test case. When the problem declares output fields, both texts
// are parsed field-by-field (driven by each field's data_type) and compared
// structurally. Numeric values are compared with a floatPrecision number
func CompareOutput(outputFields []ProblemIOField, expected, actual string, floatPrecision *int) bool {
	if len(outputFields) == 0 {
		return strings.TrimSpace(expected) == strings.TrimSpace(actual)
	}
	exp := ParseSampleValues(outputFields, expected)
	act := ParseSampleValues(outputFields, actual)
	for _, f := range outputFields {
		if !compareValue(exp[f.LineIndex], act[f.LineIndex], floatPrecision) {
			return false
		}
	}
	return true
}

// compareValue deep-compares two parsed values (scalars, or nested
// []interface{} arrays). Numbers are compared as float64 so cross-type pairs
// (int64 from a scalar parse vs float64 from JSON decoding) still match,
// with floatPrecision controlling the tolerance.
func compareValue(a, b interface{}, floatPrecision *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	aArr, aIsArr := a.([]interface{})
	bArr, bIsArr := b.([]interface{})
	// compare arr
	if aIsArr || bIsArr {
		if !aIsArr || !bIsArr || len(aArr) != len(bArr) {
			return false
		}
		for i := range aArr {
			if !compareValue(aArr[i], bArr[i], floatPrecision) {
				return false
			}
		}
		return true
	}
	// compare int, float
	if af, aok := toFloat(a); aok {
		if bf, bok := toFloat(b); bok {
			if floatPrecision != nil {
				return math.Abs(af-bf) <= math.Pow(10, -float64(*floatPrecision))
			}
			return af == bf
		}
		return false
	}
	// compare string
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}
