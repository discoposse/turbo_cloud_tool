// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	. "git.turbonomic.com/rgeyer/cloud_pricing_tool/clouds"
)

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
		create := false
		var err error
		azure := &Azure{}
		saved_logins, ok := viper.Get("azure-tokens").([]interface{})
		if ok && len(saved_logins) > 0 {
			spTokenJson := saved_logins[0].(string)
			err = azure.SetAuthServiceToken(spTokenJson)
			if err != nil {
				return err
			}
		} else if client_id != "" {
			err = azure.SetAuthClientCredentials(
				client_id,
				client_secret,
				tenant_id)
			if err != nil {
				return err
			}
		} else {
			err = azure.SetAuthDeviceFlow("ce3d555d-4fc9-4c34-898d-1d790a9b5c4b", "common")
			if err != nil {
				return fmt.Errorf("Unable to get token from device code flow. Caused By: %v", err)
			}

			if tenant_id == "" {
				tenCli, err := azure.GetTenantsClient()
				if err != nil {
					return fmt.Errorf("Unable to get Tenants Client. Caused By: %v", err)
				}

				tenList, err := tenCli.ListComplete(context.Background())
				if err != nil {
					return fmt.Errorf("Unable to get list of tenants for logged in user. Caused By: %v", err)
				}

				for {
					if !tenList.NotDone() {
						break
					}

					ten := tenList.Value()
					fmt.Println(*ten.TenantID)
					tenant_id = *ten.TenantID

					err = tenList.Next()
					if err != nil {
						return fmt.Errorf("Unable to iterate to next Tenant record in tenant list response. Caused By: %v", err)
					}
				}
			}

			// viper.Set("azure-tokens", []string{string(spTokenJson)})
			// viper.WriteConfig()
		}

		// rbc, err := azure.GetRbacGraphSignedInUserClient()
		// if err != nil {
		// 	return fmt.Errorf("Unable to get Azure Signed In User Client. Caused By: %v", err)
		// }
		//
		// me, err := rbc.Get(context.Background())
		// if err != nil {
		// 	return fmt.Errorf("Unable to determine signed in Azure user. Caused By: %v", err)
		// }
		//
		// fmt.Println(me.DisplayName)

		// Get an application client, and list all the applications (probably unecessary)
		appCli, err := azure.GetApplicationsClient(tenant_id)
		if err != nil {
			return fmt.Errorf("Unable to get Applications Client. Caused By: %v", err)
		}
		apps, err := appCli.ListComplete(context.Background(), "")
		if err != nil {
			return fmt.Errorf("Unable to list Service Principals in AD Tenant ID %s. Caused By: %v", tenant_id, err)
		}

		for {
			if !apps.NotDone() {
				break
			}
			app := apps.Value()
			if sp, ok := app.AsApplication(); ok {
				// TODO: When would this ever happen? We're querying ServicePrincipal
				// explicitly, then casting to ServicePrincipal. It should never fail..
				fmt.Printf("%s - %s\n", *sp.DisplayName, *sp.AppID)
			} else {
				return fmt.Errorf("Could not assert to ServicePrincipal")
			}
			err = apps.Next()
			if err != nil {
				return fmt.Errorf("Unable to fetch next service principal in list: %v", err)
			}
		}

		if create {
			// Time to create a new application
			pwd, err := password.Generate(32, 10, 10, false, false)
			if err != nil {
				return fmt.Errorf("Unable to generate a random secure password for the new Service Principal. Caused By: %v", err)
			}
			fmt.Println(pwd)
			displayName := "Turbonomic ARM Service Principal"
			endDate := date.Time{time.Now().Local().Add(time.Second * time.Duration(60*60*24*365))}
			credary := []graphrbac.PasswordCredential{
				graphrbac.PasswordCredential{
					EndDate: &endDate,
					Value:   &pwd,
				},
			}
			availableToOtherTenants := false

			newAppParams := graphrbac.ApplicationCreateParameters{
				DisplayName:             &displayName,
				PasswordCredentials:     &credary,
				AvailableToOtherTenants: &availableToOtherTenants,
				IdentifierUris:          &[]string{},
			}

			newApp, err := appCli.Create(context.Background(), newAppParams)
			if err != nil {
				return fmt.Errorf("Unable to create new Service Principal. Caused By: %v", err)
			}

			appBytes, err := newApp.MarshalJSON()
			if err != nil {
				return fmt.Errorf("Unable to marshal new Service Principal response. Caused By: %v", err)
			}

			fmt.Println(string(appBytes))

			// Now create a service principal associated with the application
			spCli, err := azure.GetServicePrincipalsClient(tenant_id)
			if err != nil {
				return fmt.Errorf("Unable to get Service Principal Client. Caused By: %v", err)
			}

			_, err = spCli.Create(context.Background(), graphrbac.ServicePrincipalCreateParameters{AppID: newApp.AppID, Tags: &[]string{"Turbonomic"}})
			if err != nil {
				return fmt.Errorf("Unable to create new Service Principal from application ID %s. Caused By: %v", *newApp.AppID, err)
			}
		}

		subsCli, err := azure.GetSubscriptionClient()
		if err != nil {
			return fmt.Errorf("Unable to get Azure Subscriptions client. Caused By: %v", err)
		}
		subPages, err := subsCli.ListComplete(context.Background())
		if err != nil {
			return err
		}

		for {
			if !subPages.NotDone() {
				break
			}
			sub := subPages.Value()
			subJson, err := json.MarshalIndent(sub, "", "  ")
			if err != nil {
				fmt.Printf("Unable to marshal subscription to json: %v", err)
				return err
			} else {
				fmt.Println(string(subJson))
			}
			fmt.Println(*sub.SubscriptionID)

			err = subPages.Next()
			if err != nil {
				fmt.Printf("Unable to fetch next subscrption in list: %v", err)
				return err
			}
		}

		_, err = subsCli.Get(context.Background(), "dae48673-5abb-4690-9bf6-916690da3969")
		if err != nil {
			return fmt.Errorf("Unable to get a subscription even as a global admin. Caused By: %v", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(azureCmd)

	azureCmd.Flags().StringVar(&tenant_id, "azure-tenant-id", "", "Your Azure Tenant ID")
	azureCmd.Flags().StringVar(&client_id, "azure-client-id", "", "The client id of a service principal with access to your tenant")
	azureCmd.Flags().StringVar(&client_secret, "azure-client-secret", "", "The client secret of a service principal with access to your tenant")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// azureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// azureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
