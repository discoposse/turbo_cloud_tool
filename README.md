# Turbo Cloud Tool
A golang library and CLI providing several useful utilities for dealing with cloud providers, and the Turbonomic OpsMgr.

Notably, being able to properly setup Turbnomic [targets](https://turbonomic.com/wp-content/uploads/2019/03/TargetConfiguration_6.3.1.pdf#page=39&zoom=100,0,273) in a consistent and automated fashion.

# Table of contents
* [Turbo Cloud Tool](#turbo-cloud-tool)
* [Sub commands](#sub-commands)
  * [AWS](#aws)
     * [IAM Principals](#iam-principals)
     * [Turbonomic Target](#turbonomic-target)
     * [Common use cases](#common-use-cases)
        * [Create IAM Users &amp; Turbonomic Targets](#create-iam-users--turbonomic-targets)
        * [Create IAM Roles &amp; Turbonomic Targets](#create-iam-roles--turbonomic-targets)
        * [(Only) Create IAM Users](#only-create-iam-users)
        * [(Only) Create IAM Roles](#only-create-iam-roles)
        * [(Only) Create Turbonomic Targets](#only-create-turbonomic-targets)
* [TODO](#todo)
* [Examples](#examples)
  * [Bulk add targets for IAM roles](#bulk-add-targets-for-iam-roles)
  * [Bulk add targets for IAM users](#bulk-add-targets-for-iam-users)
  * [Delete existing IAM users, with confirmation](#delete-existing-iam-users-with-confirmation)
* [Feature Ideas](#feature-ideas)

# Sub commands
The turbo cloud tool will eventually have several sub commands. For now, there is only one.

## AWS
This command can be run in several different configurations depending on the flags that you pass into it when invoking it.

The primary goal of this command is to automate the task of adding a cloud target to Turbonomic for every account in your AWS org.

Before we get into common use cases, there are some concepts you should be familiar with.

A Turbonomic target can be created with one of the following authentication methods.

1. An AWS Access Key Id and Secret Access Key associated with a user - [Docs Link](https://greencircle.vmturbo.com/docs/DOC-3828-connecting-turbonomic-to-amazon-web-services-aws)
2. An AWS Role ARN - [Docs Link](https://greencircle.vmturbo.com/docs/DOC-6176)

Option #2 is only available when the Turbonomic OpsMgr is running in AWS.

The Turbo Cloud Tool can create either type of IAM principal and add it to Turbonomic as a target.

When the Turbo Cloud Tool creates an IAM principal, and successfully adds it as a Turbonomic target, the IAM principal will have at least the following three tags.

1. Owner:Turbonomic
2. Turbonomic-Host:<hostname or IP of the Turbonomic OpsMgr>
3. Turbonomic-Target-Uuid:<unique ID of the Turbonomic target>

You can specify additional tags with the `--tag` option.

### Usage
```
./turbo_cloud_tool aws -h
Automates the creation of IAM principals, and Turbonomic targets for all accounts in your AWS org.

This command can be invoked in SEVERAL different ways with different results. It is recommended that you read the documentation before using this command.
https://github.com/turbonomiclabs/turbo_cloud_tool#aws

Usage:
  turbo_cloud_tool aws [flags]

Flags:
      --aws-access-key-id string                  An AWS Access Key ID for a user with AWSOrganizationsReadOnlyAccess and the
                                                  ability to assume a privileged role in each account in the org
      --aws-account-file string                   A filename containing an AWS account list, with optional IAM principal details
                                                  (AWS Access Key IDs and secrets).

                                                  If set in combination with --iam-principal-create, the child accounts that are
                                                  found, and the IAM principals which are created will be stored in this file.

                                                  If set in combination with --turbo-target-create this file will be used as the
                                                  source of target data.
      --aws-secret-access-key string              An AWS Secret Access Key for a user with AWSOrganizationsReadOnlyAccess and the
                                                  ability to assume a privileged role in each account in the org
  -h, --help                                      help for aws
      --iam-principal-create                      When set, the IAM principals will be created.
      --iam-principal-delete                      When set, the IAM principals matching the settings
                                                  (--iam-principal-name --iam-principal-type, and --tag) will be deleted, after verification.
      --iam-principal-name string                 The name of the IAM principal (User or Role, depending upon --iam-principal-type)
                                                  which should be created. (default "Turbonomic-OpsMgr-Target")
      --iam-principal-type string                 One of either 'role' or 'user', indicating the type of authentication to use for the Turbo target.
      --permit-automation                         When set, the IAM principals will be created with the policies necessary for
                                                  Turbonomic to automate.
      --tag strings                               One or many tags to add to the IAM principal which is created. Must be in the
                                                  form of 'Key:Value'. For instance to add an Owner tag with the value 'Sales' you
                                                  would use --tag Owner:Sales
                                                   (default [Owner:Turbonomic,Turbonomic-Host:<hostname or IP of the Turbonomic OpsMgr>,Turbonomic-Target-Uuid:<unique ID of the Turbonomic target>])
      --turbo-hostname string                     The host or ip address of your Turbonomic OpsMgr.
      --turbo-password string                     The password of an administrator on your Turbonomic OpsMgr.
      --turbo-target-create                       When set, the Turbonomic targets will be created.
      --turbo-target-delete                       When set, the Turbonomic targets will be deleted.
      --turbo-target-prefix string                A prefix to use on the name of targets created by this tool.
      --turbo-trusted-account-id string           The AWS account id where the Turbonomic OpsMgr is running.
      --turbo-trusted-account-instanceid string   The AWS EC2 instance id of Turbonomic OpsMgr.
      --turbo-trusted-account-role string         The AWS role name the Turbonomic OpsMgr is assuming.
      --turbo-username string                     The username of an administrator on your Turbonomic OpsMgr.
      --x-acct-role string                        The name of a cross account role which has privileges to create the IAM principal
                                                  in each account. (default "OrganizationAccountAccessRole")

Global Flags:
      --config string   config file (default is $HOME/.turbo_cloud_tool.yaml)
  -v, --verbose         Enables trace logging for verbose output.
  ```

### Common use cases

#### Create IAM Users & Turbonomic Targets
##### Prerequisites

#### Create IAM Roles & Turbonomic Targets

#### (Only) Create IAM Users

#### (Only) Create IAM Roles

#### (Only) Create Turbonomic Targets

# TODO
* General
  * ~Rename appropriately (cloud_target_tool?)~
  * Build/release automation which creates binaries for each arch/platform (OSX, Windows, Linux)
  * Better encapsulation
  * Struct/logic for handling stdin. Lots of duplicated code here, need a standardized reusable mechanism.
  * Ability to run idempotently, such that it could be run periodically to always add targets for new accounts.
* Azure - On hold until 6.4 and/or more discovery
* AWS
  * Separate the account file "output" from the account file "input".
    * When doing a delete, it would be good to delete the *exact* objects which were created, but the current workflow clobbers the necessary information. Need to have better logic for create/delete, or just have different files for each.
  * User Delete
    * "Force" option with clear documentation
    * ~Complete confirmation flow before deleting~
  * User "Rotate Credentials"
    * The ability to find existing users, create new credentials, add them to an existing Turbo target, and delete the old credentials.
  * Role Create
    * "Force" option with clear documentation
    * ~Bail when role already exists~
  * Role Delete
    * "Force" option with clear documentation
    * Delete all policies before deleting role
    * Complete confirmation flow before deleting
  * Turbo Target Create
    * "Force" option with clear documentation
    * ~From "in memory" principals created~
    * From principals created and stored in AWS acct file
      * Likely implemented, needs testing
    * ~Go back and tag principal with Target UUID~
  * Input validation (ensure that users provide all necessary inputs with useful error/warning/feedback)
  * Success/Fail summary
  * Separate commands to parse/rationalize the account file?
  * ~Create tags on principals with provided case (currently all tags are lowercase)~
  * Allow "Updating" of any given principal to add the correct policies, etc.

# Examples

## Bulk add targets for IAM roles
```
./turbo_cloud_tool aws --aws-access-key-id $AWS_ACCESS_KEY_ID --aws-secret-access-key "$AWS_SECRET_ACCESS_KEY" --iam-principal-type role --iam-principal-name Boo --turbo-hostname localhost --turbo-username "$TURBO_USERNAME" --turbo-password "$TURBO_PASSWORD" --iam-principal-create -v --aws-account-file output.json --turbo-trusted-account-id 385266030856 --turbo-trusted-account-role TurboSEEnablementXAccountTrust --turbo-trusted-account-instanceid i-097ec442cd152a92b
Jul 31 13:21:29.564 [INFO] Using config file: /Users/ryangeyer/.turbo_cloud_tool.yaml
Jul 31 13:21:29.564 [INFO] Querying org for list of child accounts...
Jul 31 13:21:30.367 [INFO] [TurboBlankAccount (905994805379)] Assuming the role OrganizationAccountAccessRole on account...
Jul 31 13:21:30.367 [INFO] [TurboBlankAccount (905994805379)] [Boo] Searching for roles matching the rolename "Boo" and/or tags "[Owner:Turbonomic Turbonomic-Host:localhost]"
Jul 31 13:21:32.470 [INFO] [TurboBlankAccount (905994805379)] [Boo] Matching Roles found
0: [Match Type: Exact] - [Name: Boo] - [Tags: owner:turbonomic, turbonomic-host:localhost] - [Principal Type: Role]
1: [Match Type: All Tags] - [Name: Foo] - [Tags: owner:turbonomic, turbonomic-host:localhost] - [Principal Type: Role]

Jul 31 13:21:32.470 [WARN] [TurboBlankAccount (905994805379)] [AddRole] [Boo] A role with the username "Boo" already exists. Duplicate roles are not allowed. Please try a different rolename, or delete the existing role.
Jul 31 13:21:32.470 [INFO] [Ryan J. Geyer (385266030856)] Assuming the role OrganizationAccountAccessRole on account...
Jul 31 13:21:32.470 [INFO] [Ryan J. Geyer (385266030856)] [Boo] Searching for roles matching the rolename "Boo" and/or tags "[Owner:Turbonomic Turbonomic-Host:localhost]"
Jul 31 13:21:32.629 [ERRO] [Ryan J. Geyer (385266030856)] [Boo] Failed to query roles in the account. Skipping actions in this account. Error: AccessDenied: Access denied
	status code: 403, request id: 86b3078a-b3d0-11e9-be54-51388ff0eaa9
```

## Bulk add targets for IAM users
```
/turbo_cloud_tool aws --aws-access-key-id $AWS_ACCESS_KEY_ID --aws-secret-access-key "$AWS_SECRET_ACCESS_KEY" --iam-principal-type user --iam-principal-name Foo --turbo-hostname localhost --turbo-username "$TURBO_USERNAME" --turbo-password "$TURBO_PASSWORD" --iam-principal-create -v --aws-account-file output.json
Jul 31 13:23:45.986 [INFO] Using config file: /Users/ryangeyer/.turbo_cloud_tool.yaml
Jul 31 13:23:45.987 [INFO] Querying org for list of child accounts...
Jul 31 13:23:46.564 [INFO] [TurboBlankAccount (905994805379)] Assuming the role OrganizationAccountAccessRole on account...
Jul 31 13:23:46.564 [INFO] [TurboBlankAccount (905994805379)] [Foo] Searching for users matching the username "Foo" and/or tags "[Owner:Turbonomic Turbonomic-Host:localhost]"
Jul 31 13:23:47.507 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Foo] Creating User
Jul 31 13:23:47.647 [DEBU] [TurboBlankAccount (905994805379)] [AddUser] [Foo] User Created
Jul 31 13:23:47.647 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Foo] Adding Polices to user which allow Turbonomic read only access. Policies: [AmazonEC2ReadOnlyAccess AmazonS3ReadOnlyAccess AmazonRDSReadOnlyAccess]
Jul 31 13:23:47.647 [DEBU] [TurboBlankAccount (905994805379)] [AddUser] [Foo] Attaching policy (arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess) to user
Jul 31 13:23:47.774 [DEBU] [TurboBlankAccount (905994805379)] [AddUser] [Foo] Attaching policy (arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess) to user
Jul 31 13:23:47.904 [DEBU] [TurboBlankAccount (905994805379)] [AddUser] [Foo] Attaching policy (arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess) to user
Jul 31 13:23:48.171 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Foo] User access key created.
Jul 31 13:23:48.171 [INFO] [Ryan J. Geyer (385266030856)] Assuming the role OrganizationAccountAccessRole on account...
Jul 31 13:23:48.171 [INFO] [Ryan J. Geyer (385266030856)] [Foo] Searching for users matching the username "Foo" and/or tags "[Owner:Turbonomic Turbonomic-Host:localhost]"
Jul 31 13:23:48.281 [ERRO] [Ryan J. Geyer (385266030856)] [Foo] Failed to query users in the account. Skipping actions in this account. Error: AccessDenied: Access denied
	status code: 403, request id: d78d459c-b3d0-11e9-928b-ffe8b6eb55e8
```

## Delete existing IAM users, with confirmation
```
./turbo_cloud_tool aws --aws-access-key-id $AWS_ACCESS_KEY_ID --aws-secret-access-key "$AWS_SECRET_ACCESS_KEY" --iam-principal-type user --iam-principal-name Foo --turbo-hostname localhost --turbo-username "$TURBO_USERNAME" --turbo-password "$TURBO_PASSWORD" --iam-principal-delete -v --aws-account-file output.json
Jul 31 13:24:27.572 [INFO] Using config file: /Users/ryangeyer/.turbo_cloud_tool.yaml
Jul 31 13:24:27.572 [INFO] Querying org for list of child accounts...
Jul 31 13:24:28.168 [INFO] [TurboBlankAccount (905994805379)] Assuming the role OrganizationAccountAccessRole on account...
Jul 31 13:24:28.168 [INFO] [TurboBlankAccount (905994805379)] [Foo] Searching for users matching the username "Foo" and/or tags "[Owner:Turbonomic Turbonomic-Host:localhost]"
Jul 31 13:24:29.358 [INFO] [TurboBlankAccount (905994805379)] [Foo] Matching users found
0: [Match Type: All Tags] - [Name: Bar] - [Tags: owner:turbonomic, turbonomic-host:localhost] - [Principal Type: User]
1: [Match Type: Exact] - [Name: Foo] - [Tags: owner:turbonomic, turbonomic-host:localhost] - [Principal Type: User]

Jul 31 13:24:29.358 [WARN] [TurboBlankAccount (905994805379)] [UserDelete] [Foo] Users which match either the exact username, all of the tags, or some subset have been found.

Enter the number(s) of the Users you would like to delete.

You can list several, separated by commas. I.E. 0, 1, 3. You can also simply type 'a' to delete all matched users.
a
Jul 31 13:24:31.452 [WARN] [TurboBlankAccount (905994805379)] [UserDelete] [Foo] The following users will be deleted. THIS CAN NOT BE UNDONE!!
0: [Match Type: All Tags] - [Name: Bar] - [Tags: owner:turbonomic, turbonomic-host:localhost] - [Principal Type: User]
1: [Match Type: Exact] - [Name: Foo] - [Tags: owner:turbonomic, turbonomic-host:localhost] - [Principal Type: User]


Are you very, very, very, VERY sure? (y/n)
y
Jul 31 13:24:32.493 [DEBU] [TurboBlankAccount (905994805379)] [UserDelete] [Bar] Searching for managed policies associated with user Bar
Jul 31 13:24:32.607 [DEBU] [TurboBlankAccount (905994805379)] [UserDelete] [Bar] Deleting managed role (AmazonEC2ReadOnlyAccess) from user Bar
Jul 31 13:24:32.745 [DEBU] [TurboBlankAccount (905994805379)] [UserDelete] [Bar] Deleting managed role (AmazonS3ReadOnlyAccess) from user Bar
Jul 31 13:24:32.866 [DEBU] [TurboBlankAccount (905994805379)] [UserDelete] [Bar] Deleting managed role (AmazonRDSReadOnlyAccess) from user Bar
Jul 31 13:24:33.109 [DEBU] [TurboBlankAccount (905994805379)] [UserDelete] [Bar] Deleting access key (AKIA5F4L7UCBTSDRWVU2) from user Bar
Jul 31 13:24:33.341 [DEBU] [TurboBlankAccount (905994805379)] [UserDelete] [Foo] Searching for managed policies associated with user Foo
Jul 31 13:24:33.475 [DEBU] [TurboBlankAccount (905994805379)] [UserDelete] [Foo] Deleting managed role (AmazonEC2ReadOnlyAccess) from user Foo
Jul 31 13:24:33.600 [DEBU] [TurboBlankAccount (905994805379)] [UserDelete] [Foo] Deleting managed role (AmazonS3ReadOnlyAccess) from user Foo
Jul 31 13:24:33.720 [DEBU] [TurboBlankAccount (905994805379)] [UserDelete] [Foo] Deleting managed role (AmazonRDSReadOnlyAccess) from user Foo
Jul 31 13:24:34.018 [DEBU] [TurboBlankAccount (905994805379)] [UserDelete] [Foo] Deleting access key (AKIA5F4L7UCBR4L6DYWS) from user Foo
Jul 31 13:24:34.259 [INFO] [Ryan J. Geyer (385266030856)] Assuming the role OrganizationAccountAccessRole on account...
Jul 31 13:24:34.259 [INFO] [Ryan J. Geyer (385266030856)] [Foo] Searching for users matching the username "Foo" and/or tags "[Owner:Turbonomic Turbonomic-Host:localhost]"
Jul 31 13:24:34.368 [ERRO] [Ryan J. Geyer (385266030856)] [Foo] Failed to query users in the account. Skipping actions in this account. Error: AccessDenied: Access denied
	status code: 403, request id: f305be82-b3d0-11e9-8bac-d5ea1c24e18e
```

# Feature Requests

* Validate that a given cloud account can be used.
  * Azure
    * Can download ratecard
    * Microsoft.Compute is registered (https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-manager-register-provider-errors#)
    * Other things?
  * AWS
    * Org exists, and these credentials have access?
    * Turbonomic OpsMgr Instance running?
* Provision/deploy OpsMgr in either AWS or Azure
* Discover running OpsMgr in any subscription/account of any cloud
  * If found in AWS, recommend use of role based authentication
* Create (and associate?) IAM instance role for role based authentication
* Automate enabling memory metrics in either AWS or Azure
* Automate creating & registering cost and usage report for AWS
