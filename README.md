# Feature Ideas

* Validate that a given cloud account can be used.
  * Azure
    * Can download ratecard
    * Microsoft.Compute is registered (https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-manager-register-provider-errors#)
    * Other things?
  * AWS
    * Org exists, and these credentials have access?
    * Turbonomic OpsMgr Instance running?


* Bulk add target(s)
  * Azure
    * Discover all accessible subscriptions (validate using above?)
    * Offer to add all of them (selector to allow only some?)
  * AWS


* CLI Menu - What you do first
  * cloud
    * credentials
      * list
      * save
    * ratecard
      * download
      * lookup/query
  * turbo
    * ratecard
      * download
      * lookup/query
    * credentials
      * list
      * save
    * target
      * validate
        * cloud
      * add
        * cloud
      * delete
        * cloud
