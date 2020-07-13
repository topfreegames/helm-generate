package util

import (
	"fmt"
	"io"
	"strings"

	"github.com/mitchellh/hashstructure"
	"gopkg.in/yaml.v2"
)

// Metadata represents namespaces metadata information
type Metadata struct {
	Annotations map[string]string
	Name        string `yaml:"name"`
	Labels      map[string]string
}

// Namespace represents a Kubernetes namespace resource
type Namespace struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
}

// CreateNamespace generates a map with namespace definitions
func CreateNamespace(ns string, annotations map[string]string, labels map[string]string) map[string]interface{} {
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if labels == nil {
		labels = make(map[string]string)
	}
	if annotations["fluxcd.io/sync_only"] == "" {
		annotations["fluxcd.io/sync_only"] = "true"
	}
	labels["name"] = ns
	namespace, _ := yaml.Marshal(&Namespace{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata: Metadata{
			Annotations: annotations,
			Labels:      labels,
			Name:        ns,
		},
	})
	var namespaceMap map[string]interface{}
	//nolint:errcheck
	yaml.Unmarshal(namespace, &namespaceMap)
	return namespaceMap
}

// WalkDedup iterates over a list of maps and uses hashing to deduplicate elements
func WalkDedup(list []map[string]interface{}, f func(element map[string]interface{})) {
	hashList := make(map[uint64]bool)
	for _, elem := range list {
		// This always return nil as err parameter
		itemHash, _ := hashstructure.Hash(elem, nil)
		if _, ok := hashList[itemHash]; !ok {
			hashList[itemHash] = true
			f(elem)
		}
	}
}

// NestedMapLookup as found on https://gist.github.com/ChristopherThorpe/fd3720efe2ba83c929bf4105719ee967
// m:  a map from strings to other maps or values, of arbitrary depth
// ks: successive keys to reach an internal or leaf node (variadic)
// If an internal node is reached, will return the internal map
//
// Returns: (Exactly one of these will be nil)
// rval: the target node (if found)
// err:  an error created by fmt.Errorf
//
func NestedMapLookup(m map[string]interface{}, ks ...string) (rval interface{}, err error) {
	var ok bool
	if len(ks) == 0 || len(m) == 0 {
		return nil, fmt.Errorf("NestedMapLookup needs at least one key")
	}
	if rval, ok = m[ks[0]]; !ok {
		return nil, fmt.Errorf("key not found; remaining keys: %v", ks)
	} else if len(ks) == 1 { // we've reached the final key
		return rval, nil
	} else if m, ok = rval.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("malformed structure at %#v", rval)
	} else { // 1+ more keys
		return NestedMapLookup(m, ks[1:]...)
	}
}

// ValidateValues validates if some required keys exists on a map
func ValidateValues(values map[string]interface{}, requiredFields ...string) (map[string]interface{}, error) {
	requiredValues := make(map[string]interface{})
	failedFields := []string{}
	for _, field := range requiredFields {
		value, err := NestedMapLookup(values, field)
		if err != nil {
			failedFields = append(failedFields, field)
		} else {
			requiredValues[field] = value
		}
	}

	var err error = nil
	if len(failedFields) > 0 {
		err = fmt.Errorf("Missing required field %v", failedFields)
	}
	return requiredValues, err
}

// DecodeYamls parse a list of yamls defined on a string to a list of maps
func DecodeYamls(yamlString string) ([]map[string]interface{}, error) {
	dec := yaml.NewDecoder(strings.NewReader(yamlString))
	var manifests []map[string]interface{}
	for {
		var value map[string]interface{}
		err := dec.Decode(&value)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("Error decoding yaml: %v", err)
		} else if err == io.EOF {
			break
		}
		manifests = append(manifests, value)
	}
	return manifests, nil
}
