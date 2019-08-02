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
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"

	. "git.turbonomic.com/rgeyer/cloud_pricing_tool/cloud_pricing_tool/ratecards"
	"github.com/spf13/cobra"
)

type Prices struct {
	OnDemand          float64
	OneYearReserved   float64
	ThreeYearReserved float64
}

var region = "US East (N. Virginia)"
var instanceSize = "m5.large"
var hoursPerMonth = 732
var result map[string]map[string]Prices = make(map[string]map[string]Prices)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Round(x float64) float64 {
	return math.Round(x/0.01) * 0.01
}

// templatePriceCmd represents the templatePrice command
var templatePriceCmd = &cobra.Command{
	Use:   "templatePrice",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Errorf("Searching for Instance Type %s in Region %s", instanceSize, region)
		jsonfile, err := os.Open("/Volumes/RJGShared/Turbo/AWS Pricing/srv/tomcat/data/repos/AmazonEC2.json")
		//jsonfile, err := os.Open("./littlecard.json")
		if err != nil {
			return err
		}

		defer jsonfile.Close()

		json_bytes, err := ioutil.ReadAll(jsonfile)
		if err != nil {
			return err
		}

		var rt Root

		err = json.Unmarshal(json_bytes, &rt)
		if err != nil {
			return err
		}

		var matches []Product
		for _, v := range rt.Products {
			if v.ProductFamily == "Compute Instance" &&
				v.Attributes.Location == region &&
				v.Attributes.InstanceType == instanceSize &&
				v.Attributes.Tenancy == "Shared" &&
				v.Attributes.LicenseModel == "No License required" &&
				strings.Contains(v.Attributes.UsageType, "BoxUsage") {
				matches = append(matches, v)
			}
		}

		for _, prd := range matches {
			onDemandMonthly := 0.00
			oneYearReservedMonthly := 0.00
			threeYearReservedMonthly := 0.00
			for k, offerTerms := range rt.Terms.OnDemand {
				if k == prd.SKU {
					for _, offerTerm := range offerTerms {
						for _, priceDim := range offerTerm.PriceDimensions {
							hourly, _ := strconv.ParseFloat(priceDim.PricePerUnit["USD"], 64)
							onDemandMonthly = hourly * 732
						}
					}
				}
			}

			for k, offerTerms := range rt.Terms.Reserved {
				if k == prd.SKU {
					for _, offerTerm := range offerTerms {
						for _, priceDim := range offerTerm.PriceDimensions {
							if offerTerm.TermAttributes.LeaseContractLength == "1yr" &&
								offerTerm.TermAttributes.OfferingClass == "standard" &&
								offerTerm.TermAttributes.PurchaseOption == "All Upfront" &&
								priceDim.Description == "Upfront Fee" {
								oneTime, _ := strconv.ParseFloat(priceDim.PricePerUnit["USD"], 64)
								oneYearReservedMonthly = oneTime / 12
								// if prd.Attributes.OperatingSystem == "Windows" && prd.Attributes.PreInstalledSw == "NA" {
								// 	fmt.Println(priceDim)
								// 	fmt.Println(oneTime)
								// 	fmt.Println(oneYearReservedMonthly)
								// 	fmt.Println(err)
								// }
							}

							if offerTerm.TermAttributes.LeaseContractLength == "3yr" &&
								offerTerm.TermAttributes.OfferingClass == "standard" &&
								offerTerm.TermAttributes.PurchaseOption == "All Upfront" &&
								priceDim.Description == "Upfront Fee" {
								oneTime, _ := strconv.ParseFloat(priceDim.PricePerUnit["USD"], 64)
								threeYearReservedMonthly = oneTime / 36
							}
						}
					}
				}
			}

			prices := Prices{
				OnDemand:          Round(onDemandMonthly),
				OneYearReserved:   Round(oneYearReservedMonthly),
				ThreeYearReserved: Round(threeYearReservedMonthly),
			}
			_, ok := result[prd.Attributes.OperatingSystem]
			if !ok {
				result[prd.Attributes.OperatingSystem] = make(map[string]Prices)
			}
			_, ok = result[prd.Attributes.OperatingSystem][prd.Attributes.PreInstalledSw]
			if !ok {
				result[prd.Attributes.OperatingSystem][prd.Attributes.PreInstalledSw] = prices
			} else {
				fmt.Println(prd)
				fmt.Println(prices)
				fmt.Println(result[prd.Attributes.OperatingSystem][prd.Attributes.PreInstalledSw])
			}
		}

		result_json, err := json.Marshal(result)
		if err != nil {
			return err
		}

		fmt.Println(string(result_json))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(templatePriceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// templatePriceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// templatePriceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	templatePriceCmd.Flags().StringVar(&region, "region", "", "region")
	templatePriceCmd.Flags().StringVar(&instanceSize, "instance-size", "", "Instance Size")
}
