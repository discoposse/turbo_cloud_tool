# Turbo Cloud Tool
A golang library and CLI providing several useful utilities for dealing with cloud providers, and the Turbonomic OpsMgr.

Notably, being able to properly setup Turbnomic [targets](https://turbonomic.com/wp-content/uploads/2019/03/TargetConfiguration_6.3.1.pdf#page=39&zoom=100,0,273) in a consistent and automated fashion.

# Table of contents
* [Turbo Cloud Tool](#turbo-cloud-tool)
* [Table of contents](#table-of-contents)
* [Sub commands](#sub-commands)
  * [AWS](#aws)
     * [Usage](#usage)
     * [Common use cases](#common-use-cases)
        * [Create IAM Users &amp; Turbonomic Targets](#create-iam-users--turbonomic-targets)
           * [Prerequisites](#prerequisites)
           * [Example](#example)
        * [Create IAM Roles &amp; Turbonomic Targets](#create-iam-roles--turbonomic-targets)
           * [Prerequisites](#prerequisites-1)
           * [Example](#example-1)
        * [(Only) Create IAM Users](#only-create-iam-users)
           * [Prerequisites](#prerequisites-2)
           * [Example](#example-2)
        * [(Only) Create IAM Roles](#only-create-iam-roles)
           * [Prerequisites](#prerequisites-3)
           * [Example](#example-3)
        * [(Only) Create Turbonomic Targets](#only-create-turbonomic-targets)
* [TODO](#todo)
* [Examples](#examples)
  * [Delete existing IAM users, with confirmation](#delete-existing-iam-users-with-confirmation)
* [Feature Requests](#feature-requests)

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
There are some standard (and expected) ways that you might use this tool. Those are documented below, starting with the two most common use cases. Namely, automatically creating either user based targets, or role based targets in Turbonomic for every AWS account in your organization.

#### Create IAM Users & Turbonomic Targets
This use case will iterate over every account in your AWS org.

In each account, a new IAM user will be created with the following details
* The appropriate policies for either read-only or automation access from the Turbonomic OpsMgr will be assigned to the user
* A new access key will be generated

If an error is encountered, the account will be skipped.

Once all accounts have been iterated, a Turbonomic target will be created for each account where a user was successfully created.

##### Prerequisites
* The name of a role which exist in each account with permissions to create IAM principals. By default this is the "OrganizationAccountAccessRole" `--x-acct-role`
* AWS Access Key Id and Secret of a user with permissions to list all accounts in your org, and assume the role above. `--aws-access-key-id` and `--aws-secret-access-key`
* Hostname, username, and password for a Turbonomic instance where the targets will be created. `--turbo-hostname`, `--turbo-username`, and `--turbo-password`

##### Example
```
./turbo_cloud_tool aws --aws-access-key-id $AWS_ACCESS_KEY_ID --aws-secret-access-key "$AWS_SECRET_ACCESS_KEY" --turbo-hostname localhost --turbo-username "$TURBO_USERNAME" --turbo-password "$TURBO_PASSWORD" --iam-principal-create --turbo-target-create --iam-principal-type user
Aug  2 13:07:58.270 [INFO] Using config file: /Users/ryangeyer/.turbo_cloud_tool.yaml
Aug  2 13:07:58.270 [INFO] Querying org for list of child accounts...
Aug  2 13:07:59.031 [INFO] [TurboBlankAccount (905994805379)] Assuming the role OrganizationAccountAccessRole on account...
Aug  2 13:07:59.031 [INFO] [TurboBlankAccount (905994805379)] [Turbonomic-OpsMgr-Target] Searching for users matching the username "Turbonomic-OpsMgr-Target" and/or tags "[Owner:Turbonomic Turbonomic-Host:msse01.demo.turbonomic.com]"
Aug  2 13:07:59.858 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] Creating User
Aug  2 13:07:59.980 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] User Created
Aug  2 13:07:59.980 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] Adding Polices to user which allow Turbonomic read only access. Policies: [AmazonEC2ReadOnlyAccess AmazonS3ReadOnlyAccess AmazonRDSReadOnlyAccess]
Aug  2 13:07:59.980 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess) to user
Aug  2 13:08:00.093 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess) to user
Aug  2 13:08:00.206 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess) to user
Aug  2 13:08:00.429 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] User access key created.
Aug  2 13:08:00.429 [INFO] [Ryan J. Geyer (385266030856)] Assuming the role OrganizationAccountAccessRole on account...
Aug  2 13:08:00.429 [INFO] [Ryan J. Geyer (385266030856)] [Turbonomic-OpsMgr-Target] Searching for users matching the username "Turbonomic-OpsMgr-Target" and/or tags "[Owner:Turbonomic Turbonomic-Host:msse01.demo.turbonomic.com]"
Aug  2 13:08:01.666 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] Creating User
Aug  2 13:08:01.786 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] User Created
Aug  2 13:08:01.786 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] Adding Polices to user which allow Turbonomic read only access. Policies: [AmazonEC2ReadOnlyAccess AmazonS3ReadOnlyAccess AmazonRDSReadOnlyAccess]
Aug  2 13:08:01.786 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess) to user
Aug  2 13:08:01.900 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess) to user
Aug  2 13:08:02.010 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess) to user
Aug  2 13:08:02.233 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] User access key created.
Aug  2 13:08:12.235 [INFO] [Ryan J. Geyer (385266030856)] [AddTurboTarget] Creating Turbo Target
Aug  2 13:08:12.235 [INFO] [Ryan J. Geyer (385266030856)] [AddTurboTarget] Creating AWS target Ryan J. Geyer from IAM user credentials.
Aug  2 13:08:14.289 [INFO] [Ryan J. Geyer (385266030856)] [AddTurboTarget] Turbo Target Created
Aug  2 13:08:14.289 [INFO] [Ryan J. Geyer (385266030856)] [AddTurboTarget] Tagging IAM principal Turbonomic-OpsMgr-Target with Turbo target uuid _A-aY0LVhEemJgbBmfZupWQ
Aug  2 13:08:14.506 [INFO] [TurboBlankAccount (905994805379)] [AddTurboTarget] Creating Turbo Target
Aug  2 13:08:14.506 [INFO] [TurboBlankAccount (905994805379)] [AddTurboTarget] Creating AWS target TurboBlankAccount from IAM user credentials.
Aug  2 13:08:15.577 [INFO] [TurboBlankAccount (905994805379)] [AddTurboTarget] Turbo Target Created
Aug  2 13:08:15.577 [INFO] [TurboBlankAccount (905994805379)] [AddTurboTarget] Tagging IAM principal Turbonomic-OpsMgr-Target with Turbo target uuid _BPcecLVhEemJgbBmfZupWQ
```
#### Create IAM Roles & Turbonomic Targets
This use case will iterate over every account in your AWS org.

In each account, a new IAM role will be created with the appropriate policies for either read-only or automation access from the Turbonomic OpsMgr.

If an error is encountered, the account will be skipped.

Once all accounts have been iterated, a Turbonomic target will be created for each account where a role was successfully created.

##### Prerequisites
* The name of a role which exist in each account with permissions to create IAM principals. By default this is the "OrganizationAccountAccessRole" `--x-acct-role`
* AWS Access Key Id and Secret of a user with permissions to list all accounts in your org, and assume the role above. `--aws-access-key-id` and `--aws-secret-access-key`
* Hostname, username, and password for a Turbonomic instance where the targets will be created. `--turbo-hostname`, `--turbo-username`, and `--turbo-password`
* The AWS account ID where the Turbonomic OpsMgr is running `--turbo-trusted-account-id`
* The AWS IAM role assumed by the Turbonomic OpsMgr `--turbo-tursted-account-role`
* The AWS EC2 Instance id of the Turbonomic OpsMgr `--turbo-trusted-account-instanceid`

##### Example
```
./turbo_cloud_tool aws --aws-access-key-id $AWS_ACCESS_KEY_ID --aws-secret-access-key "$AWS_SECRET_ACCESS_KEY" --turbo-hostname localhost --turbo-username "$TURBO_USERNAME" --turbo-password "$TURBO_PASSWORD" --iam-principal-create --turbo-target-create --iam-principal-type role --turbo-trusted-account-id 385266030856 --turbo-trusted-account-role TurboSEEnablementXAccountTrust --turbo-trusted-account-instanceid i-097ec442cd152a92b
Aug  2 13:39:08.003 [INFO] Using config file: /Users/ryangeyer/.turbo_cloud_tool.yaml
Aug  2 13:39:08.004 [INFO] Querying org for list of child accounts...
Aug  2 13:39:08.670 [INFO] [TurboBlankAccount (905994805379)] Assuming the role OrganizationAccountAccessRole on account...
Aug  2 13:39:08.670 [INFO] [TurboBlankAccount (905994805379)] [Turbonomic-OpsMgr-Target] Searching for roles matching the rolename "Turbonomic-OpsMgr-Target" and/or tags "[Owner:Turbonomic Turbonomic-Host:msse01.demo.turbonomic.com]"
Aug  2 13:39:10.068 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Creating Role
Aug  2 13:39:10.202 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Role Created
Aug  2 13:39:10.203 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Adding Polices to role which allow Turbonomic read only access. Policies: [AmazonEC2ReadOnlyAccess AmazonS3ReadOnlyAccess AmazonRDSReadOnlyAccess]
Aug  2 13:39:10.203 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess
Aug  2 13:39:10.312 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
Aug  2 13:39:10.422 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess
Aug  2 13:39:10.534 [INFO] [Ryan J. Geyer (385266030856)] Assuming the role OrganizationAccountAccessRole on account...
Aug  2 13:39:10.534 [INFO] [Ryan J. Geyer (385266030856)] [Turbonomic-OpsMgr-Target] Searching for roles matching the rolename "Turbonomic-OpsMgr-Target" and/or tags "[Owner:Turbonomic Turbonomic-Host:msse01.demo.turbonomic.com]"
Aug  2 13:39:12.614 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Creating Role
Aug  2 13:39:12.752 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Role Created
Aug  2 13:39:12.752 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Adding Polices to role which allow Turbonomic read only access. Policies: [AmazonEC2ReadOnlyAccess AmazonS3ReadOnlyAccess AmazonRDSReadOnlyAccess]
Aug  2 13:39:12.752 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess
Aug  2 13:39:12.865 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
Aug  2 13:39:12.978 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess
Aug  2 13:39:23.090 [INFO] [Ryan J. Geyer (385266030856)] [AddTurboTarget] Creating Turbo Target
Aug  2 13:39:23.091 [INFO] [Ryan J. Geyer (385266030856)] [AddTurboTarget] Creating AWS target Ryan J. Geyer from IAM Role Arn.
Aug  2 13:39:23.973 [INFO] [Ryan J. Geyer (385266030856)] [AddTurboTarget] Turbo Target Created
Aug  2 13:39:23.973 [INFO] [Ryan J. Geyer (385266030856)] [AddTurboTarget] Tagging IAM principal Turbonomic-OpsMgr-Target with Turbo target uuid _A-aY0LVhEemJgbBmfZupWQ
Aug  2 13:39:23.973 [INFO] [TurboBlankAccount (905994805379)] [AddTurboTarget] Creating Turbo Target
Aug  2 13:39:23.973 [INFO] [TurboBlankAccount (905994805379)] [AddTurboTarget] Creating AWS target TurboBlankAccount from IAM Role Arn.
Aug  2 13:39:24.587 [INFO] [TurboBlankAccount (905994805379)] [AddTurboTarget] Turbo Target Created
Aug  2 13:39:24.587 [INFO] [TurboBlankAccount (905994805379)] [AddTurboTarget] Tagging IAM principal Turbonomic-OpsMgr-Target with Turbo target uuid _A-aY0LVhEemJgbBmfZupWQ
```

#### (Only) Create IAM Users
This use case will iterate over every account in your AWS org.

In each account, a new IAM user will be created with the following details
* The appropriate policies for either read-only or automation access from the Turbonomic OpsMgr will be assigned to the user
* A new access key will be generated

If an error is encountered, the account will be skipped.

This should only be used with the `--aws-account-file` flag, otherwise the users will be created, but there will be no record of the user credentials. This file can then later be used with different flags to add the [Turbonomic targets](#only-create-turbonomic-targets).

##### Prerequisites
* The name of a role which exist in each account with permissions to create IAM principals. By default this is the "OrganizationAccountAccessRole" `--x-acct-role`
* AWS Access Key Id and Secret of a user with permissions to list all accounts in your org, and assume the role above. `--aws-access-key-id` and `--aws-secret-access-key`

##### Example
```
./turbo_cloud_tool aws --aws-access-key-id $AWS_ACCESS_KEY_ID --aws-secret-access-key "$AWS_SECRET_ACCESS_KEY" --iam-principal-create --iam-principal-type user --aws-account-file output.json
Aug  2 13:52:55.227 [INFO] Using config file: /Users/ryangeyer/.turbo_cloud_tool.yaml
Aug  2 13:52:55.228 [INFO] Querying org for list of child accounts...
Aug  2 13:52:56.199 [INFO] [TurboBlankAccount (905994805379)] Assuming the role OrganizationAccountAccessRole on account...
Aug  2 13:52:56.199 [INFO] [TurboBlankAccount (905994805379)] [Turbonomic-OpsMgr-Target] Searching for users matching the username "Turbonomic-OpsMgr-Target" and/or tags "[Owner:Turbonomic Turbonomic-Host:msse01.demo.turbonomic.com]"
Aug  2 13:52:57.139 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] Creating User
Aug  2 13:52:57.286 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] User Created
Aug  2 13:52:57.286 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] Adding Polices to user which allow Turbonomic read only access. Policies: [AmazonEC2ReadOnlyAccess AmazonS3ReadOnlyAccess AmazonRDSReadOnlyAccess]
Aug  2 13:52:57.286 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess) to user
Aug  2 13:52:57.449 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess) to user
Aug  2 13:52:57.616 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess) to user
Aug  2 13:52:57.982 [INFO] [TurboBlankAccount (905994805379)] [AddUser] [Turbonomic-OpsMgr-Target] User access key created.
Aug  2 13:52:57.982 [INFO] [Ryan J. Geyer (385266030856)] Assuming the role OrganizationAccountAccessRole on account...
Aug  2 13:52:57.982 [INFO] [Ryan J. Geyer (385266030856)] [Turbonomic-OpsMgr-Target] Searching for users matching the username "Turbonomic-OpsMgr-Target" and/or tags "[Owner:Turbonomic Turbonomic-Host:msse01.demo.turbonomic.com]"
Aug  2 13:52:59.400 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] Creating User
Aug  2 13:52:59.546 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] User Created
Aug  2 13:52:59.546 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] Adding Polices to user which allow Turbonomic read only access. Policies: [AmazonEC2ReadOnlyAccess AmazonS3ReadOnlyAccess AmazonRDSReadOnlyAccess]
Aug  2 13:52:59.546 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess) to user
Aug  2 13:52:59.688 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess) to user
Aug  2 13:52:59.848 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] Attaching policy (arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess) to user
Aug  2 13:53:00.348 [INFO] [Ryan J. Geyer (385266030856)] [AddUser] [Turbonomic-OpsMgr-Target] User access key created.
```

#### (Only) Create IAM Roles
This use case will iterate over every account in your AWS org.

In each account, a new IAM role will be created with the appropriate policies for either read-only or automation access from the Turbonomic OpsMgr.

If an error is encountered, the account will be skipped.

This should only be used with the `--aws-account-file` flag, otherwise the users will be created, but there will be no record of the user credentials. This file can then later be used with different flags to add the [Turbonomic targets](#only-create-turbonomic-targets).

##### Prerequisites
* The name of a role which exist in each account with permissions to create IAM principals. By default this is the "OrganizationAccountAccessRole" `--x-acct-role`
* AWS Access Key Id and Secret of a user with permissions to list all accounts in your org, and assume the role above. `--aws-access-key-id` and `--aws-secret-access-key`
* The AWS account ID where the Turbonomic OpsMgr is running `--turbo-trusted-account-id`
* The AWS IAM role assumed by the Turbonomic OpsMgr `--turbo-tursted-account-role`
* The AWS EC2 Instance id of the Turbonomic OpsMgr `--turbo-trusted-account-instanceid`

##### Example
```
./turbo_cloud_tool aws --aws-access-key-id $AWS_ACCESS_KEY_ID --aws-secret-access-key "$AWS_SECRET_ACCESS_KEY" --iam-principal-create --iam-principal-type role --turbo-trusted-account-id 385266030856 --turbo-trusted-account-role TurboSEEnablementXAccountTrust --turbo-trusted-account-instanceid i-097ec442cd152a92b --aws-account-file output.json
Aug  2 13:58:37.296 [INFO] Using config file: /Users/ryangeyer/.turbo_cloud_tool.yaml
Aug  2 13:58:37.296 [INFO] Querying org for list of child accounts...
Aug  2 13:58:38.090 [INFO] [TurboBlankAccount (905994805379)] Assuming the role OrganizationAccountAccessRole on account...
Aug  2 13:58:38.090 [INFO] [TurboBlankAccount (905994805379)] [Turbonomic-OpsMgr-Target] Searching for roles matching the rolename "Turbonomic-OpsMgr-Target" and/or tags "[Owner:Turbonomic Turbonomic-Host:msse01.demo.turbonomic.com]"
Aug  2 13:58:39.550 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Creating Role
Aug  2 13:58:39.703 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Role Created
Aug  2 13:58:39.703 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Adding Polices to role which allow Turbonomic read only access. Policies: [AmazonEC2ReadOnlyAccess AmazonS3ReadOnlyAccess AmazonRDSReadOnlyAccess]
Aug  2 13:58:39.703 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess
Aug  2 13:58:39.835 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
Aug  2 13:58:39.958 [INFO] [TurboBlankAccount (905994805379)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess
Aug  2 13:58:40.080 [INFO] [Ryan J. Geyer (385266030856)] Assuming the role OrganizationAccountAccessRole on account...
Aug  2 13:58:40.080 [INFO] [Ryan J. Geyer (385266030856)] [Turbonomic-OpsMgr-Target] Searching for roles matching the rolename "Turbonomic-OpsMgr-Target" and/or tags "[Owner:Turbonomic Turbonomic-Host:msse01.demo.turbonomic.com]"
Aug  2 13:58:42.257 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Creating Role
Aug  2 13:58:42.408 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Role Created
Aug  2 13:58:42.408 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Adding Polices to role which allow Turbonomic read only access. Policies: [AmazonEC2ReadOnlyAccess AmazonS3ReadOnlyAccess AmazonRDSReadOnlyAccess]
Aug  2 13:58:42.408 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess
Aug  2 13:58:42.533 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
Aug  2 13:58:42.659 [INFO] [Ryan J. Geyer (385266030856)] [AddRole] [Turbonomic-OpsMgr-Target] Attaching policy to role. Policy: arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess
```

#### (Only) Create Turbonomic Targets
Not implemented properly yet..

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

## Delete existing IAM users, with confirmation
```
./turbo_cloud_tool aws --aws-access-key-id $AWS_ACCESS_KEY_ID --aws-secret-access-key "$AWS_SECRET_ACCESS_KEY" --iam-principal-type user --iam-principal-name Foo --turbo-hostname localhost --turbo-username "$TURBO_USERNAME" --turbo-password "$TURBO_PASSWORD" --iam-principal-delete --aws-account-file output.json
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
