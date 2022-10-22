/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"gately/internal/app"
	"gately/internal/config"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (

	// The environment variable prefix for all environment variables
	envPrefix = "GATELY"
)

var appConfig config.AppConfig

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the gately server",
	Long:  `Starts the server and listens on port 80`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Bind env vars to server flags. If a flag is not defined, it will take the env var
		// Bind each cmdline flag to its corresponding environment variable
		// On binding cmdline flags to environment variables, ex: a cmdline flag --port
		// binds to a prefixed environment variable GATELY_PORT.

		if err := bindEnvVarsToFlags(cmd); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Final config combining both flags and env vars
		finalConf := viper.New()

		if err := finalConf.BindPFlags(cmd.Flags()); err != nil {
			fmt.Println("Unable to map flags to config")
		}

		// Convert Config map to config struct
		err := mapstructure.Decode(finalConf.AllSettings(), &appConfig)
		if err != nil {
			fmt.Println("Unable to unmarshall configs")
		}
		fmt.Printf("Final Conf %+v", appConfig)

		app.Run(appConfig)

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	// Define the cmdline flags
	// Order of precedence is is_set(Flag value) > is_set(Corresponding env var) > Default flag value
	// Passwords and usernames should come only from env vars.
	runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	runCmd.Flags().StringP("port", "p", "8080", "Gately application port")
	runCmd.Flags().StringP("redis-host", "r", "redis:6379", "Redis host")
	runCmd.Flags().StringP("redis-user", "", "", "")
	_ = runCmd.Flags().MarkHidden("redis-user")
	runCmd.Flags().StringP("redis-pass", "", "", "")
	_ = runCmd.Flags().MarkHidden("redis-pass")
	runCmd.Flags().StringP("mongo-host", "c", "mongo:27017", "MongoDB host")
	runCmd.Flags().StringP("mongo-db-name", "", "testDB",
		"Database that stores URL mappings in MongoDB")
	runCmd.Flags().StringP("mongo-collection-name", "", "testCollection",
		"Mongo Collection that stores URL mappings in MongoDB")
	runCmd.Flags().StringP("mongo-user", "", "", "")
	_ = runCmd.Flags().MarkHidden("mongo-user")
	runCmd.Flags().StringP("mongo-pass", "", "", "")
	_ = runCmd.Flags().MarkHidden("mongo-pass")
}

// Bind each cmdline flag to its corresponding environment variable
func bindEnvVarsToFlags(cmd *cobra.Command) error {
	v := viper.New()

	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Convert flag names to env vars
		envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		envVar := fmt.Sprintf("%s_%s", envPrefix, envVarSuffix)
		err := v.BindEnv(f.Name, envVar)
		fmt.Println(v.GetString(f.Name))
		if err != nil {
			fmt.Println("Unable to parse environment variable")
			os.Exit(-1)
		}

		// Set env var to config if cmdline flags are not set
		// This will be true for DB credentials
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})

	return nil
}
