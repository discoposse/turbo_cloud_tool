// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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

	. "git.turbonomic.com/rgeyer/cloud_pricing_tool/turbo"
	"github.com/spf13/cobra"
)

var srcfile string
var destfile string
var memFudgeFactor int
var cpuFudgeFactor int

// templateMatchCmd represents the templateMatch command
var templateMatchCmd = &cobra.Command{
	Use:   "templateMatch [template]",
	Short: "Matches the provided template from the --source cost topology file with templates in the --dest cost topology file within the given --fudge-factor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src_topo, err := NewCostTopoFile(srcfile)
		if err != nil {
			return fmt.Errorf("Unable to load source topology file. Error: %s", err)
		}
		dst_topo, err := NewCostTopoFile(destfile)
		if err != nil {
			return fmt.Errorf("Unable to load destination topology file. Error: %s", err)
		}

		// Find the source template first
		src_profiles, err := src_topo.GetProfiles()
		if err != nil {
			return fmt.Errorf("Unable to get source templates. Error: %s", err)
		}

		dst_profiles, err := dst_topo.GetProfiles()
		if err != nil {
			return fmt.Errorf("Unable to get destination templates. Error: %s", err)
		}

		dst_costs, err := dst_topo.GetOnDemandCosts()
		if err != nil {
			return fmt.Errorf("Unable to get destination costs. Error: %s", err)
		}
		for _, profile := range src_profiles {
			if profile.Name == args[0] {
				fmt.Println(fmt.Sprintf("Found profile matching %s in source topology, looking for a suitable equivalent in the destination topology", args[0]))
				dest_profile, other_candidates := profile.FindClosestMatch(dst_profiles, dst_costs, memFudgeFactor, cpuFudgeFactor)
				output := map[string]interface{}{
					"source_template":  profile,
					"chosen_template":  dest_profile,
					"other_candidates": other_candidates,
				}
				json_bytes, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return fmt.Errorf("Unable to format output response to JSON. Error: %s", err)
				}
				fmt.Print(string(json_bytes))
				return nil
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(templateMatchCmd)

	templateMatchCmd.Flags().StringVarP(&srcfile, "source", "s", "", "The cost.topology file containing the [template] passed into the command")
	templateMatchCmd.Flags().StringVarP(&destfile, "dest", "d", "", "The cost.topology file containing the destination template options to compare with [template]")
	templateMatchCmd.Flags().IntVarP(&memFudgeFactor, "mem-fudge-factor", "m", 10, "The amount (in percent) the target template can be under or over memory and still match.")
	templateMatchCmd.Flags().IntVarP(&cpuFudgeFactor, "cpu-fudge-factor", "c", 10, "The amount (in percent) the target template can be under or over CPU and still match.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// templateMatchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// templateMatchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
