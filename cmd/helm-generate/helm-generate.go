package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"gopkg.in/yaml.v2"

	"github.com/topfreegames/helm-generate/pkg/helm"
	"github.com/topfreegames/helm-generate/pkg/util"

	"helm.sh/helm/v3/pkg/chartutil"
)

func getManifestsForPath(fullFilePath string, h *helm.Configuration) ([]map[string]interface{}, error) {
	path, filename := filepath.Split(fullFilePath)
	// Stores default configuration
	var manifests []map[string]interface{}
	if filename == h.ValuesYaml {
		helmYamlFullPath := path + h.HelmYaml
		file, _ := os.Open(helmYamlFullPath)
		defer file.Close()
		_ = h.BuildHelmConfig(file)
		vals, err := chartutil.ReadValuesFile(fullFilePath)
		if err != nil {
			return manifests, fmt.Errorf("Read Values: %v", err)
		}
		for k, v := range h.KeyValueAssignments {
			vals[k] = v
		}
		manifests, err = h.InstallChart(vals)
		if err != nil {
			return manifests, fmt.Errorf("Error generating manifests for chart %v: %v", h.Chart, err)
		}
	}
	return manifests, nil
}

func yamlEncoder(encoder *yaml.Encoder) func(element map[string]interface{}) {
	return func(value map[string]interface{}) {
		//nolint:errcheck
		encoder.Encode(value)
	}
}

func helmGenerate(cmd *cobra.Command, args []string) (bytes.Buffer, error) {
	var manifests []map[string]interface{}
	var rootPath string
	if len(args) > 0 {
		rootPath = args[0]
	} else {
		rootPath = "."
	}

	var keyValueAssignmentMap map[string]string
	if flag := cmd.Flag(flagSetKeyValue); flag != nil {
		if sliceValue, ok := flag.Value.(pflag.SliceValue); ok && sliceValue != nil {
			m, err := parseKeyValueAssignments(sliceValue.GetSlice())
			if err != nil {
				return bytes.Buffer{}, fmt.Errorf("error parsing key-value assignments: %w", err)
			}
			keyValueAssignmentMap = m
		}
	}

	err := filepath.Walk(rootPath,
		func(fullFilePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			config := &helm.Configuration{
				Chart:               cmd.Flag(flagDefaultChart).Value.String(),
				ChartVersion:        cmd.Flag(flagDefaultChartVersion).Value.String(),
				HelmYaml:            cmd.Flag(flagHelmYamlFilename).Value.String(),
				ValuesYaml:          cmd.Flag(flagHelmValuesFilename).Value.String(),
				PostRenderBinary:    cmd.Flag(flagPostRenderBinary).Value.String(),
				KeyValueAssignments: keyValueAssignmentMap,
			}
			chartManifests, err := getManifestsForPath(fullFilePath, config)
			if err != nil {
				return err
			}
			manifests = append(manifests, chartManifests...)
			return nil
		})
	if err != nil {
		return bytes.Buffer{}, err
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	util.WalkDedup(manifests, yamlEncoder(enc))

	return buf, nil
}

func parseKeyValueAssignments(keyValueAssignments []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, keyValue := range keyValueAssignments {
		s := strings.Split(keyValue, "=")
		if len(s) != 2 {
			return nil, fmt.Errorf("key-value assignment string is not of the form <key>=<value>: %s", keyValue)
		}
		key := s[0]
		value := s[1]
		if key == "" {
			return nil, errors.New("key-value assignment string cannot have empty key")
		}
		m[key] = value
	}
	return m, nil
}
