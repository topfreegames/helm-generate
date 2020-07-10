package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	Expected interface{}
	Sample   interface{}
	Name     string
}

type ReturnWithError struct {
	Value interface{}
	Error bool
}

func TestCreateNamespace(t *testing.T) {
	nsDefaultLabel := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[interface{}]interface{}{
			"annotations": map[interface{}]interface{}{"fluxcd.io/ignore": "sync_only"},
			"labels":      map[interface{}]interface{}{"name": "nsDefaultLabel"},
			"name":        "nsDefaultLabel",
		},
	}
	nsMultipleLabels := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[interface{}]interface{}{
			"annotations": map[interface{}]interface{}{"fluxcd.io/ignore": "sync_only"},
			"labels":      map[interface{}]interface{}{"name": "nsMultipleLabels", "anotherLabel": "coolValue"},
			"name":        "nsMultipleLabels",
		},
	}
	nsMultipleAnnotations := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[interface{}]interface{}{
			"annotations": map[interface{}]interface{}{"fluxcd.io/ignore": "sync_only", "dopeannotation": "nah"},
			"labels":      map[interface{}]interface{}{"name": "nsMultipleAnnotations"},
			"name":        "nsMultipleAnnotations",
		},
	}
	tests := []TestCase{
		{
			Sample: map[string]interface{}{
				"ns":          "nsDefaultLabel",
				"annotations": nil,
				"labels":      nil,
			},
			Expected: nsDefaultLabel,
			Name:     "Namespace with default label",
		},
		{
			Sample: map[string]interface{}{
				"ns":          "nsMultipleLabels",
				"annotations": nil,
				"labels":      map[string]string{"anotherLabel": "coolValue"},
			},
			Expected: nsMultipleLabels,
			Name:     "Namespace with additional labels",
		},
		{
			Sample: map[string]interface{}{
				"ns":          "nsMultipleAnnotations",
				"annotations": map[string]string{"dopeannotation": "nah"},
				"labels":      nil,
			},
			Expected: nsMultipleAnnotations,
			Name:     "Namespace with default Label",
		},
	}
	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)
		ns := test.Sample.(map[string]interface{})["ns"].(string)
		var annotations, labels map[string]string
		if test.Sample.(map[string]interface{})["annotations"] != nil {
			annotations = test.Sample.(map[string]interface{})["annotations"].(map[string]string)
		} else {
			annotations = nil
		}
		if test.Sample.(map[string]interface{})["labels"] != nil {
			labels = test.Sample.(map[string]interface{})["labels"].(map[string]string)
		} else {
			labels = nil
		}
		result := CreateNamespace(ns, annotations, labels)
		t.Logf("%v", result)
		assert.Equal(t, test.Expected, result)
	}
}

func TestWalkDedup(t *testing.T) {
	a := map[string]interface{}{
		"hello":  "world",
		"number": 42,
		"complex": map[string]string{
			"deep": "structure",
		},
	}
	b := map[string]interface{}{
		"something": "else",
	}
	var empty []map[string]interface{}

	tests := []TestCase{
		{
			Expected: append(empty, a, b),
			Sample:   append(empty, a, a, a, b),
			Name:     "duplicated A",
		},
		{
			Expected: append(empty, a, b),
			Sample:   append(empty, a, a, a, b, b, b),
			Name:     "duplicated As and Bs",
		},
		{
			Expected: append(empty, a, b),
			Sample:   append(empty, a, b, a),
			Name:     "non-sequential duplicated As",
		},
		{
			Expected: append(empty, b, a),
			Sample:   append(empty, b, a, b, a),
			Name:     "non-sequential duplicated As and Bs",
		},
	}

	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)
		var actual []map[string]interface{}
		calledTimes := 0
		WalkDedup(test.Sample.([]map[string]interface{}), func(elem map[string]interface{}) {
			calledTimes = calledTimes + 1
			actual = append(actual, elem)
		})

		assert.Equal(t, len(test.Expected.([]map[string]interface{})), calledTimes, "walk function should be called only once for each unique element")
		assert.Equal(t, test.Expected, actual, "should have deduped list")
	}
}

func TestNestedMapLookup(t *testing.T) {
	m := map[string]interface{}{
		"root": map[string]interface{}{
			"internal": map[string]interface{}{
				"int":     42,
				"string":  "string",
				"float64": 6.022e23,
				"malformed": map[int]int{
					1: 1,
				},
			},
		},
	}

	tests := []TestCase{
		{
			Name: "existing integer value",
			Expected: ReturnWithError{
				Value: m["root"].(map[string]interface{})["internal"].(map[string]interface{})["int"],
				Error: false,
			},
			Sample: []string{"root", "internal", "int"},
		},
		{
			Name: "existing string value",
			Expected: ReturnWithError{
				Value: m["root"].(map[string]interface{})["internal"].(map[string]interface{})["string"],
				Error: false,
			},
			Sample: []string{"root", "internal", "string"},
		},
		{
			Name: "existing float64 value",
			Expected: ReturnWithError{
				Value: m["root"].(map[string]interface{})["internal"].(map[string]interface{})["float64"],
				Error: false,
			},
			Sample: []string{"root", "internal", "float64"},
		},
		{
			Name: "existing complex value",
			Expected: ReturnWithError{
				Value: m["root"].(map[string]interface{})["internal"],
				Error: false,
			},
			Sample: []string{"root", "internal"},
		},
		{
			Name: "non-existent value",
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
			Sample: []string{"root", "infernal"},
		},
		{
			Name: "no query",
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
			Sample: []string{},
		},
		{
			Name: "malformed structure",
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
			Sample: []string{"root", "internal", "malformed", "anything"},
		},
	}

	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)

		expected := test.Expected.(ReturnWithError)
		sample := test.Sample.([]string)

		val, err := NestedMapLookup(m, sample...)

		if expected.Error {
			assert.Error(t, err, "should return an error")
		} else {
			assert.Nil(t, err, "should not return error")
		}
		assert.Equal(t, expected.Value, val, "value should match struct value")
	}
}

func TestValidateValues(t *testing.T) {
	m := map[string]interface{}{
		"first": map[string]interface{}{
			"internal": map[string]interface{}{
				"int":     42,
				"string":  "string",
				"float64": 6.022e23,
			},
		},
		"second": 42,
		"third":  "hello",
	}

	tests := []TestCase{
		{
			Name: "existing required field",
			Expected: ReturnWithError{
				Value: map[string]interface{}{
					"first": m["first"],
				},
				Error: false,
			},
			Sample: []string{"first"},
		},
		{
			Name: "multiple existing required fields",
			Expected: ReturnWithError{
				Value: map[string]interface{}{
					"first": m["first"], "second": m["second"], "third": m["third"],
				},
				Error: false,
			},
			Sample: []string{"first", "second", "third"},
		},
		{
			Name: "non-existing required field",
			Expected: ReturnWithError{
				Value: map[string]interface{}{},
				Error: true,
			},
			Sample: []string{"fourth"},
		},
		{
			Name: "non-existing required field and existing required",
			Expected: ReturnWithError{
				Value: map[string]interface{}{
					"first": m["first"],
				},
				Error: true,
			},
			Sample: []string{"first", "fourth"},
		},
		{
			Name: "non-existing required field and existing required out of order",
			Expected: ReturnWithError{
				Value: map[string]interface{}{
					"first": m["first"], "second": m["second"],
				},
				Error: true,
			},
			Sample: []string{"first", "fifth", "fourth", "second"},
		},
		{
			Name: "no required field",
			Expected: ReturnWithError{
				Value: make(map[string]interface{}),
				Error: false,
			},
			Sample: []string{},
		},
	}

	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)

		expected := test.Expected.(ReturnWithError)
		sample := test.Sample.([]string)

		val, err := ValidateValues(m, sample...)

		if expected.Error {
			assert.Error(t, err, "should return an error")
		} else {
			assert.Nil(t, err, "should not return error")
		}
		assert.Equal(t, expected.Value.(map[string]interface{}), val, "value should match struct value")
	}
}

func TestDecodeYamls(t *testing.T) {
	tests := []TestCase{
		{
			Name: "successful decoding",
			Sample: `
---
number: 42
string: string
deep:
  number: 42
---
number: 43
string: otherstring
deep:
  number: 43`,
			Expected: ReturnWithError{
				Value: []map[string]interface{}{
					{
						"number": 42,
						"string": "string",
						"deep": map[interface{}]interface{}{
							"number": 42,
						},
					},
					{
						"number": 43,
						"string": "otherstring",
						"deep": map[interface{}]interface{}{
							"number": 43,
						},
					},
				},
				Error: false,
			},
		},
		{
			Name:   "failed decode",
			Sample: `notAYaml`,
			Expected: ReturnWithError{
				Error: true,
			},
		},
	}

	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)

		expected := test.Expected.(ReturnWithError)
		sample := test.Sample.(string)

		val, err := DecodeYamls(sample)

		if expected.Error {
			assert.Error(t, err, "should return an error")
			assert.Nil(t, val, "return should be nil on error")
		} else {
			assert.Nil(t, err, "should not return error")
			assert.Equal(t, expected.Value.([]map[string]interface{}), val, "decoded value should match expected")
		}
	}
}
