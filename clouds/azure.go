package clouds

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/graphrbac/graphrbac"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// Mocks for testing
// type TenantsClient interface {
// 	List(context.Context) (subscriptions.TenantListResultPage, error)
// }

type DeviceFlowConfig interface {
	ServicePrincipalToken() (*adal.ServicePrincipalToken, error)
}

// The main struct
type Azure struct {
	TenantId                         string
	ClientId                         string
	ClientSecret                     string
	spToken                          *adal.ServicePrincipalToken
	authorizer                       autorest.Authorizer
	tenantsClient                    *subscriptions.TenantsClient
	subscriptionsClient              *subscriptions.Client
	graphRbacSignedInUserClient      *graphrbac.SignedInUserClient
	graphRbacServicePrincipalsClient *graphrbac.ServicePrincipalsClient
	graphRbacApplicationsClient      *graphrbac.ApplicationsClient
	roleAssignmentsClient            *authorization.RoleAssignmentsClient
	deviceFlowConfig                 DeviceFlowConfig
}

func (a *Azure) isAuthorized() (bool, error) {
	if a.spToken == nil {
		return false, errors.New("No authorization information for Azure. Please first call one of the SetAuth* methods to authenticate with Azure.")
	}
	return true, nil
}

// Authentication options
func (a *Azure) SetAuthServiceToken(tokenjson string) error {
	err := json.Unmarshal([]byte(tokenjson), &a.spToken)
	if err != nil {
		return fmt.Errorf("Unable to parse JSON for service token. Caused By: %v", err)
	}
	return nil
}

func (a *Azure) SetAuthDeviceFlow(clientid string, tenantid string) error {
	if a.deviceFlowConfig == nil {
		devFlowConfig := auth.NewDeviceFlowConfig(clientid, tenantid)
		a.deviceFlowConfig = devFlowConfig
	}
	token, err := a.deviceFlowConfig.ServicePrincipalToken()
	a.spToken = token
	return err
}

func (a *Azure) SetAuthClientCredentials(clientId string, clientSecret string, tenantId string) error {
	config := auth.NewClientCredentialsConfig(clientId, clientSecret, tenantId)
	config.AADEndpoint = azure.PublicCloud.ActiveDirectoryEndpoint
	token, err := config.ServicePrincipalToken()
	if err != nil {
		return err
	}
	a.spToken = token
	return nil
}

func (a *Azure) GetSubscriptionClient() (subscriptions.Client, error) {
	if authd, err := a.isAuthorized(); !authd {
		return subscriptions.Client{}, err
	}
	if a.subscriptionsClient == nil {
		subs := subscriptions.NewClient()
		a.spToken.RefreshExchange(azure.PublicCloud.ResourceManagerEndpoint)
		subs.Authorizer = autorest.NewBearerAuthorizer(a.spToken)
		a.subscriptionsClient = &subs
	}
	return *a.subscriptionsClient, nil
}

func (a *Azure) GetRbacGraphSignedInUserClient() (graphrbac.SignedInUserClient, error) {
	if authd, err := a.isAuthorized(); !authd {
		return graphrbac.SignedInUserClient{}, err
	}
	if a.graphRbacSignedInUserClient == nil {
		siCli := graphrbac.NewSignedInUserClient("common")
		siCli.Authorizer = a.authorizer
		a.graphRbacSignedInUserClient = &siCli
	}
	return *a.graphRbacSignedInUserClient, nil
}

func (a *Azure) GetServicePrincipalsClient(tenantId string) (graphrbac.ServicePrincipalsClient, error) {
	if authd, err := a.isAuthorized(); !authd {
		return graphrbac.ServicePrincipalsClient{}, err
	}
	if a.graphRbacServicePrincipalsClient == nil {
		appcli := graphrbac.NewServicePrincipalsClient(tenantId)
		err := a.spToken.RefreshExchange(azure.PublicCloud.GraphEndpoint)
		if err != nil {
			return graphrbac.ServicePrincipalsClient{}, fmt.Errorf("Unable to refresh authorization token with graph endpoint. Caused By: %v", err)
		}
		appcli.Authorizer = autorest.NewBearerAuthorizer(a.spToken)
		a.graphRbacServicePrincipalsClient = &appcli
	}
	return *a.graphRbacServicePrincipalsClient, nil
}

func (a *Azure) GetApplicationsClient(tenantId string) (graphrbac.ApplicationsClient, error) {
	if authd, err := a.isAuthorized(); !authd {
		return graphrbac.ApplicationsClient{}, err
	}
	if a.graphRbacApplicationsClient == nil {
		apps := graphrbac.NewApplicationsClient(tenantId)
		err := a.spToken.RefreshExchange(azure.PublicCloud.GraphEndpoint)
		if err != nil {
			return graphrbac.ApplicationsClient{}, fmt.Errorf("Unable to refresh authorization token with graph endpoint. Caused By: %v", err)
		}
		apps.Authorizer = autorest.NewBearerAuthorizer(a.spToken)
		a.graphRbacApplicationsClient = &apps
	}
	return *a.graphRbacApplicationsClient, nil
}

func (a *Azure) GetTenantsClient() (subscriptions.TenantsClient, error) {
	if authd, err := a.isAuthorized(); !authd {
		return subscriptions.TenantsClient{}, err
	}
	if a.tenantsClient == nil {
		tens := subscriptions.NewTenantsClient()
		a.spToken.RefreshExchange(azure.PublicCloud.ResourceManagerEndpoint)
		tens.Authorizer = autorest.NewBearerAuthorizer(a.spToken)
		a.tenantsClient = &tens
	}
	return *a.tenantsClient, nil
}

func (a *Azure) GetRoleAssignmentsClient(subscriptionId string) (authorization.RoleAssignmentsClient, error) {
	if authd, err := a.isAuthorized(); !authd {
		return authorization.RoleAssignmentsClient{}, err
	}
	if a.roleAssignmentsClient == nil {
		roleCli := authorization.NewRoleAssignmentsClient(subscriptionId)
		a.spToken.RefreshExchange(azure.PublicCloud.ResourceManagerEndpoint)
		roleCli.Authorizer = autorest.NewBearerAuthorizer(a.spToken)
		a.roleAssignmentsClient = &roleCli
	}
	return *a.roleAssignmentsClient, nil
}
