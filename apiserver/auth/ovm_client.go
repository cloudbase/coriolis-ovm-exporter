package auth

import (
	"fmt"

	"github.com/dbgeek/go-ovm-helper/ovmHelper"

	gErrors "coriolis-ovm-exporter/errors"
)

// NewOVMClient returns a new OVMClient
func NewOVMClient(username, password, endpoint string) *OVMClient {
	client := ovmHelper.NewClient(username, password, endpoint)
	return &OVMClient{
		client: client,
	}
}

type repo struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	URI   string `json:"uri"`
	Value string `json:"value"`
}

// OVMClient is a helper OVM client. We use it to validate authentication
// data of the client.
type OVMClient struct {
	client *ovmHelper.Client
}

// AttemptRequest makes an authenticated request to the OVM API endpoint,
// to validate that the supplied username and password are correct. In this
// case, we simply attempt to list repositories.
func (o *OVMClient) AttemptRequest() error {
	req, err := o.client.NewRequest("GET", "/ovm/core/wsapi/rest/Repository/id", nil, nil)
	if err != nil {
		return gErrors.NewUnauthorizedError(
			fmt.Sprintf("failed to login: %s", err))
	}

	var m []repo
	_, err = o.client.Do(req, &m)

	if err != nil {
		return gErrors.NewUnauthorizedError(
			fmt.Sprintf("failed to login: %s", err))
	}

	return nil
}
