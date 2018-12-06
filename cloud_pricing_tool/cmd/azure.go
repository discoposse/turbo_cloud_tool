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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var subscription_id string
var tenant_id string
var client_id string
var client_secret string

// azureCmd represents the azure command
var azureCmd = &cobra.Command{
	Use:   "azure",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Validation of the flags, which are all required.
		client_secret = viper.GetString("azure-client-secret")
		client_id = viper.GetString("azure-client-id")
		subscription_id = viper.GetString("azure-subscription-id")
		tenant_id = viper.GetString("azure-tenant-id")

		os.Setenv("AZURE_CLIENT_SECRET", client_secret)
		os.Setenv("AZURE_CLIENT_ID", client_id)
		os.Setenv("AZURE_TENANT_ID", tenant_id)

		client := &http.Client{}
		a, err := auth.NewAuthorizerFromEnvironment()
		if err != nil {
			return err
		}

		filter := "%24filter=OfferDurableId%20eq%20'MS-AZR-0003p'%20and%20Currency%20eq%20'USD'%20and%20Locale%20eq%20'en-US'%20and%20RegionInfo%20eq%20'US'"
		url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Commerce/RateCard?api-version=2016-08-31-preview&%s",
			subscription_id,
			filter,
		)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")
		dec_req, err := autorest.Prepare(req, a.WithAuthorization())
		if err != nil {
			return err
		}
		resp, err := client.Do(dec_req)
		if err != nil {
			return err
		}

		ratecard_json, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		ratecard_filename := fmt.Sprintf("%d-ratecard.json", int32(time.Now().Unix()))
		err = ioutil.WriteFile(ratecard_filename, ratecard_json, 0644)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	ratecardDownloadCmd.AddCommand(azureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// azureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// azureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	azureCmd.Flags().StringVar(&subscription_id, "azure-subscription-id", "", "Your Azure Subscription ID")
	azureCmd.Flags().StringVar(&tenant_id, "azure-tenant-id", "", "Your Azure Tenant ID")
	azureCmd.Flags().StringVar(&client_id, "azure-client-id", "", "The client id of a service principal with access to your tenant")
	azureCmd.Flags().StringVar(&client_id, "azure-client-secret", "", "The client secret of a service principal with access to your tenant")
	viper.BindPFlag("azure-subscription-id", azureCmd.Flags().Lookup("azure-subscription-id"))
	viper.BindPFlag("azure-tenant-id", azureCmd.Flags().Lookup("azure-tenant-id"))
	viper.BindPFlag("azure-client-id", azureCmd.Flags().Lookup("azure-client-id"))
	viper.BindPFlag("azure-client-secret", azureCmd.Flags().Lookup("azure-client-secret"))
}
