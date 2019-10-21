
package cmd

import (
	"fmt"

	"github.com/tempuslabs/s3s2/encrypt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var keydir string
var keyprefix string

// genkeyCmd represents the genkey command
var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate new gpg keys.",
	Long:  `Generate new gpg keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Generating new keys with name %s in: %s", keyprefix, keydir)
		encrypt.GenerateKeys(keydir, keyprefix, 4096)
	},
}

func init() {
	rootCmd.AddCommand(genkeyCmd)

	genkeyCmd.PersistentFlags().StringVar(&keydir, "keydir", "", "The directory to write the key files to.")
	genkeyCmd.PersistentFlags().StringVar(&keyprefix, "keyprefix", "", "The directory to write the key files to.")

	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
}
