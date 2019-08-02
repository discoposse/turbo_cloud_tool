# TODO
* General
  * Rename appropriately (cloud_target_tool?)
  * Build/release automation which creates binaries for each arch/platform (OSX, Windows, Linux)
  * Better encapsulation
  * Struct/logic for handling stdin. Lots of duplicated code here, need a standardized reusable mechanism.
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
    * From "in memory" principals created
    * From principals created and stored in AWS acct file
    * Go back and tag principal with Target UUID
  * Input validation (ensure that users provide all necessary inputs with useful error/warning/feedback)
  * Success/Fail summary
  * Separate commands to parse/rationalize the account file?
  * Create tags on principals with provided case (currently all tags are lowercase)
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

# Feature Ideas

* Validate that a given cloud account can be used.
  * Azure
    * Can download ratecard
    * Microsoft.Compute is registered (https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-manager-register-provider-errors#)
    * Other things?
  * AWS
    * Org exists, and these credentials have access?
    * Turbonomic OpsMgr Instance running?
