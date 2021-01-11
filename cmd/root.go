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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
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
// TODO refactor to make this func "pluggable"(pass filters earlier as a parameter)
// todo implement optimized deleting of empty fields
func Format(in string) (out string, err error) {
	var containerCount int
	out = in
	// check if manifest has containers
	kind := gjson.Get(in, "kind").String()
	if kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" || kind == "Pod" {
		containerCount = int(gjson.Get(in, "spec.template.spec.containers.#").Int())
	}
	// get filters and iterate over them
	filters := gjson.Get(defaultFilters, "filters")
	filters.ForEach(func(key, filter gjson.Result) bool {
		if strings.Contains(filter.String(), "*") {
			if containerCount != 0 {
				for i := 0; i <= containerCount; i++ {
					out, _ = sjson.Delete(out, strings.Replace(filter.String(), "*", fmt.Sprint(i), 1))
				}
			} else {
			}
		} else {
			out, _ = sjson.Delete(out, filter.String())
		}
		result := gjson.Get(out, "spec")
		result.ForEach(func(key, value gjson.Result) bool {
			return true // keep iterating
		})
		return true // keep iterating
	})
	m, _ := gjson.Parse(out).Value().(map[string]interface{})
	deepCleanJSON(m)
	temp, err := json.Marshal(m)
	if err != nil {
		return "nil", fmt.Errorf("error marshalling json after deep cleaning : %v", err)
	}
	out = string(temp)
	return

}

func deepCleanJSON(m map[string]interface{}) {
	// so we range over items
	for k, v := range m {
		// if we find map[string]interface{} we check if it is empty.
		// if it is we delete it and if its not we resursively call the same function
		if reflect.TypeOf(map[string]interface{}{}).Kind() == reflect.TypeOf(v).Kind() {
			if len(v.(map[string]interface{})) == 0 {
				delete(m, k)
			} else {
				deepCleanJSON(v.(map[string]interface{}))
			}
			// if we find json array we call deepCleanJSONArray function
		} else if reflect.TypeOf([]interface{}{}).Kind() == reflect.TypeOf(v).Kind() {
			deepCleanJSONArray(v.([]interface{}))
		}
	}
}
func deepCleanJSONArray(m []interface{}) {
	valueType := reflect.TypeOf(map[string]interface{}{}).Kind()
	_ = valueType
	// so we iterate over json array items
	for k, v := range m {
		// if we find map[string]interface{} we call deepCleanJSON
		if valueType == reflect.TypeOf(v).Kind() {
			deepCleanJSON(v.(map[string]interface{}))
			_ = k
			// and if not we check if it is another array, if it is we call deepCleanJSONArray recursively
		} else if reflect.TypeOf([]interface{}{}).Kind() == reflect.TypeOf(v).Kind() {
			deepCleanJSONArray(v.([]interface{}))
		}
	}
}
func isEmpty(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}
