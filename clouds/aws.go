package clouds

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/sirupsen/logrus"
)

/**
 * Borrowed a lot of authentication logic from https://maori.geek.nz/assuming-roles-in-aws-with-go-aeeb28fab418
 */
type Aws struct {
	session     *session.Session
	credentials *credentials.Credentials
	configs     map[string]*aws.Config
	Logger      *logrus.Logger
}

func (a *Aws) GetLogger() *logrus.Logger {
	if a.Logger == nil {
		a.Logger = logrus.New()
		a.Logger.SetOutput(nil)
	}
	return a.Logger
}

type AwsPrincipalMatch struct {
	Principal     interface{}
	Principalname string
	IamTags       []iam.Tag
	StringTags    []string
}

func (a *AwsPrincipalMatch) AsUser() (*iam.User, error) {
	if a.Principal == nil {
		return nil, fmt.Errorf("Principal not set")
	}
	if user, ok := a.Principal.(*iam.User); ok {
		return user, nil
	}
	return nil, fmt.Errorf("Principal is not a user. It is a %v", reflect.TypeOf(a.Principal))
}

func ConvertStringTags(tags []string) []iam.Tag {
	retval := []iam.Tag{}
	if tags != nil {
		for _, strTag := range tags {
			tagParts := strings.Split(strTag, ":")
			key := tagParts[0]
			value := tagParts[1]

			tag := iam.Tag{
				Value: &value,
				Key:   &key,
			}

			retval = append(retval, tag)
		}
	}
	return retval
}

func ConvertStringTagsPointer(tags []string) []*iam.Tag {
	retval := []*iam.Tag{}
	if tags != nil {
		for _, strTag := range tags {
			tagParts := strings.Split(strTag, ":")
			key := tagParts[0]
			value := tagParts[1]

			tag := &iam.Tag{
				Value: &value,
				Key:   &key,
			}

			retval = append(retval, tag)
		}
	}
	return retval
}

func (m *AwsPrincipalMatch) MatchedTags() []iam.Tag {
	retval := []iam.Tag{}
	if m.Principal == nil {
		return retval
	}

	var principalTags []*iam.Tag
	if principal, ok := m.Principal.(*iam.Role); ok {
		principalTags = principal.Tags
	}

	if principal, ok := m.Principal.(*iam.User); ok {
		principalTags = principal.Tags
	}

	if principalTags == nil {
		return retval
	}

	if m.IamTags == nil && m.StringTags == nil {
		return retval
	}

	if m.StringTags != nil {
		m.IamTags = ConvertStringTags(m.StringTags)
	}

	for _, iamTag := range m.IamTags {
		for _, principalTag := range principalTags {
			if strings.ToLower(*principalTag.Key) == strings.ToLower(*iamTag.Key) &&
				strings.ToLower(*principalTag.Value) == strings.ToLower(*iamTag.Value) {
				retval = append(retval, iamTag)
			}
		}
	}

	return retval
}

func (m *AwsPrincipalMatch) PrincipalnameMatch() bool {
	if m.Principal == nil {
		return false
	}

	var name *string
	if principal, ok := m.Principal.(*iam.Role); ok {
		name = principal.RoleName
	}

	if principal, ok := m.Principal.(*iam.User); ok {
		name = principal.UserName
	}

	if name == nil {
		return false
	}

	return strings.ToLower(*name) == strings.ToLower(m.Principalname)
}

func (m *AwsPrincipalMatch) AnyTagsMatch() bool {
	return len(m.MatchedTags()) > 0
}

func (m *AwsPrincipalMatch) AllTagsMatch() bool {
	matchedTags := m.MatchedTags()
	return len(matchedTags) == len(m.IamTags)
}

func (m *AwsPrincipalMatch) ExactMatch() bool {
	return m.PrincipalnameMatch() && m.AllTagsMatch()
}

func (m *AwsPrincipalMatch) AnyMatch() bool {
	return m.PrincipalnameMatch() || m.AnyTagsMatch()
}

func (m AwsPrincipalMatch) String() string {
	var principalType string
	var matchType string
	var principalName string
	var principalTags []*iam.Tag

	if role, ok := m.Principal.(*iam.Role); ok {
		principalType = "Role"
		principalName = *role.RoleName
		principalTags = role.Tags
	}

	if user, ok := m.Principal.(*iam.User); ok {
		principalType = "User"
		principalName = *user.UserName
		principalTags = user.Tags
	}

	if m.ExactMatch() {
		matchType = "Exact"
	} else if m.PrincipalnameMatch() && m.AnyTagsMatch() {
		matchType = "Name & Subset of Tags"
	} else if m.PrincipalnameMatch() {
		matchType = "Name"
	} else if m.AllTagsMatch() {
		matchType = "All Tags"
	} else if m.AnyTagsMatch() {
		matchType = "Subset of Tags"
	}

	var tagBytes bytes.Buffer
	for _, tag := range principalTags {
		tagBytes.WriteString(fmt.Sprintf("%v:%v, ", *tag.Key, *tag.Value))
	}
	// TODO: Surely there is a more elegant way to remove the trailing comma?
	tagBytes.Truncate(tagBytes.Len() - 2)

	return fmt.Sprintf(
		"[Match Type: %s] - [Name: %s] - [Tags: %v] - [Principal Type: %s]",
		matchType,
		principalName,
		tagBytes.String(),
		principalType,
	)
}

func (a *Aws) GetSession() *session.Session {
	if a.session != nil {
		return a.session
	}
	if a.credentials != nil {
		sess := session.Must(session.NewSession(&aws.Config{
			Region:      aws.String("us-east-1"),
			Credentials: a.credentials,
		}))
		a.session = sess
		return a.session
	}
	sess := session.Must(session.NewSession())
	a.session = sess
	return a.session
}

func (a *Aws) GetConfig(region string, account_id string, role string) *aws.Config {
	arn := fmt.Sprintf("arn:aws:iam::%v:role/%v", account_id, role)
	key := fmt.Sprintf("%v::%v", region, arn)

	if a.configs != nil && a.configs[key] != nil {
		return a.configs[key]
	}

	creds := stscreds.NewCredentials(a.GetSession(), arn)

	config := aws.NewConfig().
		WithCredentials(creds).
		WithRegion(region).
		WithMaxRetries(10)

	if a.configs == nil {
		a.configs = map[string]*aws.Config{}
	}

	a.configs[key] = config
	return a.configs[key]
}

func (a *Aws) SetCredentials(access_key string, secret_key string) {
	a.credentials = credentials.NewStaticCredentials(access_key, secret_key, "")
}

func (a *Aws) ListAllOrgAccounts() ([]*organizations.Account, error) {
	orgCli := organizations.New(a.GetSession())
	retval := []*organizations.Account{}

	err := orgCli.ListAccountsPagesWithContext(context.Background(), &organizations.ListAccountsInput{},
		func(lo *organizations.ListAccountsOutput, lastPage bool) bool {
			for _, acct := range lo.Accounts {
				retval = append(retval, acct)
			}
			return true // Always continue
		})
	return retval, err
}

func (a *Aws) FindMatchingUsers(account_id string, role string, username string, tags []string) ([]AwsPrincipalMatch, error) {
	iamCli := iam.New(a.GetSession(), a.GetConfig("us-east-1", account_id, role))
	retval := []AwsPrincipalMatch{}
	var tagErr error
	err := iamCli.ListUsersPagesWithContext(context.Background(), &iam.ListUsersInput{}, func(arg1 *iam.ListUsersOutput, arg2 bool) bool {
		for _, user := range arg1.Users {
			tagResp, err := iamCli.ListUserTagsWithContext(context.Background(), &iam.ListUserTagsInput{UserName: user.UserName})
			if err != nil {
				tagErr = fmt.Errorf("Unable to query tags for user %s. Error: %v", username, err)
				return false
			}
			user.Tags = tagResp.Tags
			match := AwsPrincipalMatch{
				Principal:     user,
				Principalname: username,
				StringTags:    tags,
			}
			if match.AnyMatch() {
				retval = append(retval, match)
			}
		}
		return true
	})
	if tagErr != nil {
		return retval, tagErr
	}
	return retval, err
}

func (a *Aws) FindMatchingRoles(account_id string, role string, rolename string, tags []string) ([]AwsPrincipalMatch, error) {
	iamCli := iam.New(a.GetSession(), a.GetConfig("us-east-1", account_id, role))
	retval := []AwsPrincipalMatch{}
	var tagErr error
	err := iamCli.ListRolesPagesWithContext(context.Background(), &iam.ListRolesInput{}, func(arg1 *iam.ListRolesOutput, arg2 bool) bool {
		for _, iterRole := range arg1.Roles {
			tagResp, err := iamCli.ListRoleTags(&iam.ListRoleTagsInput{RoleName: iterRole.RoleName})
			if err != nil {
				tagErr = fmt.Errorf("Unable to query tags for role %s. Error: %v", rolename, err)
				return false
			}
			iterRole.Tags = tagResp.Tags
			match := AwsPrincipalMatch{
				Principal:     iterRole,
				Principalname: rolename,
				StringTags:    tags,
			}
			if match.AnyMatch() {
				retval = append(retval, match)
			}
		}
		return true
	})
	if tagErr != nil {
		return retval, tagErr
	}
	return retval, err
}
