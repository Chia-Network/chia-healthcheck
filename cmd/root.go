package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chia-healthcheck",
	Short: "Simple healthcheck for Chia Blockchain",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var (
		hostname        string
		healthcheckPort int
	)

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chia-healthcheck.yaml)")

	rootCmd.PersistentFlags().StringVar(&hostname, "hostname", "localhost", "The hostname to connect to")
	rootCmd.PersistentFlags().IntVar(&healthcheckPort, "healthcheck-port", 9914, "The port the metrics server binds to")

	cobra.CheckErr(viper.BindPFlag("hostname", rootCmd.PersistentFlags().Lookup("hostname")))
	cobra.CheckErr(viper.BindPFlag("healthcheck-port", rootCmd.PersistentFlags().Lookup("healthcheck-port")))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".chia-healthcheck" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".chia-healthcheck")
	}

	viper.SetEnvPrefix("CHIA_HEALTHCHECK")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
