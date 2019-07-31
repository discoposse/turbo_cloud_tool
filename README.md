# TODO
* General
  * Rename appropriately (cloud_target_tool?)
  * Build/release automation which creates binaries for each arch/platform (OSX, Windows, Linux)
  * Better encapsulation
  * Struct/logic for handling stdin. Lots of duplicated code here, need a standardized reusable mechanism.
* Azure - On hold until 6.4 and/or more discovery
* AWS
  * User Delete
    * "Force" option with clear documentation
    * ~Complete confirmation flow before deleting~
  * Role Create
    * "Force" option with clear documentation
    * Bail when role already exists
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


# Feature Ideas

* Validate that a given cloud account can be used.
  * Azure
    * Can download ratecard
    * Microsoft.Compute is registered (https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-manager-register-provider-errors#)
    * Other things?
  * AWS
    * Org exists, and these credentials have access?
    * Turbonomic OpsMgr Instance running?
