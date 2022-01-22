package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/saucam/airflow-runner/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile     string
	flowFile    string
	airflowHost string
	dateRange   string
	env         string

	rootCmd = &cobra.Command{
		Use:   "airflow-runner",
		Short: "A tool to run airflow jobs concurrently",
		Long: `airflow-runner is a CLI library to automate running airflow jobs.
This application is a tool to run airflow workloads using configuration file
to run multiple jobs concurrently.`,
		Run: func(cmd *cobra.Command, args []string) {
			var config runner.FlowConfig
			rk := viper.New()
			rk.SetConfigFile(flowFile)
			if err := rk.ReadInConfig(); err == nil {
				fmt.Println("Using flow file:", rk.ConfigFileUsed())
			}
			err := rk.Unmarshal(&config)

			if err != nil {
				panic("Unable to unmarshal config")
			}

			runner.ExecuteFlow(airflowHost, config, dateRange)
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	currTime := time.Now()

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file to run the jobs, default to $HOME/.airflowrun.yaml")
	rootCmd.PersistentFlags().StringVar(&flowFile, "flow", "", "flow config file to run the jobs, must be supplied")
	rootCmd.MarkFlagRequired("flow")
	rootCmd.PersistentFlags().StringVar(&airflowHost, "host", "localhost:8080", "host-port of airflow service")
	rootCmd.PersistentFlags().StringVar(&dateRange, "date", currTime.Format("2006-01-02"), "date range in either comma separated format yyyy-MM-dd,yyyy-MM-dd or range yyyy-MM-dd:yyyy-MM-dd")
	rootCmd.PersistentFlags().StringVar(&env, "env", "dev", "Environment name where jobs are supposed to run")

	// rootCmd.PersistentFlags().StringP("host", "h", "", "author name for copyright attribution")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("date", rootCmd.PersistentFlags().Lookup("date"))
	viper.BindPFlag("env", rootCmd.PersistentFlags().Lookup("env"))

	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	viper.SetDefault("license", "apache")

	// rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".airflowrun" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".airflowrun")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
