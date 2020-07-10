package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"gopkg.in/yaml.v2"

	"git.topfreegames.com/sre/helm-generate/pkg/helm"
	"git.topfreegames.com/sre/helm-generate/pkg/util"

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

	err := filepath.Walk(rootPath,
		func(fullFilePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			config := &helm.Configuration{
				Chart:            cmd.Flag(flagDefaultChart).Value.String(),
				ChartVersion:     cmd.Flag(flagDefaultChartVersion).Value.String(),
				HelmYaml:         cmd.Flag(flagHelmYamlFilename).Value.String(),
				ValuesYaml:       cmd.Flag(flagHelmValuesFilename).Value.String(),
				PostRenderBinary: cmd.Flag(flagPostRenderBinary).Value.String(),
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
