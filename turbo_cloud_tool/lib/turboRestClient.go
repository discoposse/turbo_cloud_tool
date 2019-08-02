package lib

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type TurboCloudTargetDto struct {
	Category    string
	Type        string
	InputFields map[string]string
}

func (t *TurboCloudTargetDto) MarshalJSON() ([]byte, error) {
	inputFields := []map[string]string{}
	for name, value := range t.InputFields {
		inputFields = append(inputFields, map[string]string{"name": name, "value": value})
	}
	customMarshal := map[string]interface{}{
		"category":    t.Category,
		"type":        t.Type,
		"inputFields": inputFields,
	}
	return json.Marshal(customMarshal)
}

type TurboRestErrorResponse struct {
	Type      uint   `json:"type,omitempty"`
	Exception string `json:"exception,omitempty"`
	Message   string `json:"message,omitempty"`
}

type TurboTargetCreateResponse struct {
	Type          string `json:"type,omitempty"`
	Exception     string `json:"exception,omitempty"`
	Message       string `json:"message,omitempty"`
	Uuid          string `json:"uuid,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
	Category      string `json:"category,omitempty"`
	LastValidated string `json:"lastValidated,omitempty"`
	Status        string `json:"status,omitempty"`
}

type TurboRestClient struct {
	Client   *http.Client
	Hostname string
	username string
	password string
}

func NewTurboRestClient(hostname string, username string, password string) TurboRestClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return TurboRestClient{
		Client:   client,
		Hostname: hostname,
		username: username,
		password: password,
	}
}

func (t *TurboRestClient) DeleteTarget(uuid string) (*TurboRestErrorResponse, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://%v/vmturbo/rest/targets/%v", t.Hostname, uuid), nil)
	req.SetBasicAuth(t.username, t.password)
	req.Header.Add("Accept", "application/json")
	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP error attempting to delete target. Error: %v", err)
	}

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read API response to delete Target. Error: %v", err)
	}

	if resp.StatusCode != 200 {
		var retval TurboRestErrorResponse
		err = json.Unmarshal(bodyText, &retval)
		if err != nil {
			return nil, fmt.Errorf("Unable to marshal API error response to an object. Error: %v\nResponse: %s", err, string(bodyText))
		}
		return &retval, fmt.Errorf("Turbonomic API error attempting to delete target. Status Code: %v\nResponse: %s", resp.StatusCode, string(bodyText))
	}
	return nil, nil
}

func (t *TurboRestClient) addTarget(inputDto TurboCloudTargetDto) (*TurboTargetCreateResponse, error) {
	bodyBytes, err := json.Marshal(inputDto)
	body := ioutil.NopCloser(bytes.NewReader(bodyBytes))

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%v/vmturbo/rest/targets/", t.Hostname), body)
	req.SetBasicAuth(t.username, t.password)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP error attempting to create AWS target. Error: %v\nRequest: %s", err, string(bodyBytes))
	}

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read API response to create AWS Target. Error: %v", err)
	}

	var retval TurboTargetCreateResponse
	err = json.Unmarshal(bodyText, &retval)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal API response to an object. Error: %v\nResponse: %s", err, string(bodyText))
	}

	if resp.StatusCode != 200 {
		return &retval, fmt.Errorf("Turbonomic API error attempting to create AWS target. Status Code: %v\nRequest: %s\nResponse: %s", resp.StatusCode, string(bodyBytes), string(bodyText))
	}

	return &retval, err
}

func (t *TurboRestClient) AddAwsRoleCloudTarget(target_name string, role_arn string) (*TurboTargetCreateResponse, error) {
	inputDto := TurboCloudTargetDto{
		Category: "Cloud Management",
		Type:     "AWS",
		InputFields: map[string]string{
			"address":  target_name,
			"username": "key",
			"password": "secret",
			"iamRole":  role_arn,
		},
	}

	return t.addTarget(inputDto)
}

func (t *TurboRestClient) AddAwsUserCloudTarget(target_name string, username string, password string) (*TurboTargetCreateResponse, error) {
	inputDto := TurboCloudTargetDto{
		Category: "Cloud Management",
		Type:     "AWS",
		InputFields: map[string]string{
			"address":  target_name,
			"username": username,
			"password": password,
		},
	}

	return t.addTarget(inputDto)
}
