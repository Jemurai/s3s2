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

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

// BaseConfig is the core configuration that s3s3 uses.
type BaseConfig struct {
	Bucket string
	Org    string
	Dir    string
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Build a configuration file",
	Long:  `Build a configuration file so that we can run the tool with exactly the options we want.`,
	Run: func(cmd *cobra.Command, args []string) {
		fn, _ := cmd.PersistentFlags().GetString("file")

		fmt.Println("Please specify a bucket.")
		bucket := prompt.Input("> ", completer)

		fmt.Println("Please specify an org.")
		org := prompt.Input("> ", completer)

		fmt.Println("Please specify a working directory.")
		dir := prompt.Input("> ", completer)

		bc := BaseConfig{bucket, org, dir}
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
	s := []prompt.Suggest{
	//		{Text: "users", Description: "Store the username and age"},
	//		{Text: "articles", Description: "Store the article text posted by user"},
	//		{Text: "comments", Description: "Store the text commented to articles"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
