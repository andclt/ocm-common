/*
Copyright (c) 2026 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ocm

import (
	"fmt"
	"net/http"

	sdk "github.com/openshift-online/ocm-sdk-go"
	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocmerrors "github.com/openshift-online/ocm-sdk-go/errors"
)

// OCMClient provides helper methods for interacting with OCM API
type OCMClient struct {
	connection *sdk.Connection
}

// NewOCMClient creates a new OCM client wrapper
func NewOCMClient(connection *sdk.Connection) *OCMClient {
	return &OCMClient{connection: connection}
}

// GetPolicies retrieves AWS STS policies from OCM filtered by policy type
// Example policy types: "OCMRole", "AccountRole", "OperatorRole"
func (c *OCMClient) GetPolicies(policyType string) (map[string]*cmv1.AWSSTSPolicy, error) {
	query := fmt.Sprintf("policy_type = '%s'", policyType)

	m := make(map[string]*cmv1.AWSSTSPolicy)

	stmt := c.connection.ClustersMgmt().V1().AWSInquiries().STSPolicies().List()
	if policyType != "" {
		stmt = stmt.Search(query)
	}
	policiesResponse, err := stmt.Send()
	if err != nil {
		return m, handleErr(policiesResponse.Error(), err)
	}
	policiesResponse.Items().Each(func(awsPolicy *cmv1.AWSSTSPolicy) bool {
		m[awsPolicy.ID()] = awsPolicy
		return true
	})
	return m, nil
}

// GetCurrentAccount retrieves the current account from OCM
func (c *OCMClient) GetCurrentAccount() (*amsv1.Account, error) {
	response, err := c.connection.AccountsMgmt().V1().
		CurrentAccount().
		Get().
		Send()
	if err != nil {
		if response.Status() == http.StatusNotFound {
			return nil, nil
		}
		return nil, handleErr(response.Error(), err)
	}
	return response.Body(), nil
}

// GetCurrentOrganization retrieves the current organization ID and external ID from OCM
func (c *OCMClient) GetCurrentOrganization() (id string, externalID string, err error) {
	acctResponse, err := c.GetCurrentAccount()
	if err != nil {
		return
	}
	id = acctResponse.Organization().ID()
	externalID = acctResponse.Organization().ExternalID()
	return
}

// handleErr converts OCM SDK errors to standard Go errors
func handleErr(res *ocmerrors.Error, err error) error {
	msg := err.Error()
	// Need to check nil as .Error will try to access internal values of the pointer
	if res != nil && res.Reason() != "" {
		msg = res.Error()
	}
	return fmt.Errorf("%s", msg)
}
