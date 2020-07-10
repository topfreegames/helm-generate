package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "helm-generate [root-path]",
	Short: "templates helm charts and prints it to stdout",
	Long:  ``,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		buf, err := helmGenerate(cmd, args)
		if err != nil {
			log.Fatalf("Error generating helm template: %s", err)
		}
		fmt.Printf("%v", buf.String())
	},
}

var (
	flagDefaultChart        = "default-chart"
	flagDefaultChartVersion = "default-chart-version"
	flagPostRenderBinary    = "post-render-binary"
	flagHelmYamlFilename    = "helm-yaml"
	flagHelmValuesFilename  = "values-yaml"
)

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
}

func init() {
	cobra.OnInitialize(initConfig)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	if err := viper.BindEnv(flagDefaultChart, "HELM_DEFAULT_CHART"); err != nil {
		log.Fatalf("error initializing viper for env HELM_DEFAULT_CHART")
	}

	rootCmd.PersistentFlags().String(flagDefaultChart, viper.GetString(flagDefaultChart), "Chart to be used to render values.yaml's by default (Defaults to HELM_DEFAULT_CHART env var)")
	if err := viper.BindPFlag(flagDefaultChart, rootCmd.PersistentFlags().Lookup(flagDefaultChart)); err != nil {
		log.Fatalf("error binding viper for flag HELM_DEFAULT_CHART")
	}
	if err := viper.BindEnv(flagDefaultChartVersion, "HELM_DEFAULT_CHART_VERSION"); err != nil {
		log.Fatalf("error initializing viper for env HELM_DEFAULT_CHART_VERSION")
	}
	rootCmd.PersistentFlags().String(flagDefaultChartVersion, viper.GetString(flagDefaultChartVersion), "Version of the default chart (Default to HELM_DEFAULT_CHART_VERSION env var)")
	if err := viper.BindPFlag(flagDefaultChartVersion, rootCmd.PersistentFlags().Lookup(flagDefaultChartVersion)); err != nil {
		log.Fatalf("error binding viper for flag HELM_DEFAULT_CHART_VERSION")
	}

	rootCmd.Flags().String(flagHelmYamlFilename, ".helm.yaml", "File to look for helm chart configuration (Defaults to .helm.yaml)")
	rootCmd.Flags().StringP(flagHelmValuesFilename, "f", "values.yaml", "Filename of the helm values file (Defaults to values.yaml)")
	rootCmd.Flags().StringP(flagPostRenderBinary, "p", "", "A command to run after rendering the Helm templates")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf(err.Error())
	}
}
