
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionString = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the version of s3s2",
	Long:  `Displays the version of s3s2`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("S3S2 Version", versionString)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
