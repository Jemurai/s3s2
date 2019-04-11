// Copyright © 2019 Matt Konda <mkonda@jemurai.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/user"

	"github.com/jemurai/s3s2/options"

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
			fmt.Println(err)
			panic(err)
		} else {
			fmt.Println("Your config was written to", fn, ". You can invoke with s3s2 --config", fn)
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
