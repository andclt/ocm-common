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

package ocmrole

import (
	"fmt"
	"strings"

	"github.com/openshift-online/ocm-common/pkg/aws/utils"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const (
	// OCMRole is the role type for OCM roles
	OCMRole = "OCM"

	// OCMRolePolicyFile is the policy file name for standard OCM role
	OCMRolePolicyFile = "ocm"

	// OCMAdminRolePolicyFile is the policy file name for admin OCM role
	OCMAdminRolePolicyFile = "ocm_admin"

	// OCMNoConsoleRolePolicyFile is the policy file name for no-console OCM role
	OCMNoConsoleRolePolicyFile = "ocm_no_console"
)

// JumpAccounts are the various AWS accounts used for the installer jump role in the various OCM environments
var JumpAccounts = map[string]string{
	"production":  "710019948333",
	"staging":     "644306948063",
	"integration": "896164604406",
	"local":       "765374464689",
	"local-proxy": "765374464689",
	"crc":         "765374464689",
}

// GetJumpAccount returns the AWS account ID for the OCM jump role for a given environment
func GetJumpAccount(env string) string {
	return JumpAccounts[env]
}

// GetOCMRoleName constructs the OCM role name from prefix and external ID
func GetOCMRoleName(prefix string, role string, postfix string) string {
	name := fmt.Sprintf("%s-%s-Role-%s", prefix, role, postfix)
	return utils.TruncateRoleName(name)
}

// GetNoConsolePolicyName returns the policy name for a no-console OCM role
func GetNoConsolePolicyName(name string) string {
	return fmt.Sprintf("%s-NoConsole-Policy", name)
}

// GetAdminPolicyName returns the policy name for an admin OCM role
func GetAdminPolicyName(name string) string {
	return fmt.Sprintf("%s-Admin-Policy", name)
}

// GetPolicyName returns the standard policy name for an OCM role
func GetPolicyName(name string) string {
	return fmt.Sprintf("%s-Policy", name)
}

// GetNoConsolePolicyARN returns the ARN for a no-console policy
func GetNoConsolePolicyARN(partition string, accountID string, name string, path string) string {
	return getPolicyARN(partition, accountID, GetNoConsolePolicyName(name), path)
}

// GetAdminPolicyARN returns the ARN for an admin policy
func GetAdminPolicyARN(partition string, accountID string, name string, path string) string {
	return getPolicyARN(partition, accountID, GetAdminPolicyName(name), path)
}

// GetPolicyArnWithSuffix returns the ARN for a standard policy with suffix
func GetPolicyArnWithSuffix(partition string, accountID string, name string, path string) string {
	return getPolicyARN(partition, accountID, GetPolicyName(name), path)
}

// getPolicyARN constructs an IAM policy ARN
func getPolicyARN(partition string, accountID string, name string, path string) string {
	str := fmt.Sprintf("arn:%s:iam::%s:policy", partition, accountID)
	if path != "" {
		str = fmt.Sprintf("%s%s", str, path)
		return fmt.Sprintf("%s%s", str, name)
	}
	return fmt.Sprintf("%s/%s", str, name)
}

// GetPolicyDetails retrieves the policy details from the policies map
// Works for both managed and unmanaged policies
func GetPolicyDetails(policies map[string]*cmv1.AWSSTSPolicy, key string) string {
	policy, ok := policies[key]
	if ok {
		return policy.Details()
	}
	return ""
}

// GetManagedPolicyARN retrieves the managed policy ARN from the policies map
func GetManagedPolicyARN(policies map[string]*cmv1.AWSSTSPolicy, key string) (string, error) {
	policy, ok := policies[key]
	if !ok {
		return "", fmt.Errorf("policy %s not found", key)
	}
	if policy.ARN() == "" {
		return "", fmt.Errorf("policy %s does not have a managed ARN", key)
	}
	return policy.ARN(), nil
}

// InterpolatePolicyDocument replaces placeholders in policy documents with actual values
// Placeholders are in the format %{key}
func InterpolatePolicyDocument(partition string, doc string, replacements map[string]string) string {
	for key, val := range replacements {
		doc = strings.ReplaceAll(doc, fmt.Sprintf("%%{%s}", key), val)
	}

	// Handle partition replacement for legacy policies
	doc = strings.ReplaceAll(doc, "arn:aws:", fmt.Sprintf("arn:%s:", partition))

	return doc
}
