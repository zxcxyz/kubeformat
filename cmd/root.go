/*
Package cmd blah

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
)

var filtersPath string
var output string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubeformat",
	Short: "Tool designed to remove junk from kubectl manifests",
	Long: `Usage:
	kubectl get deployment -o yaml/json | kubeformat`,
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
		if strings.ToLower(output) == "json" {
			out, _ = yaml.YAMLToJSON(out)

			if err != nil {
				cmd.Print(fmt.Errorf("error converting from yaml to json : %v", err))
			}

		}
		cmd.Print(string(out))
		// if output == "yaml" {
		// 	cmd.Print(string(out))
		// } else if output == "json" {
		// 	jsonbytes, err := yaml.YAMLToJSON()
		// 	if err != nil {
		// 		return err
		// 	}
		// 	cmd.Print(string(jsonbytes))
		// }

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
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&filtersPath, "filtersPath", "p", "", "Path to your filters json. For right json template please refer to https://github.com/zxcxyz/kubeformat/blob/master/cmd/defaults.go")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "yaml", "Output format. json/yaml. Default yaml.")
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
	var filters gjson.Result
	out = in
	// check if manifest has containers
	kind := gjson.Get(in, "kind").String()
	if kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" || kind == "Pod" {
		containerCount = int(gjson.Get(in, "spec.template.spec.containers.#").Int())
	}

	// get filters and iterate over them
	if filtersPath != "" {
		d, err := ioutil.ReadFile(filtersPath)
		if err != nil {
			// why we return "nil" here instead of nil???????
			return "nil", fmt.Errorf("error reading filters from file : %v", err)
		}
		filters = gjson.Get(string(d), "filters")
	} else {
		filters = gjson.Get(defaultFilters, "filters")
	}
	// anonymous function. why it is here? nobody knows
	filters.ForEach(func(key, filter gjson.Result) bool {
		if strings.Contains(filter.String(), "*") {
			if containerCount != 0 {
				for i := 0; i <= containerCount; i++ {
					out, _ = sjson.Delete(out, strings.Replace(filter.String(), "*", fmt.Sprint(i), 1))
				}
			}
		} else {
			out, _ = sjson.Delete(out, filter.String())
		}
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

// this function is used to strip json of fields without a value
func deepCleanJSON(m map[string]interface{}) {
	// so we range over json fields
	for k, v := range m {
		// if we find map[string]interface{} which is just a container with unknown stuff we check if it is empty.
		// if it is we delete it and if its not we resursively call the same function
		if reflect.TypeOf(map[string]interface{}{}).Kind() == reflect.TypeOf(v).Kind() {
			if len(v.(map[string]interface{})) == 0 {
				delete(m, k)
			} else {
				deepCleanJSON(v.(map[string]interface{}))
			}
			// all code above isnt smart enough to traverse into arrays, here we fix that
		} else if reflect.TypeOf([]interface{}{}).Kind() == reflect.TypeOf(v).Kind() {
			for _, j := range v.([]interface{}) {
				// if we find map[string]interface{} we go deeper, if not its fine
				if reflect.TypeOf(map[string]interface{}{}).Kind() == reflect.TypeOf(j).Kind() {
					deepCleanJSON(j.(map[string]interface{}))
				} else if reflect.TypeOf([]interface{}{}).Kind() == reflect.TypeOf(j).Kind() {
					// dunno if this is needed tbh, this is for recursively traversing in arrays
					deepCleanJSON(j.(map[string]interface{}))
				}
			}
		}
	}
}
func isEmpty(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}
