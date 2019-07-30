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
	"strings"

	"git.turbonomic.com/rgeyer/cloud_pricing_tool/clouds"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var aws_access_key_id string
var aws_secret_access_key string
var aws_acct_file string
var tags []string
var permit_automation bool
var x_acct_role string
var iam_principal_name string
var iam_principal_create bool
var iam_principal_delete bool
var iam_principal_type string
var turbo_hostname string
var turbo_username string
var turbo_password string
var turbo_target_create bool
var turbo_target_delete bool
var turbo_target_prefix string
var turbo_trusted_account_id = "905994805379"
var turbo_trusted_account_role = "foo"
var turbo_trusted_account_instanceid = "*"

var readonly_policy_names = []string{
	"AmazonEC2ReadOnlyAccess",
	"AmazonS3ReadOnlyAccess",
	"AmazonRDSReadOnlyAccess",
}
var automation_policy_names = []string{
	"AmazonEC2FullControl",
	"AmazonS3ReadOnlyAccess",
	"AmazonRDSFullControl",
}

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tags = append(tags, fmt.Sprintf("Turbonomic-Host:%s", turbo_hostname))
		if iam_principal_create || iam_principal_delete {
			aws := clouds.Aws{}
			aws.SetCredentials(aws_access_key_id, aws_secret_access_key)

			log.Info("Querying org for list of child accounts...")

			accts, err := aws.ListAllOrgAccounts()
			if err != nil {
				return fmt.Errorf("Unable to list accounts in the organization. Caused By: %v", err)
			}

			for _, acct := range accts {
				iamCli := iam.New(aws.GetSession(), aws.GetConfig("us-east-1", *acct.Id, x_acct_role))
				acctLog := log.WithField("account", fmt.Sprintf("%v (%v)", *acct.Name, *acct.Id))
				acctLog.Infof("Assuming the role %v on account...", x_acct_role)

				// User flow
				if strings.ToLower(iam_principal_type) == "user" {
					userLog := acctLog.WithField("Username", iam_principal_name)
					userLog.Infof("Searching for users matching the username \"%s\" and/or tags \"%v\"", iam_principal_name, tags)
					matches, err := aws.FindMatchingUsers(*acct.Id, x_acct_role, iam_principal_name, tags)
					if err != nil {
						userLog.Errorf("Failed to query users in the account. Skipping actions in this account. Error: %v", err)
						continue
					}

					// User Create flow
					if iam_principal_create {
						userLog = userLog.WithField("action", "AddUser")
						if len(matches) > 0 {
							// Ask them if they're sure?
							userLog.Warn("A user with the specified username already exists. Duplicate users are not allowed. Please try a different username, or delete the existing user.")
						} else {
							userLog.Info("Creating User")
							cuo, err := iamCli.CreateUser(&iam.CreateUserInput{UserName: &iam_principal_name, Tags: clouds.ConvertStringTagsPointer(tags)})
							if err != nil {
								userLog.Errorf("Failed to create user. Skipping actions in this account. Error: %v", err)
								continue
							}
							userLog.Debug("User Created")

							var policy_set []string

							if permit_automation {
								policy_set = automation_policy_names
								userLog.Infof("Adding Polices to user which allow Turbonomic automation. Policies: %v", policy_set)
							} else {
								policy_set = readonly_policy_names
								userLog.Infof("Adding Polices to user which allow Turbonomic read only access. Policies: %v", policy_set)
							}

							for _, policyname := range policy_set {
								policy_arn := fmt.Sprintf("arn:aws:iam::aws:policy/%s", policyname)
								userLog.Debugf("Attaching policy (%v) to user", policy_arn)
								_, err = iamCli.AttachUserPolicy(&iam.AttachUserPolicyInput{
									UserName:  cuo.User.UserName,
									PolicyArn: &policy_arn,
								})

								if err != nil {
									userLog.Errorf("Unable to attach policy (%v) to user. Continuing with remaining tasks. Error: %v", policy_arn, err)
								}
							}

							access_key_out, err := iamCli.CreateAccessKey(&iam.CreateAccessKeyInput{UserName: cuo.User.UserName})
							if err != nil {
								userLog.Errorf("Failed to create access key and secret for user. Error: %v", err)
							} else {
								// TODO: This is dangerous and shouldn't make it into the prod
								// build!
								userLog.WithFields(logrus.Fields{"Access Key Id": *access_key_out.AccessKey.AccessKeyId, "Secret Access Key": *access_key_out.AccessKey.SecretAccessKey}).Debug("User access key created.")
							}
						}
					} // User Create Flow

					// Delete flow
					if iam_principal_delete {
						userLog = userLog.WithField("action", "UserDelete")
						if len(matches) > 0 {
							// TODO: Show the user the matches, and ask if they are sure!
							userLog.Debug("Deleting the specified user, without any confirmation")

							userLog.Debug("Searching for managed policies associated with user")
							userReadyForDelete := true
							err = iamCli.ListAttachedUserPoliciesPages(&iam.ListAttachedUserPoliciesInput{UserName: matches[0].User.UserName}, func(arg1 *iam.ListAttachedUserPoliciesOutput, arg2 bool) bool {
								for _, policy := range arg1.AttachedPolicies {
									userLog.Debugf("Deleting managed role (%v) from user", *policy.PolicyName)
									_, err = iamCli.DetachUserPolicy(&iam.DetachUserPolicyInput{UserName: &iam_principal_name, PolicyArn: policy.PolicyArn})
									if err != nil {
										userLog.Errorf("Failed to delete managed policy (%v) associated with user. User will not be deleted. Error: %v", *policy.PolicyArn, err)
										userReadyForDelete = false
										return false
									}
								}
								return true
							})
							if err != nil {
								userLog.Errorf("Failed to list policies associated with user. User will not be deleted. Error: %v", err)
								continue
							}

							err = iamCli.ListAccessKeysPages(&iam.ListAccessKeysInput{UserName: matches[0].User.UserName}, func(arg1 *iam.ListAccessKeysOutput, arg2 bool) bool {
								for _, key := range arg1.AccessKeyMetadata {
									userLog.Debugf("Deleting access key (%v) from user", *key.AccessKeyId)
									_, err = iamCli.DeleteAccessKey(&iam.DeleteAccessKeyInput{AccessKeyId: key.AccessKeyId, UserName: matches[0].User.UserName})
									if err != nil {
										userLog.Errorf("Failed to delete access key (%v) associated with user. User will not be deleted. Error: %v", *key.AccessKeyId, err)
										userReadyForDelete = false
										return false
									}
								}
								return true
							})
							if err != nil {
								userLog.Errorf("Failed to list access keys associated with user. User will not be deleted. Error: %v", err)
								continue
							}

							if !userReadyForDelete {
								continue
							}

							_, err := iamCli.DeleteUser(&iam.DeleteUserInput{UserName: matches[0].User.UserName})
							if err != nil {
								userLog.Errorf("Failed to delete user. Error: %v", err)
							}
						} else {
							userLog.Warn("No users matching desired username and tags were found to delete")
						}
					}
				} // User Flow

				// Role flow
				if strings.ToLower(iam_principal_type) == "role" {
					roleLog := acctLog.WithField("Role", iam_principal_name)
					// TODO: Find existing roles as matches first

					// Role create flow
					if iam_principal_create {
						if false { // Existing role

						} else { // END Existing Role - START Role does not already exist
							roleLog.Info("Creating Role")
							assumeRolePolicyDocument := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::%s:root"
      },
      "Action": "sts:AssumeRole",
      "Condition": {}
    }
  ]
}
`

							// assumeRolePolicyDocument = fmt.Sprintf(assumeRolePolicyDocument, turbo_trusted_account_id, turbo_trusted_account_role, turbo_trusted_account_instanceid)
							assumeRolePolicyDocument = fmt.Sprintf(assumeRolePolicyDocument, turbo_trusted_account_id)
							roleLog.Debug(assumeRolePolicyDocument)
							description := "Need a great description here"

							cro, err := iamCli.CreateRole(&iam.CreateRoleInput{
								AssumeRolePolicyDocument: &assumeRolePolicyDocument,
								Description:              &description,
								RoleName:                 &iam_principal_name,
								Tags:                     clouds.ConvertStringTagsPointer(tags),
							})
							if err != nil {
								roleLog.Errorf("Failed to create new role. Error: %v", err)
								continue
							}
							roleLog.Debug("Role Created")

							var policy_set []string

							if permit_automation {
								policy_set = automation_policy_names
								roleLog.Infof("Adding Polices to role which allow Turbonomic automation. Policies: %v", policy_set)
							} else {
								policy_set = readonly_policy_names
								roleLog.Infof("Adding Polices to role which allow Turbonomic read only access. Policies: %v", policy_set)
							}

							for _, policyname := range policy_set {
								policy_arn := fmt.Sprintf("arn:aws:iam::aws:policy/%s", policyname)
								roleLog.Debugf("Attaching policy to role. Policy: %v", policy_arn)
								_, err = iamCli.AttachRolePolicy(&iam.AttachRolePolicyInput{
									RoleName:  cro.Role.RoleName,
									PolicyArn: &policy_arn,
								})

								if err != nil {
									roleLog.Errorf("Unable to attach policy (%s) to role. Continuing with remaining tasks. Error: %v", policyname, err)
								}
							}

							roleLog.Debugf("This ARN to be used to add target. Arn: %v", *cro.Role.Arn)
						} // Role does not already exist
					} // Role create flow
				}
			}

			// 			iamCli := iam.New(sess, &aws.Config{Credentials: stscred})
			// 			// List users, this can be removed ultimately.
			// 			err := iamCli.ListUsersPagesWithContext(context.Background(), &iam.ListUsersInput{},
			// 				func(uo *iam.ListUsersOutput, lastPage bool) bool {
			// 					for _, user := range uo.Users {
			//             if *user.UserName == iam_principal_name
			// 						acctLog.WithFields(log.Fields{"Username": *user.UserName, "User Arn": *user.Arn}).Debug("User iterated")
			// 					}
			// 					return true
			// 				})
			// 			if err != nil {
			// 				acctLog.WithFields(log.Fields{"Caused By": err}).Errorf("Failed to list users for current account. Ignoring and continuing with next account.")
			// 			}
			// 			acctLog.Info("End operating on new account")
			// 		}
			// 		return true // Always continue
			// 	})
			// if err != nil {
			// 	return fmt.Errorf("Unable to list accounts in the organization. Caused By: %v", err)
			// }
		} else {
			log.Info("No IAM principal create, nor delete was requested. Skipping the AWS org account list request.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().StringVar(&aws_access_key_id, "aws-access-key-id", "", "An AWS Access Key ID for a user with AWSOrganizationsReadOnlyAccess and the ability to assume a privileged role in each account in the org")
	awsCmd.Flags().StringVar(&aws_secret_access_key, "aws-secret-access-key", "", "An AWS Secret Access Key for a user with AWSOrganizationsReadOnlyAccess and the ability to assume a privileged role in each account in the org")
	awsCmd.Flags().StringVar(&aws_acct_file, "aws-account-file", "",
		`A filename containing an AWS account list, with optional IAM principal details (AWS Access Key IDs and secrets).

    If set in combination with --iam-principal-create, the child accounts that are found, and the IAM principals which are created will be stored in this file.

    If set in combination with --turbo-target-create this file will be used as the source of target data.`)
	awsCmd.Flags().StringSliceVarP(&tags, "tag", "", []string{"Owner:Turbonomic"}, "One or many tags to add to the IAM principal which is created. Must be in the form of 'Key:Value'. For instance to add an Owner tag with the value 'Sales' you would use --tag Owner:Sales")
	awsCmd.Flags().BoolVarP(&permit_automation, "permit-automation", "", false, "When set, the IAM principals will be created with the policies necessary for Turbonomic to automate.")
	awsCmd.Flags().StringVarP(&x_acct_role, "x-acct-role", "", "OrganizationAccountAccessRole", "The name of a cross account role which has privileges to create the IAM principal in each account.")
	awsCmd.Flags().StringVarP(&iam_principal_name, "iam-principal-name", "", "Turbonomic-OpsMgr-Target", "The name of the IAM principal (User or Role, depending upon --iam-principal-type) which should be created.")
	awsCmd.Flags().BoolVarP(&iam_principal_create, "iam-principal-create", "", false, "When set, the IAM principals will be created.")
	awsCmd.Flags().BoolVarP(&iam_principal_delete, "iam-principal-delete", "", false, "When set, the IAM principals matching the settings (--iam-principal-name --iam-principal-type, and --tag) will be deleted, after verification.")
	awsCmd.Flags().StringVar(&iam_principal_type, "iam-principal-type", "", "One of either 'role' or 'user', indicating the type of authentication to use for the Turbo target.")

	// Move these to a parent, or the root most likely
	awsCmd.Flags().StringVar(&turbo_hostname, "turbo-hostname", "", "The host or ip address of your Turbonomic OpsMgr.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// awsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// awsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
