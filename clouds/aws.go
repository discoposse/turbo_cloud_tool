package clouds

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	log "github.com/sirupsen/logrus"
)

/**
 * Borrowed a lot of authentication logic from https://maori.geek.nz/assuming-roles-in-aws-with-go-aeeb28fab418
 */
type Aws struct {
	session     *session.Session
	credentials *credentials.Credentials
	configs     map[string]*aws.Config
}

type AwsUserMatch struct {
	User       *iam.User
	Username   string
	IamTags    []iam.Tag
	StringTags []string
}

func ConvertStringTags(tags []string) []iam.Tag {
	retval := []iam.Tag{}
	if tags != nil {
		for _, strTag := range tags {
			tagParts := strings.Split(strings.ToLower(strTag), ":")
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
			tagParts := strings.Split(strings.ToLower(strTag), ":")
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

func (m *AwsUserMatch) MatchedTags() []iam.Tag {
	retval := []iam.Tag{}
	if m.User == nil {
		return retval
	}

	if m.User.Tags == nil {
		return retval
	}

	if m.IamTags == nil && m.StringTags == nil {
		return retval
	}

	if m.StringTags != nil {
		m.IamTags = ConvertStringTags(m.StringTags)
	}

	for _, iamTag := range m.IamTags {
		for _, userTag := range m.User.Tags {
			if strings.ToLower(*userTag.Key) == *iamTag.Key &&
				strings.ToLower(*userTag.Value) == *iamTag.Value {
				retval = append(retval, iamTag)
			}
		}
	}

	return retval
}

func (m *AwsUserMatch) UsernameMatch() bool {
	if m.User == nil {
		return false
	}

	if m.User.UserName == nil {
		return false
	}

	return strings.ToLower(*m.User.UserName) == strings.ToLower(m.Username)
}

func (m *AwsUserMatch) AnyTagsMatch() bool {
	return len(m.MatchedTags()) > 0
}

func (m *AwsUserMatch) AllTagsMatch() bool {
	matchedTags := m.MatchedTags()
	return len(matchedTags) == len(m.IamTags)
}

func (m *AwsUserMatch) ExactMatch() bool {
	return m.UsernameMatch() && m.AllTagsMatch()
}

func (m *AwsUserMatch) AnyMatch() bool {
	return m.UsernameMatch() || m.AnyTagsMatch()
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

func (a *Aws) FindMatchingUsers(account_id string, role string, username string, tags []string) ([]AwsUserMatch, error) {
	iamCli := iam.New(a.GetSession(), a.GetConfig("us-east-1", account_id, role))
	retval := []AwsUserMatch{}
	err := iamCli.ListUsersPagesWithContext(context.Background(), &iam.ListUsersInput{}, func(arg1 *iam.ListUsersOutput, arg2 bool) bool {
		for _, user := range arg1.Users {
			log.WithField("User", user).Debug("Checking user for match")
			match := AwsUserMatch{
				User:       user,
				Username:   username,
				StringTags: tags,
			}
			if match.AnyMatch() {
				retval = append(retval, match)
			}
		}
		return true
	})
	return retval, err
}
