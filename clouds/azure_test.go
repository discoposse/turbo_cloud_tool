package clouds

import (
	. "git.turbonomic.com/rgeyer/cloud_pricing_tool"
	// . "git.turbonomic.com/rgeyer/cloud_pricing_tool/clouds/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = Describe("Azure", func() {
	var (
		t        GinkgoTestReporter
		mockCtrl *gomock.Controller
		token    string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(t)
		token = `
    {
      "token": {
        "access_token": "access_token",
        "refresh_token": "refresh_token",
        "expires_in": 3600,
        "expires_on": 1562806114,
        "not_before": 1562802214,
        "resource": "https://management.azure.com/",
        "token_type": "Bearer"
      },
      "secret": {
        "type": "ServicePrincipalNoSecret"
      },
      "oauth": {
        "authorityEndpoint": {
          "Scheme": "https",
          "Opaque": "",
          "User": null,
          "Host": "login.microsoftonline.com",
          "Path": "/common",
          "RawPath": "",
          "ForceQuery": false,
          "RawQuery": "",
          "Fragment": ""
        },
        "authorizeEndpoint": {
          "Scheme": "https",
          "Opaque": "",
          "User": null,
          "Host": "login.microsoftonline.com",
          "Path": "/common/oauth2/authorize",
          "RawPath": "",
          "ForceQuery": false,
          "RawQuery": "api-version=1.0",
          "Fragment": ""
        },
        "tokenEndpoint": {
          "Scheme": "https",
          "Opaque": "",
          "User": null,
          "Host": "login.microsoftonline.com",
          "Path": "/common/oauth2/token",
          "RawPath": "",
          "ForceQuery": false,
          "RawQuery": "api-version=1.0",
          "Fragment": ""
        },
        "deviceCodeEndpoint": {
          "Scheme": "https",
          "Opaque": "",
          "User": null,
          "Host": "login.microsoftonline.com",
          "Path": "/common/oauth2/devicecode",
          "RawPath": "",
          "ForceQuery": false,
          "RawQuery": "api-version=1.0",
          "Fragment": ""
        }
      },
      "clientID": "ce3d555d-4fc9-4c34-898d-1d790a9b5c4b",
      "resource": "https://management.azure.com/",
      "autoRefresh": true,
      "refreshWithin": 300000000000
    }
              `
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Authorization", func() {
		PContext("Bearer Token", func() {
			Context("Token Provided", func() {
				It("Tests Token", func() {
					// azure := &Azure{}
					// err := azure.AuthWithServiceToken(token)
					// Ω(err).Should(BeNil())
					// Ω(azure.authorizer).ShouldNot(BeNil())
				})
			})

			Context("New Token from Device Flow", func() {
				It("Assigns Authorizer", func() {
					// Fail("Not implemented")
					// devFlowConfig := NewMockDeviceFlowConfig(mockCtrl)
					// devFlowConfig.EXPECT().
					// 	ServicePrincipalToken().
					// 	Return(&adal.ServicePrincipalToken{}, nil).
					// 	Times(1)
					// azure := &Azure{
					// 	deviceFlowConfig: devFlowConfig,
					// }
					// _, err := azure.AuthWithDeviceFlow("client id", "tenant id (common)")
					// Ω(err).Should(BeNil())
					// Ω(azure.authorizer).ShouldNot(BeNil())
				})
			})
		})

		PContext("Service Principal Key & Secret", func() {
			It("Assigns Authorizer", func() {
				// Fail("Not implemented")
				// azure := &Azure{}
				// _, err := azure.AuthWithClientCredentials("clientId", "clientSecret", "tenantId")
				// Ω(err).Should(BeNil())
				// Ω(azure.authorizer).ShouldNot(BeNil())
			})
		})
	})
})
