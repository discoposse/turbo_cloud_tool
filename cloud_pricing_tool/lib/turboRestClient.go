package lib

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type TurboTargetCreateResponse struct {
	Type string `json:"type,omitempty"`
	// Exception     string `json:"exception,omitempty"`
	// Message       string `json:"message,omitempty"`
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

func (t *TurboRestClient) DeleteTarget(uuid string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://%v/vmturbo/rest/targets/%v", t.Hostname, uuid), nil)
	req.SetBasicAuth(t.username, t.password)
	req.Header.Add("Accept", "application/json")
	resp, err := t.Client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP error attempting to delete target. Error: %v", err)
	}

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Unable to read API response to delete Target. Error: %v", err)
	}

	var retval TurboTargetCreateResponse
	err = json.Unmarshal(bodyText, &retval)
	if err != nil {
		return fmt.Errorf("Unable to marshal API response to an object. Error: %v\nResponse: %s", err, string(bodyText))
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Turbonomic API error attempting to delete target. Status Code: %v\nResponse: %s", resp.StatusCode, string(bodyText))
	}
	return nil
}

func (t *TurboRestClient) AddAwsUserCloudTarget(target_name string, username string, password string) (*TurboTargetCreateResponse, error) {
	inputDto := map[string]interface{}{
		"category": "Cloud Management",
		"type":     "AWS",
		"inputFields": []interface{}{
			map[string]string{
				"name":  "address",
				"value": target_name,
			},
			map[string]string{
				"name":  "username",
				"value": username,
			},
			map[string]string{
				"name":  "password",
				"value": password,
			},
		},
	}

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
