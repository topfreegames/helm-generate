package helm

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/postrender"
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

func TestBuildHelmConfig(t *testing.T) {
	chartName := "example"
	chartVersion := "3.14"
	nonExistentFile, _ := os.Open("i don't exist")
	defer nonExistentFile.Close()
	tests := []TestCase{
		{
			Name: "sucessful build with .helm.yaml",
			Sample: map[string]interface{}{
				"reader": strings.NewReader(fmt.Sprintf(`
chart: %s
chartVersion: %s
`, chartName, chartVersion)),
				"helmConf": Configuration{},
			},
			Expected: ReturnWithError{
				Value: Configuration{
					Chart:        chartName,
					ChartVersion: chartVersion,
				},
				Error: false,
			},
		},
		{
			Name: "error loading .helm.yaml config file",
			Sample: map[string]interface{}{
				"reader":   nonExistentFile,
				"helmConf": Configuration{},
			},
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
		},
		{
			Name: "error unmarshaling .helm.yaml",
			Sample: map[string]interface{}{
				"reader": strings.NewReader(`
kelly: [key:
wrong
`),
				"helmConf": Configuration{},
			},
			Expected: ReturnWithError{
				Value: Configuration{},
				Error: true,
			},
		},
		{
			Name: "non existent .helm.yaml file and valid defaults",
			Sample: map[string]interface{}{
				"reader": nil,
				"helmConf": Configuration{
					Chart:        "example",
					ChartVersion: "v0.1.2",
				},
			},
			Expected: ReturnWithError{
				Value: Configuration{
					Chart:        "example",
					ChartVersion: "v0.1.2",
				},
				Error: false,
			},
		},
		{
			Name: "empty .helm.yaml file contents and empty defaults",
			Sample: map[string]interface{}{
				"reader":   nil,
				"helmConf": Configuration{},
			},
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
		},
		{
			Name: "empty .helm.yaml file contents and partial defaults",
			Sample: map[string]interface{}{
				"reader": nil,
				"helmConf": Configuration{
					Chart: chartName,
				},
			},
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
		},
		{
			Name: "override all default values with file contents",
			Sample: map[string]interface{}{
				"reader": strings.NewReader(`
chartVersion: 18.0.0-rc2-slim
chart: not an example
`),
				"helmConf": Configuration{
					Chart:        chartName,
					ChartVersion: chartVersion,
				},
			},
			Expected: ReturnWithError{
				Value: Configuration{
					Chart:        "not an example",
					ChartVersion: "18.0.0-rc2-slim",
				},
				Error: false,
			},
		},
		{
			Name: "partial override default values with file contents",
			Sample: map[string]interface{}{
				"reader": strings.NewReader(`
chartVersion: 18.0.0-rc2-slim
`),
				"helmConf": Configuration{
					Chart:        chartName,
					ChartVersion: chartVersion,
				},
			},
			Expected: ReturnWithError{
				Value: Configuration{
					Chart:        chartName,
					ChartVersion: "18.0.0-rc2-slim",
				},
				Error: false,
			},
		},
	}
	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)

		expected := test.Expected.(ReturnWithError)
		sample := test.Sample.(map[string]interface{})

		reader := sample["reader"]
		h := sample["helmConf"].(Configuration)
		var err error
		if reader != nil {
			err = h.BuildHelmConfig(reader.(io.Reader))
		} else {
			err = h.BuildHelmConfig(nil)
		}

		if expected.Error {
			assert.Error(t, err, "should return an error")
		} else {
			assert.Nil(t, err, "should not return error")
			assert.Equal(t, expected.Value.(Configuration), h, "helm configuration should match the expected value")
		}
	}
}
func TestBuildHelmClient(t *testing.T) {
	settings := cli.New()
	actionConfig := new(action.Configuration)
	//nolint:errcheck
	actionConfig.Init(settings.RESTClientGetter(), "ns", os.Getenv("HELM_DRIVER"), log.Printf)
	mockClient := action.NewInstall(actionConfig)
	mockClient.ReleaseName = "releaseName"
	mockClient.Namespace = "ns"
	mockClient.DryRun = true
	mockClient.ClientOnly = true
	mockClient.UseReleaseName = true
	mockClient.PostRenderer, _ = postrender.NewExec("ls")
	tests := []TestCase{
		{
			Name: "sucessful helm client",
			Sample: map[string]interface{}{
				"helmConf": Configuration{
					PostRenderBinary: "ls",
				},
				"name":      "releaseName",
				"namespace": "ns",
			},
			Expected: ReturnWithError{
				Value: mockClient,
				Error: false,
			},
		},
		{
			Name: "non-existent postrenderer binary",
			Sample: map[string]interface{}{
				"helmConf": Configuration{
					PostRenderBinary: "non-existent-binary",
				},
				"name":      "releaseName",
				"namespace": "ns",
			},
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
		},
	}
	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)

		expected := test.Expected.(ReturnWithError)
		sample := test.Sample.(map[string]interface{})

		name := sample["name"].(string)
		namespace := sample["namespace"].(string)
		h := sample["helmConf"].(Configuration)
		client, err := h.buildHelmClient(name, namespace)
		if expected.Error {
			assert.Error(t, err, "should return an error")
		} else {
			assert.Nil(t, err, "should not return error")
			expectedClient := expected.Value.(*action.Install)
			assert.Equal(t, expectedClient.Namespace, client.Namespace, "Helm client must have matching namespace")
			assert.Equal(t, expectedClient.ReleaseName, client.ReleaseName, "Helm client must have matching release name")
		}
	}
}

func TestInstallChart(t *testing.T) {
	tests := []TestCase{
		{
			Name: "missing required fields",
			Sample: map[string]interface{}{
				"values": chartutil.Values{
					"key1": "value1",
					"key2": "value2",
				},
				"helmConf": Configuration{},
			},
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
		},
		{
			Name: "missing values configuration",
			Sample: map[string]interface{}{
				"values":   nil,
				"helmConf": Configuration{},
			},
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
		},
		{
			Name: "non-existent postrenderer binary",
			Sample: map[string]interface{}{
				"values": chartutil.Values{
					"releaseName": "releaseTest",
					"namespace":   "namespaceTest",
				},
				"helmConf": Configuration{
					PostRenderBinary: "non-existent-binary",
				},
			},
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
		},
		{
			Name: "invalid chart",
			Sample: map[string]interface{}{
				"values": chartutil.Values{
					"releaseName": "releaseTest",
					"namespace":   "namespaceTest",
				},
				"helmConf": Configuration{
					Chart:        "invalid-chart",
					ChartVersion: "x",
				},
			},
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
		},
	}
	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)

		expected := test.Expected.(ReturnWithError)
		sample := test.Sample.(map[string]interface{})

		values := sample["values"]
		h := sample["helmConf"].(Configuration)
		var manifests []map[string]interface{}
		var err error
		if values != nil {
			manifests, err = h.InstallChart(values.(chartutil.Values))
		} else {
			manifests, err = h.InstallChart(nil)
		}

		if expected.Error {
			assert.Error(t, err, "should return an error")
		} else {
			assert.Nil(t, err, "should not return error")
			assert.Equal(t, expected.Value.([]map[string]interface{}), manifests, "generated manifests should match the expected value")
		}
	}
}

func TestAddNamespaceMetadata(t *testing.T) {
	emptyMap := map[string]interface{}{}
	namespace := "test-namespace"
	tests := []TestCase{
		{
			Name: "manifest with no namespace",
			Sample: []map[string]interface{}{
				{
					"metadata": map[interface{}]interface{}{
						"name": "test-manifest",
					},
				},
			},
			Expected: ReturnWithError{
				Value: []map[string]interface{}{
					{
						"metadata": map[interface{}]interface{}{
							"name":      "test-manifest",
							"namespace": namespace,
						},
					},
				},
				Error: false,
			},
		},
		{
			Name: "manifest with no metadata",
			Sample: []map[string]interface{}{
				{
					"not-metadata": "yes",
				},
			},
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
		},
		{
			Name:   "empty manifest",
			Sample: []map[string]interface{}{emptyMap},

			Expected: ReturnWithError{
				Value: []map[string]interface{}{emptyMap},
				Error: false,
			},
		},
		{
			Name: "manifest with different namespace",
			Sample: []map[string]interface{}{
				{
					"metadata": map[interface{}]interface{}{
						"name":      "test-manifest",
						"namespace": "another-namespace",
					},
				},
			},
			Expected: ReturnWithError{
				Value: []map[string]interface{}{
					{
						"metadata": map[interface{}]interface{}{
							"name":      "test-manifest",
							"namespace": namespace,
						},
					},
				},
				Error: false,
			},
		},
	}

	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)

		expected := test.Expected.(ReturnWithError)
		sample := test.Sample.([]map[string]interface{})

		val, err := addNamespaceMetadata(sample, namespace)

		if expected.Error {
			assert.Error(t, err, "should return an error")
			assert.Nil(t, val, "return should be nil on error")
		} else {
			assert.Nil(t, err, "should not return error")
			assert.Equal(t, expected.Value.([]map[string]interface{}), val, "decoded value should match expected")
		}
	}
}
func TestLoadChart(t *testing.T) {
	settings := cli.New()
	actionConfig := new(action.Configuration)
	//nolint:errcheck
	actionConfig.Init(settings.RESTClientGetter(), "ns", os.Getenv("HELM_DRIVER"), log.Printf)
	client := action.NewInstall(actionConfig)
	tests := []TestCase{
		{
			Name: "valid install",
			Sample: Configuration{
				Chart:        "tests/chart",
				ChartVersion: "0.1.0",
			},
			Expected: ReturnWithError{
				Value: chart.Metadata{
					Name:        "chart",
					Home:        "",
					Version:     "0.1.0",
					Description: "A Helm chart for Kubernetes",
					APIVersion:  "v2",
					AppVersion:  "1.16.0",
					Type:        "application",
				},

				Error: false,
			},
		},
		{
			Name: "non-existent chart",
			Sample: Configuration{
				Chart:        "invalid-chart",
				ChartVersion: "9",
			},
			Expected: ReturnWithError{
				Value: nil,
				Error: true,
			},
		},
	}
	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)
		sample := test.Sample.(Configuration)
		expected := test.Expected.(ReturnWithError)

		chart, err := sample.loadChart(client)

		if expected.Error {
			assert.Error(t, err, "should return an error")
			assert.Nil(t, chart, "return should be nil on error")
		} else {
			assert.Nil(t, err, "should not return error")
			assert.Equal(t, expected.Value, *chart.Metadata, "decoded value should match expected")
		}
	}
}
