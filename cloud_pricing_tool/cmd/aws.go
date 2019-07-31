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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"unicode"

	"git.turbonomic.com/rgeyer/cloud_pricing_tool/cloud_pricing_tool/lib"
	"git.turbonomic.com/rgeyer/cloud_pricing_tool/clouds"
	"github.com/aws/aws-sdk-go/service/iam"
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
var turbo_trusted_account_id string
var turbo_trusted_account_role string
var turbo_trusted_account_instanceid string

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
		stdinReader := bufio.NewReader(os.Stdin)
		var output *lib.AwsTargetCmdOutput
		if len(aws_acct_file) > 0 {
			if _, err := os.Stat(aws_acct_file); err == nil {
				jFile, err := os.Open(aws_acct_file)
				if err != nil {
					log.Fatalf("Unable to open AWS account file %s. Error: %v", aws_acct_file, err)
					return nil
				}
				jFileBytes, err := ioutil.ReadAll(jFile)
				if err != nil {
					log.Fatalf("Unable to read contents of AWS account file %s. Error: %v", aws_acct_file, err)
					return nil
				}
				err = json.Unmarshal(jFileBytes, &output)
				if err != nil {
					log.Fatalf("AWS account file %s, either was not valid JSON, or did not have the correct content. Please verify the file content and try again. Error: %v", aws_acct_file, err)
					return nil
				}
			}
		}
		if output == nil {
			output = &lib.AwsTargetCmdOutput{
				Accounts: map[string]*lib.AwsTargetCmdAccount{},
			}
		}
		tags = append(tags, fmt.Sprintf("Turbonomic-Host:%s", turbo_hostname))
		if iam_principal_create || iam_principal_delete {
			aws := clouds.Aws{
				Logger: log,
			}
			aws.SetCredentials(aws_access_key_id, aws_secret_access_key)

			log.Info("Querying org for list of child accounts...")

			accts, err := aws.ListAllOrgAccounts()
			if err != nil {
				msg := fmt.Sprintf("Unable to list accounts in the organization. Error: %v", err)
				output.Errors = append(output.Errors, msg)
				log.Fatal(msg)
				return nil
			}

			// Account Loop
			for _, acct := range accts {
				outputAcct := &lib.AwsTargetCmdAccount{Id: *acct.Id}
				output.Accounts[*acct.Id] = outputAcct
				iamCli := iam.New(aws.GetSession(), aws.GetConfig("us-east-1", *acct.Id, x_acct_role))
				acctLog := log.WithField("account", fmt.Sprintf("%v (%v)", *acct.Name, *acct.Id))
				acctLog.Infof("Assuming the role %v on account...", x_acct_role)

				// User flow
				if strings.ToLower(iam_principal_type) == "user" {
					userLog := acctLog.WithField("Username", iam_principal_name)
					userLog.Infof("Searching for users matching the username \"%s\" and/or tags \"%v\"", iam_principal_name, tags)
					matches, err := aws.FindMatchingUsers(*acct.Id, x_acct_role, iam_principal_name, tags)
					if err != nil {
						msg := fmt.Sprintf("Failed to query users in the account. Skipping actions in this account. Error: %v", err)
						outputAcct.Errors = append(outputAcct.Errors, msg)
						userLog.Errorf(msg)
						continue
					}
					if len(matches) > 0 {
						var matchesBuf bytes.Buffer
						for idx, match := range matches {
							matchesBuf.WriteString(fmt.Sprintf("%v: %v\n", idx, match.String()))
						}
						userLog.Infof("Matching users found\n%v", matchesBuf.String())
					}

					// User Create flow
					if iam_principal_create {
						outputAcct.Principal = &lib.AwsTargetCmdPrincipal{
							PrincipalType: "User",
							Name:          iam_principal_name,
						}
						userLog = userLog.WithField("action", "AddUser")
						// TODO: We're currently blocking only when the username matches.
						// Logic should be added to consider (re)using existing users which
						// are similar (some tags match).
						canNotProceed := false
						for _, match := range matches {
							if match.PrincipalnameMatch() {
								canNotProceed = true
							}
						}

						if canNotProceed {
							msg := fmt.Sprintf("A user with the username \"%s\" already exists. Duplicate users are not allowed. Please try a different username, or delete the existing user.", iam_principal_name)
							outputAcct.Principal.Errors = append(outputAcct.Principal.Errors, msg)
							userLog.Warnf(msg)
							continue
						} else {
							userLog.Info("Creating User")
							cuo, err := iamCli.CreateUser(&iam.CreateUserInput{UserName: &iam_principal_name, Tags: clouds.ConvertStringTagsPointer(tags)})
							if err != nil {
								msg := fmt.Sprintf("Failed to create user. Skipping actions in this account. Error: %v", err)
								outputAcct.Principal.Errors = append(outputAcct.Principal.Errors, msg)
								userLog.Errorf(msg)
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
									msg := fmt.Sprintf("Unable to attach policy (%v) to user. Continuing with remaining tasks. Error: %v", policy_arn, err)
									outputAcct.Principal.Errors = append(outputAcct.Principal.Errors, msg)
									userLog.Errorf(msg)
								} else {
									outputAcct.Principal.Policies = append(outputAcct.Principal.Policies, policyname)
								}
							}

							access_key_out, err := iamCli.CreateAccessKey(&iam.CreateAccessKeyInput{UserName: cuo.User.UserName})
							if err != nil {
								msg := fmt.Sprintf("Failed to create access key and secret for user. Error: %v", err)
								userLog.Errorf(msg)
								outputAcct.Principal.Errors = append(outputAcct.Principal.Errors, msg)
							} else {
								outputAcct.Principal.AccessKeyId = *access_key_out.AccessKey.AccessKeyId
								outputAcct.Principal.SecretAccessKey = *access_key_out.AccessKey.SecretAccessKey
								userLog.Info("User access key created.")
							}
						}
					} // User Create Flow

					// Delete flow
					if iam_principal_delete {
						userLog = userLog.WithField("action", "UserDelete")
						deleteUsers := matches
						if len(matches) > 0 {
							// TODO: Show the user the matches, and ask if they are sure!
							userLog.Warn(
								`Users which match either the exact username, all of the tags, or some subset have been found.

Enter the number(s) of the Users you would like to delete.

You can list several, separated by commas. I.E. 0, 1, 3. You can also simply type 'a' to delete all matched users.`)

							deleteChoices, err := stdinReader.ReadString('\n')
							if err != nil {
								userLog.Errorf("Failed to read the answer you provided. Users will not be deleted. Error: %v", err)
								continue // TODO: Does this bail out of the account entirely? Should it?
							}
							deleteChoices = strings.TrimRight(deleteChoices, "\n")
							if strings.ToLower(deleteChoices) != "a" {
								deleteUsers = []clouds.AwsPrincipalMatch{}
								for _, idxStr := range strings.Split(strings.ToLower(deleteChoices), ",") {
									idx, err := strconv.Atoi(idxStr)
									if err != nil {
										userLog.Errorf("One of your choices of users to delete was not a number. Tried to convert %v to a number. Users will not be deleted. Error: %v", idxStr, err)
										continue // TODO: Does this bail out of the account entirely? Should it?
									}
									if len(matches) > idx {
										deleteUsers = append(deleteUsers, matches[idx])
									}
								}
							}
							var delConfirmBuf bytes.Buffer
							for idx, delMatch := range deleteUsers {
								delConfirmBuf.WriteString(fmt.Sprintf("%v: %v\n", idx, delMatch.String()))
							}
							userLog.Warnf("The following users will be deleted. THIS CAN NOT BE UNDONE!!\n%v\n\nAre you very, very, very, VERY sure? (y/n)", delConfirmBuf.String())
							answer, _, err := stdinReader.ReadRune()
							if err != nil {
								userLog.Errorf("Failed to read the answer you provided. Users will not be deleted. Error: %v", err)
								continue // TODO: Does this bail out of the account entirely? Should it?
							}
							if unicode.ToLower(answer) == 'y' {
								for _, deletePrincipal := range deleteUsers {
									deleteUser, err := deletePrincipal.AsUser()
									if err != nil {
										userLog.Errorf("This user match wasn't a user at all. Something has gone horribly wrong. Error: %v", err)
										continue
									}
									userLog = userLog.WithField("Username", *deleteUser.UserName)
									userLog.Debugf("Searching for managed policies associated with user %s", *deleteUser.UserName)
									userReadyForDelete := true
									err = iamCli.ListAttachedUserPoliciesPages(&iam.ListAttachedUserPoliciesInput{UserName: deleteUser.UserName}, func(arg1 *iam.ListAttachedUserPoliciesOutput, arg2 bool) bool {
										for _, policy := range arg1.AttachedPolicies {
											userLog.Debugf("Deleting managed role (%v) from user %s", *policy.PolicyName, *deleteUser.UserName)
											_, err = iamCli.DetachUserPolicy(&iam.DetachUserPolicyInput{UserName: deleteUser.UserName, PolicyArn: policy.PolicyArn})
											if err != nil {
												userLog.Errorf("Failed to delete managed policy (%v) associated with user %s. User will not be deleted. Error: %v", *policy.PolicyArn, *deleteUser.UserName, err)
												userReadyForDelete = false
												return false
											}
										}
										return true
									})
									if err != nil {
										userLog.Errorf("Failed to list policies associated with user %s. User will not be deleted. Error: %v", *deleteUser.UserName, err)
										continue
									}

									err = iamCli.ListAccessKeysPages(&iam.ListAccessKeysInput{UserName: deleteUser.UserName}, func(arg1 *iam.ListAccessKeysOutput, arg2 bool) bool {
										for _, key := range arg1.AccessKeyMetadata {
											userLog.Debugf("Deleting access key (%v) from user %s", *key.AccessKeyId, *deleteUser.UserName)
											_, err = iamCli.DeleteAccessKey(&iam.DeleteAccessKeyInput{AccessKeyId: key.AccessKeyId, UserName: deleteUser.UserName})
											if err != nil {
												userLog.Errorf("Failed to delete access key (%v) associated with user %s. User will not be deleted. Error: %v", *key.AccessKeyId, *deleteUser.UserName, err)
												userReadyForDelete = false
												return false
											}
										}
										return true
									})
									if err != nil {
										userLog.Errorf("Failed to list access keys associated with user %s. User will not be deleted. Error: %v", *deleteUser.UserName, err)
										continue
									}

									if !userReadyForDelete {
										continue
									}

									_, err = iamCli.DeleteUser(&iam.DeleteUserInput{UserName: deleteUser.UserName})
									if err != nil {
										userLog.Errorf("Failed to delete user %s. Error: %v", *deleteUser.UserName, err)
									}
								}
							} else if unicode.ToLower(answer) == 'n' {
								userLog.Info("You answered no. Users will not be deleted")
								continue
							} else {
								userLog.Errorf("Did not understand your answer. Expected y or n, you provided %v. Users will not be deleted.", answer)
								continue // TODO: Does this bail out of the account entirely? Should it?
							}
						} else {
							userLog.Warn("No users matching desired username and tags were found to delete")
						}
					}
				} // User Flow

				// Role flow
				if strings.ToLower(iam_principal_type) == "role" {
					roleLog := acctLog.WithField("Role", iam_principal_name)
					roleLog.Infof("Searching for roles matching the rolename \"%s\" and/or tags \"%v\"", iam_principal_name, tags)
					matches, err := aws.FindMatchingRoles(*acct.Id, x_acct_role, iam_principal_name, tags)
					if err != nil {
						roleLog.Errorf("Failed to query roles in the account. Skipping actions in this account. Error: %v", err)
						continue
					}

					// Role create flow
					if iam_principal_create {
						roleLog = roleLog.WithField("action", "RoleAdd")
						if len(matches) > 0 { // Existing role

						} else { // END Existing Role - START Role does not already exist
							roleLog.Info("Creating Role")
							assumeRolePolicyDocument := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com",
        "AWS": "arn:aws:sts::%s:assumed-role/%s/%s"
      },
      "Action": "sts:AssumeRole",
      "Condition": {}
    }
  ]
}
`

							assumeRolePolicyDocument = fmt.Sprintf(assumeRolePolicyDocument, turbo_trusted_account_id, turbo_trusted_account_role, turbo_trusted_account_instanceid)
							roleLog.Debugf("Role trust relationship is defined as;\n%v", assumeRolePolicyDocument)
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
					} // Role create Flow

					if iam_principal_delete { // Role delete flow

					} // Role delete flow
				} // Role Flow
			} // Account Loop
		} else {
			log.Info("No IAM principal create, nor delete was requested. Skipping the AWS org account list request.")
		}
		if len(aws_acct_file) > 0 {
			outputJson, err := json.MarshalIndent(output, "", " ")
			if err != nil {
				log.Errorf("Unable to marshal results into AWS account file. Error: %v", err)
				return nil
			}

			err = ioutil.WriteFile(aws_acct_file, outputJson, 0644)
			if err != nil {
				log.Errorf("Unable to write AWS account file to disk. Error: %v", err)
				return nil
			}
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
	awsCmd.Flags().StringVar(&turbo_trusted_account_id, "turbo-trusted-account-id", "", "The AWS account id where the Turbonomic OpsMgr is running.")
	awsCmd.Flags().StringVar(&turbo_trusted_account_role, "turbo-trusted-account-role", "", "The AWS role name the Turbonomic OpsMgr is assuming.")
	awsCmd.Flags().StringVar(&turbo_trusted_account_instanceid, "turbo-trusted-account-instanceid", "", "The AWS EC2 instance id of Turbonomic OpsMgr.")

	// Move these to a parent, or the root most likely
	awsCmd.Flags().StringVar(&turbo_hostname, "turbo-hostname", "", "The host or ip address of your Turbonomic OpsMgr.")
	awsCmd.Flags().StringVar(&turbo_username, "turbo-username", "", "The username of an administrator on your Turbonomic OpsMgr.")
	awsCmd.Flags().StringVar(&turbo_password, "turbo-password", "", "The password of an administrator on your Turbonomic OpsMgr.")
	awsCmd.Flags().BoolVarP(&turbo_target_create, "turbo-target-create", "", false, "When set, the Turbonomic targets will be created.")
	awsCmd.Flags().BoolVarP(&turbo_target_delete, "turbo-target-delete", "", false, "When set, the Turbonomic targets will be deleted.")
	awsCmd.Flags().StringVar(&turbo_target_prefix, "turbo-target-prefix", "", "A prefix to use on the name of targets created by this tool.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// awsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// awsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
