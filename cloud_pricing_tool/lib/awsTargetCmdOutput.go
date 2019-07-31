package lib

import "encoding/json"

type AwsTargetCmdAccount struct {
	Id          string
	Principal   *AwsTargetCmdPrincipal   `json:",omitempty"`
	TurboTarget *AwsTargetCmdTurboTarget `json:",omitempty"`
	Errors      []string
}

type AwsTargetCmdTurboTarget struct {
	Hostname   string
	TargetUuid string
	Errors     []string
}

type AwsTargetCmdPrincipal struct {
	PrincipalType   string
	Policies        []string `json:",omitempty"`
	Name            string
	AccessKeyId     string `json:",omitempty"`
	SecretAccessKey string `json:",omitempty"`
	Arn             string `json:",omitempty"`
	Errors          []string
}

type AwsTargetCmdOutput struct {
	Insecure bool `json:"-"`
	Accounts map[string]*AwsTargetCmdAccount
	Errors   []string
}

func (o *AwsTargetCmdOutput) MarshalJSON() ([]byte, error) {
	if !o.Insecure {
		for _, acct := range o.Accounts {
			if acct.Principal != nil {
				acct.Principal.SecretAccessKey = ""
			}
		}
	}
	mapToMarshal := make(map[string]interface{})
	mapToMarshal["Accounts"] = o.Accounts
	mapToMarshal["Errors"] = o.Errors
	return json.Marshal(mapToMarshal)
}
