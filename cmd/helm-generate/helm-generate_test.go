package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	Expected interface{}
	Sample   interface{}
	Name     string
}

func TestInstallChartWithArgs(t *testing.T) {
	sampleDir := "tests/samples"
	expectedDir := "tests/expected"
	samples, err := ioutil.ReadDir(sampleDir)
	if err != nil {
		t.Logf("%v", err)
		t.Fail()
	}
	expected, err := ioutil.ReadDir(expectedDir)
	if err != nil {
		t.Logf("%v", err)
		t.Fail()
	}
	assert.Equal(t, len(samples), len(expected), "Length Test Samples don't match the Expected")
	for i := range samples {
		assert.Equal(t, samples[i].Name(), expected[i].Name(), "Sample folder don't match Expected folder")
	}
	var tests []TestCase
	for i, sample := range samples {
		test := TestCase{
			Name:     fmt.Sprintf("Parsing test folder %s/%s", sampleDir, sample.Name()),
			Sample:   fmt.Sprintf("%s/%s", sampleDir, sample.Name()),
			Expected: fmt.Sprintf("%s/%s", expectedDir, expected[i].Name()),
		}
		tests = append(tests, test)
	}

	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)
		var mockCmd = &cobra.Command{
			Use:   "helm-generate [root-path]",
			Short: "templates helm charts and prints it to stdout",
			Long:  ``,
			Args:  cobra.RangeArgs(0, 1),
			Run: func(cmd *cobra.Command, args []string) {
				//nolint:errcheck
				helmGenerate(cmd, args)
			},
		}
		mockCmd.PersistentFlags().String(flagDefaultChart, "tests/chart", "")
		mockCmd.PersistentFlags().String(flagDefaultChartVersion, "1.0.0", "")
		mockCmd.Flags().String(flagHelmYamlFilename, ".helm.yaml", "")
		mockCmd.Flags().StringP(flagHelmValuesFilename, "f", "values.yaml", "")
		mockCmd.Flags().StringP(flagPostRenderBinary, "p", "", "")

		b, testError := helmGenerate(mockCmd, []string{test.Sample.(string)})
		cmdOutput, err := ioutil.ReadAll(&b)
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}
		expectedPath := test.Expected.(string) + "/output.yaml"
		expectedOutput, err := ioutil.ReadFile(expectedPath)

		shouldFail := false
		if os.IsNotExist(err) {
			shouldFail = true
		} else if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}

		if shouldFail {
			assert.Error(t, testError, "should return error if output is empty")
		} else {
			assert.Equal(t, expectedOutput, cmdOutput, "generated manifests should be the same as the expected output. Error: '%s'", testError)
		}

	}
}

func TestInstallChartWithoutArgs(t *testing.T) {
	currentDir, _ := os.Getwd()
	err := os.Chdir("tests/samples/single-app")
	if err != nil {
		t.Fatalf("couldn't change working directory")
	}
	sampleDir := "samples"
	expectedDir := "../../expected/single-app"
	var tests []TestCase
	test := TestCase{
		Name:     fmt.Sprintf("Parsing test folder %s", sampleDir),
		Sample:   sampleDir,
		Expected: expectedDir,
	}
	tests = append(tests, test)

	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)
		var mockCmd = &cobra.Command{
			Run: func(cmd *cobra.Command, args []string) {
				//nolint:errcheck
				helmGenerate(cmd, args)
			},
		}
		mockCmd.PersistentFlags().String(flagDefaultChart, "../../chart", "")
		mockCmd.PersistentFlags().String(flagDefaultChartVersion, "1.0.0", "")
		mockCmd.Flags().String(flagHelmYamlFilename, ".helm.yaml", "")
		mockCmd.Flags().StringP(flagHelmValuesFilename, "f", "values.yaml", "")
		mockCmd.Flags().StringP(flagPostRenderBinary, "p", "", "")

		b, testError := helmGenerate(mockCmd, nil)
		cmdOutput, err := ioutil.ReadAll(&b)
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}
		expectedOutput, err := ioutil.ReadFile(test.Expected.(string) + "/output.yaml")
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}

		if len(expectedOutput) == 0 {
			assert.Error(t, testError, "should return error if output is empty")
		}
		assert.Equal(t, expectedOutput, cmdOutput, "generated manifests should be the same as the expected output. Error: '%s'", testError)
	}
	err = os.Chdir(currentDir)
	if err != nil {
		t.Fatalf("couldn't change back to default directory")
	}
}

func TestInstallNonExistentPath(t *testing.T) {
	sampleDir := "tests/samples/invalid-dir"
	var tests []TestCase
	test := TestCase{
		Name:     "Testing invalid symlink directory",
		Sample:   sampleDir,
		Expected: bytes.Buffer{},
	}
	tests = append(tests, test)

	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)
		var mockCmd = &cobra.Command{
			Use:   "helm-generate [root-path]",
			Short: "templates helm charts and prints it to stdout",
			Long:  ``,
			Args:  cobra.RangeArgs(0, 1),
			Run: func(cmd *cobra.Command, args []string) {
				//nolint:errcheck
				helmGenerate(cmd, args)
			},
		}
		mockCmd.PersistentFlags().String(flagDefaultChart, "tests/chart", "")
		mockCmd.PersistentFlags().String(flagDefaultChartVersion, "1.0.0", "")
		mockCmd.Flags().String(flagHelmYamlFilename, ".helm.yaml", "")
		mockCmd.Flags().StringP(flagHelmValuesFilename, "f", "values.yaml", "")
		mockCmd.Flags().StringP(flagPostRenderBinary, "p", "", "")

		b, err := helmGenerate(mockCmd, []string{sampleDir})
		assert.Error(t, err, "should have returned an error for a non-existent path")
		assert.Equal(t, test.Expected.(bytes.Buffer), b, "Invalid dir should generate an empty output")
	}
}

func TestInstallEmptyPath(t *testing.T) {
	sampleDir := "tests/samples/empty-dir"
	var tests []TestCase
	test := TestCase{
		Name:     "Testing empty directory",
		Sample:   sampleDir,
		Expected: bytes.Buffer{},
	}
	tests = append(tests, test)

	for i, test := range tests {
		t.Logf("Test case %d: %s", i, test.Name)
		var mockCmd = &cobra.Command{
			Use:   "helm-generate [root-path]",
			Short: "templates helm charts and prints it to stdout",
			Long:  ``,
			Args:  cobra.RangeArgs(0, 1),
			Run: func(cmd *cobra.Command, args []string) {
				//nolint:errcheck
				helmGenerate(cmd, args)
			},
		}
		mockCmd.PersistentFlags().String(flagDefaultChart, "tests/chart", "")
		mockCmd.PersistentFlags().String(flagDefaultChartVersion, "1.0.0", "")
		mockCmd.Flags().String(flagHelmYamlFilename, ".helm.yaml", "")
		mockCmd.Flags().StringP(flagHelmValuesFilename, "f", "values.yaml", "")
		mockCmd.Flags().StringP(flagPostRenderBinary, "p", "", "")

		b, err := helmGenerate(mockCmd, []string{sampleDir})
		assert.NoError(t, err, "should not have returned an error for an empty path")
		assert.Equal(t, test.Expected.(bytes.Buffer), b, "Empty dir should generate an empty output")
	}
}
