/*
Copyright © 2021 zxcxyz <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubeformat",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var in, out []byte
		var err error
		stdin := cmd.InOrStdin()
		in, err = ioutil.ReadAll(stdin)
		if err != nil {
			return err
		}
		out, err = ToFormattedYaml(in)
		if err != nil {
			return err
		}
		cmd.Print(string(out))
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubeformat.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kubeformat" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kubeformat")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// ToFormattedYaml used to format input json or yaml to clean yaml
func ToFormattedYaml(in []byte) (out []byte, err error) {
	var injson string
	isYaml := !isJSON(in)
	if isYaml {
		injsonbytes, err := yaml.YAMLToJSON(in)
		if err != nil {
			return nil, fmt.Errorf("error converting from yaml to json : %v", err)
		}
		injson = string(injsonbytes)
	} else {
		injson = string(in)
	}
	// if we got list flatten it into array of json strings
	kind := gjson.Get(injson, "kind").String()
	if kind == "List" {
		items := gjson.Get(injson, "items").Array()
		for i, item := range items {
			_ = i
			itemFormatted, err := Format(item.String())
			if err != nil {
				return nil, fmt.Errorf("error formatting json to yaml in a list: %v", err)
			}
			var temp []byte
			temp, err = yaml.JSONToYAML([]byte(itemFormatted))
			if err != nil {
				return nil, fmt.Errorf("error converting from json to yaml : %v", err)
			}
			if i+1 != len(items) {
				temp = append(temp, []byte("\n---\n\n")...)
			}
			out = append(out, temp...)
		}
	} else {
		itemFormatted, err := Format(injson)
		if err != nil {
			return nil, fmt.Errorf("error formatting single json to yaml : %v", err)
		}
		out, err = yaml.JSONToYAML([]byte(itemFormatted))
		if err != nil {
			return nil, fmt.Errorf("error converting from json to yaml : %v", err)
		}
	}
	return
}

// Format removes useless fields from kubernetes manifest
func Format(in string) (out string, err error) {
	out = in
	// read filters from json array defaultFilters and delete fields according to them
	filterCount := int(gjson.Get(defaultFilters, "#").Int())
	for i := 0; i < filterCount; i++ {
		out, _ = sjson.Delete(out, gjson.Get(defaultFilters, fmt.Sprint(i)).String())
	}
	kind := gjson.Get(in, "kind").String()
	switch kind {
	case
		"Deployment",
		"StatefulSet",
		"DaemonSet",
		"Pod":
		// gets container count and loops over each container with filters from containerFilters
		containerFilterCount := int(gjson.Get(containerFilters, "#").Int())
		containerCount := int(gjson.Get(in, "spec.template.spec.containers.#").Int())

		for i := 0; i <= containerCount; i++ {
			containerNumber := fmt.Sprint(i)
			for j := 0; j < containerFilterCount; j++ {
				filterNumber := fmt.Sprint(j)
				filter := strings.Replace(gjson.Get(containerFilters, filterNumber).String(), "*", containerNumber, 1)
				out, _ = sjson.Delete(out, filter)
				// fmt.Println(filter)
			}
		}
	default:
	}
	return
}
