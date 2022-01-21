package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of airflow-runner",
	Long:  `All software has versions. This is airflow-runner's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("airflow-runner v0.1 -- HEAD")
	},
}
