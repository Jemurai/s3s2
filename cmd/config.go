
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/user"

	"github.com/jemurai/s3s2/options"
	log "github.com/sirupsen/logrus"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Build a configuration file",
	Long:  `Build a configuration file so that we can run the tool with exactly the options we want.`,
	Run: func(cmd *cobra.Command, args []string) {
		fn, _ := cmd.PersistentFlags().GetString("file")

		fmt.Println("Please specify a bucket.")
		bucket := prompt.Input("> ", completer)

		fmt.Println("Please specify a region.")
		region := prompt.Input("> ", completer)

		fmt.Println("Please specify an org.")
		org := prompt.Input("> ", completer)

		fmt.Println("Please specify a working directory.")
		dir := prompt.Input("> ", completer)

		fmt.Println("Please specify a file prefix (nothing sensitive).")
		prefix := prompt.Input("> ", completer)

		fmt.Println("Please specify a public key to use (file path or url).")
		pubkey := prompt.Input("> ", completer)

		bc := options.Options{
			Directory: dir,
			Bucket:    bucket,
			Org:       org,
			Region:    region,
			Prefix:    prefix,
			PubKey:    pubkey,
		}
		data, _ := json.MarshalIndent(bc, "", " ")
		err := ioutil.WriteFile(fn, data, 0644)
		if err != nil {
			log.Warn(err)
			log.Panic(err)
		} else {
			log.Info("Your config was written to", fn, ". You can invoke with s3s2 --config", fn)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	usr, _ := user.Current()
	defaultPath := usr.HomeDir + "/.s3s2"
	configCmd.PersistentFlags().String("file", defaultPath, "The config file to write.")
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
