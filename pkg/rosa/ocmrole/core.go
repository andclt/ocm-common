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
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/openshift-online/ocm-common/pkg/aws/aws_client"
	common "github.com/openshift-online/ocm-common/pkg/aws/validations"
	"github.com/openshift-online/ocm-common/pkg/log"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// AWSClientInterface defines the AWS operations needed for OCM role management
// This interface allows for testing with mocks
type AWSClientInterface interface {
	GetRole(roleName string) (*types.Role, error)
	ListAttachedRolePolicies(roleName string) ([]types.AttachedPolicy, error)
	TagRole(roleName string, tags map[string]string) error
	CreateRole(roleName, assumeRolePolicyDocument, permissionBoundary string, tags map[string]string, path string) (types.Role, error)
	CreateIAMPolicy(policyName, policyDocument string, tags map[string]string) (*types.Policy, error)
	AttachIAMPolicy(roleName, policyARN string) error
}

// Ensure aws_client.AWSClient implements the interface
var _ AWSClientInterface = (*aws_client.AWSClient)(nil)

// ErrRoleExistsWrongProfile is returned when a role exists but with a different profile than requested
var ErrRoleExistsWrongProfile = errors.New("role exists with different profile")

// RoleProfile defines the type of OCM role to create
type RoleProfile string

const (
	ProfileStandard  RoleProfile = "standard"
	ProfileAdmin     RoleProfile = "admin"
	ProfileNoConsole RoleProfile = "no-console"
)

// Creator contains AWS account metadata
type Creator struct {
	Partition string
	AccountID string
}

// DetermineProfile selects the appropriate role profile based on flags
func DetermineProfile(isAdmin, isNoConsole bool) RoleProfile {
	switch {
	case isAdmin:
		return ProfileAdmin
	case isNoConsole:
		return ProfileNoConsole
	default:
		return ProfileStandard
	}
}

// CheckRoleExists checks if an OCM role exists and validates it matches the requested profile.
// Performs self-healing for policy/tag mismatches.
//
// Returns:
//   - roleARN: The ARN of the existing role (if exists=true)
//   - exists: Whether the role exists
//   - error: Any error encountered, including ErrRoleExistsWrongProfile when role exists but profile doesn't match
func CheckRoleExists(
	awsClient AWSClientInterface,
	creator *Creator,
	roleName string,
	profile RoleProfile,
	rolePath string,
) (string, bool, error) {
	role, err := awsClient.GetRole(roleName)
	if err != nil {
		// Role doesn't exist
		return "", false, nil
	}

	roleARN := *role.Arn
	isExistingRoleAdmin, err := isAdminRole(role)
	if err != nil {
		return "", true, err
	}
	isExistingRoleNoConsole, err := isNoConsoleRole(role)
	if err != nil {
		return "", true, err
	}

	log.LogInfo("Role '%s' already exists", roleName)

	switch profile {
	case ProfileStandard:
		if isExistingRoleAdmin {
			return roleARN, true, fmt.Errorf("the existing role is an admin role."+
				" To remove admin capabilities please delete the admin policy and the '%s' tag",
				AdminRole)
		}
		if isExistingRoleNoConsole {
			return roleARN, true, fmt.Errorf("the existing role is a no-console role." +
				" To use standard permissions please delete the role and recreate it")
		}
		return roleARN, true, nil

	case ProfileAdmin:
		if isExistingRoleNoConsole {
			return roleARN, true, fmt.Errorf("the existing role is a no-console role." +
				" To use admin permissions please delete the role and recreate it")
		}

		if isExistingRoleAdmin {
			return roleARN, true, nil
		}

		// Role appears to be standard - check if admin policy is actually attached (self-healing)
		attachedPolicies, err := awsClient.ListAttachedRolePolicies(roleName)
		if err != nil {
			return "", true, err
		}

		// Check if admin policy is attached (exact ARN match)
		expectedAdminPolicyARN := GetAdminPolicyARN(creator.Partition, creator.AccountID, roleName, rolePath)
		if containsPolicyARN(attachedPolicies, expectedAdminPolicyARN) {
			// Self-heal: admin policy exists but tag is missing
			log.LogDebug("Admin policy found but tag missing - adding tag")
			err = awsClient.TagRole(roleName, map[string]string{AdminRole: True})
			if err != nil {
				return "", true, fmt.Errorf("failed to add admin role tag: %w", err)
			}
			return roleARN, true, nil
		}

		return roleARN, true, fmt.Errorf("%w: role is standard, requested admin", ErrRoleExistsWrongProfile)

	case ProfileNoConsole:
		if isExistingRoleAdmin {
			return roleARN, true, fmt.Errorf("the existing role is an admin role." +
				" To use no-console permissions please delete the role and recreate it")
		}

		if isExistingRoleNoConsole {
			// Verify no-console policy is actually attached
			attachedPolicies, err := awsClient.ListAttachedRolePolicies(roleName)
			if err != nil {
				return "", true, err
			}

			expectedNoConsolePolicyARN := GetNoConsolePolicyARN(
				creator.Partition, creator.AccountID, roleName, rolePath)
			if !containsPolicyARN(attachedPolicies, expectedNoConsolePolicyARN) {
				// Tag exists but policy is missing - incomplete manual run
				return "", true, fmt.Errorf("the role has the no-console tag but the no-console policy is not attached." +
					" Please attach the policy or remove the tag and recreate the role")
			}

			return roleARN, true, nil
		}

		// Role appears to be standard - check if no-console policy is actually attached (self-healing)
		attachedPolicies, err := awsClient.ListAttachedRolePolicies(roleName)
		if err != nil {
			return "", true, err
		}

		// Check if no-console policy is attached (exact ARN match)
		expectedNoConsolePolicyARN := GetNoConsolePolicyARN(creator.Partition, creator.AccountID, roleName, rolePath)
		if containsPolicyARN(attachedPolicies, expectedNoConsolePolicyARN) {
			// Self-heal: no-console policy exists but tag is missing
			log.LogDebug("No-console policy found but tag missing - adding tag")
			err = awsClient.TagRole(roleName, map[string]string{NoConsoleRole: True})
			if err != nil {
				return "", true, fmt.Errorf("failed to add no-console role tag: %w", err)
			}
			return roleARN, true, nil
		}

		// Existing is standard, cannot convert
		return roleARN, true, fmt.Errorf("the existing role is a standard role." +
			" To use no-console permissions please delete the role and recreate it")

	default:
		return "", false, fmt.Errorf("invalid profile: %s (must be one of: %s, %s, %s)",
			profile, ProfileStandard, ProfileAdmin, ProfileNoConsole)
	}
}

// CreateRole creates an OCM IAM role with the specified profile and policies
func CreateRole(
	awsClient AWSClientInterface,
	creator *Creator,
	prefix string,
	roleName string,
	rolePath string,
	permissionsBoundary string,
	orgID string,
	env string,
	profile RoleProfile,
	policies map[string]*cmv1.AWSSTSPolicy,
	managedPolicies bool,
) (string, error) {
	var policyARN string
	var err error

	if profile != ProfileNoConsole {
		if managedPolicies {
			policyARN, err = GetManagedPolicyARN(policies, fmt.Sprintf("sts_%s_permission_policy", OCMRolePolicyFile))
			if err != nil {
				return "", err
			}
		} else {
			policyARN = GetPolicyArnWithSuffix(creator.Partition, creator.AccountID, roleName, rolePath)
		}
	}

	// Build trust policy
	filename := fmt.Sprintf("sts_%s_trust_policy", OCMRolePolicyFile)
	policyDetail := GetPolicyDetails(policies, filename)
	policy := InterpolatePolicyDocument(creator.Partition, policyDetail, map[string]string{
		"partition":           creator.Partition,
		"aws_account_id":      GetJumpAccount(env),
		"ocm_organization_id": orgID,
	})

	// Build IAM tags
	iamTags := map[string]string{
		RolePrefix:    prefix,
		RoleType:      OCMRole,
		Environment:   env,
		RedHatManaged: True,
	}
	if managedPolicies {
		iamTags[common.ManagedPolicies] = True
	}

	// Validate profile before creating any AWS resources
	if profile != ProfileStandard && profile != ProfileAdmin && profile != ProfileNoConsole {
		return "", fmt.Errorf("invalid profile: %s (must be one of: %s, %s, %s)",
			profile, ProfileStandard, ProfileAdmin, ProfileNoConsole)
	}

	log.LogDebug("Creating role '%s'", roleName)

	role, err := awsClient.CreateRole(roleName, policy, permissionsBoundary, iamTags, rolePath)
	if err != nil {
		return "", err
	}
	roleARN := *role.Arn
	log.LogInfo("Created role '%s' with ARN '%s'", roleName, roleARN)

	switch profile {
	case ProfileStandard:
		filename = fmt.Sprintf("sts_%s_permission_policy", OCMRolePolicyFile)
		policyDetail = GetPolicyDetails(policies, filename)
		err = createAndAttachPermissionPolicy(awsClient, policyARN, iamTags, roleName, rolePath, policyDetail, managedPolicies)
		if err != nil {
			return "", err
		}

	case ProfileAdmin:
		// standard policy first
		filename = fmt.Sprintf("sts_%s_permission_policy", OCMRolePolicyFile)
		policyDetail = GetPolicyDetails(policies, filename)
		err = createAndAttachPermissionPolicy(awsClient, policyARN, iamTags, roleName, rolePath, policyDetail, managedPolicies)
		if err != nil {
			return "", err
		}

		// create and attach the admin policy to the role
		filename = fmt.Sprintf("sts_%s_permission_policy", OCMAdminRolePolicyFile)
		if managedPolicies {
			policyARN, err = GetManagedPolicyARN(policies, filename)
			if err != nil {
				return "", err
			}
		} else {
			policyARN = GetAdminPolicyARN(creator.Partition, creator.AccountID, roleName, rolePath)
		}
		iamTags[AdminRole] = True
		policyDetail = GetPolicyDetails(policies, filename)
		err = createAndAttachPermissionPolicy(awsClient, policyARN, iamTags, roleName, rolePath, policyDetail, managedPolicies)
		if err != nil {
			return "", err
		}

		// tag role with admin tag
		err = awsClient.TagRole(roleName, map[string]string{AdminRole: True})
		if err != nil {
			return "", err
		}

	case ProfileNoConsole:
		filename = fmt.Sprintf("sts_%s_permission_policy", OCMNoConsoleRolePolicyFile)

		// create and attach the no-console policy to the role
		if managedPolicies {
			policyARN, err = GetManagedPolicyARN(policies, filename)
			if err != nil {
				return "", err
			}
		} else {
			policyARN = GetNoConsolePolicyARN(creator.Partition, creator.AccountID, roleName, rolePath)
		}
		iamTags[NoConsoleRole] = True
		policyDetail = GetPolicyDetails(policies, filename)
		err = createAndAttachPermissionPolicy(awsClient, policyARN, iamTags, roleName, rolePath, policyDetail, managedPolicies)
		if err != nil {
			return "", err
		}

		// tag role with no-console tag
		err = awsClient.TagRole(roleName, map[string]string{NoConsoleRole: True})
		if err != nil {
			return "", err
		}

	default:
		return "", fmt.Errorf("invalid profile: %s (must be one of: %s, %s, %s)",
			profile, ProfileStandard, ProfileAdmin, ProfileNoConsole)
	}

	return roleARN, nil
}

// createAndAttachPermissionPolicy creates and attaches a permission policy to an IAM role
func createAndAttachPermissionPolicy(
	awsClient AWSClientInterface,
	policyARN string,
	iamTags map[string]string,
	roleName string,
	rolePath string,
	policyDetail string,
	managedPolicies bool,
) error {
	log.LogDebug("Creating permission policy '%s'", policyARN)
	if !managedPolicies {
		// Extract policy name from ARN
		policyName := getPolicyNameFromARN(policyARN)
		_, err := awsClient.CreateIAMPolicy(policyName, policyDetail, iamTags)
		if err != nil {
			return err
		}
	}

	log.LogDebug("Attaching permission policy to role '%s'", roleName)
	err := awsClient.AttachIAMPolicy(roleName, policyARN)
	if err != nil {
		return err
	}

	return nil
}

// Helper functions

// isAdminRole checks if a role has the admin role tag
func isAdminRole(role *types.Role) (bool, error) {
	if role == nil {
		return false, nil
	}

	for _, tag := range role.Tags {
		if tag.Key != nil && *tag.Key == AdminRole && tag.Value != nil && *tag.Value == True {
			return true, nil
		}
	}

	return false, nil
}

// isNoConsoleRole checks if a role has the no-console role tag
func isNoConsoleRole(role *types.Role) (bool, error) {
	if role == nil {
		return false, nil
	}

	for _, tag := range role.Tags {
		if tag.Key != nil && *tag.Key == NoConsoleRole && tag.Value != nil && *tag.Value == True {
			return true, nil
		}
	}

	return false, nil
}

// containsPolicyARN checks if a policy ARN exists in the list of attached policies
func containsPolicyARN(attachedPolicies []types.AttachedPolicy, targetARN string) bool {
	arns := make([]string, 0, len(attachedPolicies))
	for _, policy := range attachedPolicies {
		if policy.PolicyArn != nil {
			arns = append(arns, *policy.PolicyArn)
		}
	}
	return slices.Contains(arns, targetARN)
}

// getPolicyNameFromARN extracts the policy name from a policy ARN
// Example: arn:aws:iam::123456789012:policy/MyRole-Policy -> MyRole-Policy
func getPolicyNameFromARN(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return arn
}
