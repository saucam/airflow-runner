package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/saucam/airflow-runner/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)
const Version string = "v0.1"

var src = rand.NewSource(time.Now().UnixNano())
var (
	// Used for flags.
	cfgFile     string
	flowFile    string
	airflowHost string
	dateRange   string
	env         string
)

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func NewRootCommand() *cobra.Command {
	currTime := time.Now()
	rootCmd := &cobra.Command{
		Use:   "airflow-runner",
		Short: "A tool to run airflow jobs concurrently",
		Long: `airflow-runner is a CLI library to automate running airflow jobs.
This application is a tool to run airflow workloads using configuration file
to run multiple jobs concurrently.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			return initConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			var config runner.FlowConfig
			rk := viper.New()
			rk.SetConfigFile(flowFile)
			if err := rk.ReadInConfig(); err == nil {
				runner.Zlog.Info().Msgf("Using flow file: %s", rk.ConfigFileUsed())
			}
			err := rk.Unmarshal(&config)

			if err != nil {
				panic("Unable to unmarshal config")
			}

			runner.ExecuteFlow(airflowHost, config, dateRange)
		},
	}
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file to run the jobs, default to $HOME/.airflowrun.yaml")
	rootCmd.Flags().StringVarP(&flowFile, "flow", "f", "", "flow config file to run the jobs, must be supplied")
	rootCmd.MarkFlagRequired("flow")
	rootCmd.Flags().StringVar(&airflowHost, "host", "localhost:8080", "host-port of airflow service")
	rootCmd.Flags().StringVar(&dateRange, "date", currTime.Format("2006-01-02"), "date range in either comma separated format yyyy-MM-dd,yyyy-MM-dd or range yyyy-MM-dd:yyyy-MM-dd")
	rootCmd.Flags().StringVar(&env, "env", "dev", "Environment name where jobs are supposed to run")

	// rootCmd.PersistentFlags().StringP("host", "h", "", "author name for copyright attribution")
	rootCmd.Flags().Bool("viper", true, "use Viper for configuration")
	rootCmd.AddCommand(VersionCmd)
	return rootCmd
}

// Execute executes the root command.
func Execute() error {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	return nil
}

func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func initConfig(cmd *cobra.Command) error {
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
	viper.SetDefault("license", "apache")
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file: " + viper.ConfigFileUsed())
	}
	// Bind the current command's flags to viper
	bindFlags(cmd)
	// Set up logging
	logDir := "log"
	if viper.IsSet("logdir") {
		logDir = viper.GetString("logdir")
	}
	logFileName := fileNameWithoutExtension(flowFile) + "_" + RandString(7) + ".log"
	// Init logger
	logConfig := &runner.LogConfig{
		ConsoleLoggingEnabled: true,
		FileLoggingEnabled:    true,
		Directory:             logDir,
		FileName:              logFileName}

	runner.Zlog = runner.ConfigureLogger(logConfig)
	runner.Zlog.Info().Msgf("Running airflow-runner version %s", Version)
	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
