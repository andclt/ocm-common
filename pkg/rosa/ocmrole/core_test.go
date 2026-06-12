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

package ocmrole_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"go.uber.org/mock/gomock"

	. "github.com/openshift-online/ocm-common/pkg/rosa/ocmrole"
)

// Helper to create a role with tags
func makeRole(arn, name string, tags map[string]string) *types.Role {
	roleTags := []types.Tag{}
	for k, v := range tags {
		key := k
		val := v
		roleTags = append(roleTags, types.Tag{Key: &key, Value: &val})
	}
	return &types.Role{
		Arn:      &arn,
		RoleName: &name,
		Tags:     roleTags,
	}
}

var _ = Describe("CheckRoleExists", func() {
	var (
		ctrl       *gomock.Controller
		mockClient *MockAWSClientInterface
		creator    *Creator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClient = NewMockAWSClientInterface(ctrl)
		creator = &Creator{
			AccountID: "123456789012",
			Partition: "aws",
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("when role does not exist", func() {
		It("should return exists=false", func() {
			mockClient.EXPECT().GetRole("test-role").Return(nil, fmt.Errorf("role not found"))

			roleARN, exists, err := CheckRoleExists(mockClient, creator, "test-role", ProfileStandard, "/")

			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())
			Expect(roleARN).To(BeEmpty())
		})
	})

	Context("when requesting standard profile", func() {
		It("should succeed if existing role is standard", func() {
			mockClient.EXPECT().GetRole("test-role").Return(
				makeRole("arn:aws:iam::123456789012:role/test-role", "test-role", map[string]string{}), nil)

			roleARN, exists, err := CheckRoleExists(mockClient, creator, "test-role", ProfileStandard, "/")

			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())
			Expect(roleARN).To(Equal("arn:aws:iam::123456789012:role/test-role"))
		})

		It("should error if existing role is admin", func() {
			mockClient.EXPECT().GetRole("test-role").Return(
				makeRole("arn:aws:iam::123456789012:role/test-role", "test-role",
					map[string]string{AdminRole: "true"}), nil)

			roleARN, exists, err := CheckRoleExists(mockClient, creator, "test-role", ProfileStandard, "/")

			Expect(err).To(HaveOccurred())
			Expect(exists).To(BeTrue())
			Expect(roleARN).To(Equal("arn:aws:iam::123456789012:role/test-role"))
			Expect(err.Error()).To(ContainSubstring("the existing role is an admin role"))
		})

		It("should error if existing role is no-console", func() {
			mockClient.EXPECT().GetRole("test-role").Return(
				makeRole("arn:aws:iam::123456789012:role/test-role", "test-role",
					map[string]string{NoConsoleRole: "true"}), nil)

			_, exists, err := CheckRoleExists(mockClient, creator, "test-role", ProfileStandard, "/")

			Expect(err).To(HaveOccurred())
			Expect(exists).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring("the existing role is a no-console role"))
		})
	})

	Context("when requesting admin profile", func() {
		It("should succeed if existing role is admin (idempotent)", func() {
			mockClient.EXPECT().GetRole("test-role").Return(
				makeRole("arn:aws:iam::123456789012:role/test-role", "test-role",
					map[string]string{AdminRole: "true"}), nil)

			roleARN, exists, err := CheckRoleExists(mockClient, creator, "test-role", ProfileAdmin, "/")

			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())
			Expect(roleARN).To(Equal("arn:aws:iam::123456789012:role/test-role"))
		})

		It("should self-heal if admin policy exists but tag is missing", func() {
			mockClient.EXPECT().GetRole("test-role").Return(
				makeRole("arn:aws:iam::123456789012:role/test-role", "test-role", map[string]string{}), nil)

			adminPolicyArn := "arn:aws:iam::123456789012:policy/test-role-Admin-Policy"
			mockClient.EXPECT().ListAttachedRolePolicies("test-role").Return(
				[]types.AttachedPolicy{{PolicyArn: &adminPolicyArn}}, nil)

			mockClient.EXPECT().TagRole("test-role", gomock.Any()).DoAndReturn(
				func(roleName string, tags map[string]string) error {
					Expect(tags[AdminRole]).To(Equal("true"))
					return nil
				})

			roleARN, exists, err := CheckRoleExists(mockClient, creator, "test-role", ProfileAdmin, "/")

			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())
			Expect(roleARN).To(Equal("arn:aws:iam::123456789012:role/test-role"))
		})
	})

	Context("when requesting no-console profile", func() {
		It("should succeed if existing role is no-console", func() {
			mockClient.EXPECT().GetRole("test-role").Return(
				makeRole("arn:aws:iam::123456789012:role/test-role", "test-role",
					map[string]string{NoConsoleRole: "true"}), nil)

			noConsolePolicyArn := "arn:aws:iam::123456789012:policy/test-role-NoConsole-Policy"
			mockClient.EXPECT().ListAttachedRolePolicies("test-role").Return(
				[]types.AttachedPolicy{{PolicyArn: &noConsolePolicyArn}}, nil)

			roleARN, exists, err := CheckRoleExists(mockClient, creator, "test-role", ProfileNoConsole, "/")

			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())
			Expect(roleARN).To(Equal("arn:aws:iam::123456789012:role/test-role"))
		})

		It("should self-heal when no-console policy exists but tag is missing", func() {
			mockClient.EXPECT().GetRole("test-role").Return(
				makeRole("arn:aws:iam::123456789012:role/test-role", "test-role", map[string]string{}), nil)

			noConsolePolicyArn := "arn:aws:iam::123456789012:policy/test-role-NoConsole-Policy"
			mockClient.EXPECT().ListAttachedRolePolicies("test-role").Return(
				[]types.AttachedPolicy{{PolicyArn: &noConsolePolicyArn}}, nil)

			mockClient.EXPECT().TagRole("test-role", gomock.Any()).DoAndReturn(
				func(roleName string, tags map[string]string) error {
					Expect(tags[NoConsoleRole]).To(Equal("true"))
					return nil
				})

			roleARN, exists, err := CheckRoleExists(mockClient, creator, "test-role", ProfileNoConsole, "/")

			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())
			Expect(roleARN).To(Equal("arn:aws:iam::123456789012:role/test-role"))
		})

		It("should error when no-console tag exists but policy is not attached", func() {
			mockClient.EXPECT().GetRole("test-role").Return(
				makeRole("arn:aws:iam::123456789012:role/test-role", "test-role",
					map[string]string{NoConsoleRole: "true"}), nil)

			standardPolicyArn := "arn:aws:iam::123456789012:policy/test-role-Policy"
			mockClient.EXPECT().ListAttachedRolePolicies("test-role").Return(
				[]types.AttachedPolicy{{PolicyArn: &standardPolicyArn}}, nil)

			_, exists, err := CheckRoleExists(mockClient, creator, "test-role", ProfileNoConsole, "/")

			Expect(err).To(HaveOccurred())
			Expect(exists).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring("the role has the no-console tag but the no-console policy is not attached"))
		})
	})
})

var _ = Describe("CreateRole", func() {
	var (
		ctrl       *gomock.Controller
		mockClient *MockAWSClientInterface
		creator    *Creator
		policies   map[string]*cmv1.AWSSTSPolicy
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClient = NewMockAWSClientInterface(ctrl)
		creator = &Creator{
			AccountID: "123456789012",
			Partition: "aws",
		}

		// Create test policies
		trustPolicy, _ := cmv1.NewAWSSTSPolicy().
			ID("sts_ocm_trust_policy").
			Details(`{"Version":"2012-10-17","Statement":[]}`).
			Build()
		standardPolicy, _ := cmv1.NewAWSSTSPolicy().
			ID("sts_ocm_permission_policy").
			Details(`{"Version":"2012-10-17","Statement":[]}`).
			Build()
		adminPolicy, _ := cmv1.NewAWSSTSPolicy().
			ID("sts_ocm_admin_permission_policy").
			Details(`{"Version":"2012-10-17","Statement":[]}`).
			Build()
		noConsolePolicy, _ := cmv1.NewAWSSTSPolicy().
			ID("sts_ocm_no_console_permission_policy").
			Details(`{"Version":"2012-10-17","Statement":[]}`).
			Build()

		policies = map[string]*cmv1.AWSSTSPolicy{
			"sts_ocm_trust_policy":                 trustPolicy,
			"sts_ocm_permission_policy":            standardPolicy,
			"sts_ocm_admin_permission_policy":      adminPolicy,
			"sts_ocm_no_console_permission_policy": noConsolePolicy,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should create standard OCM role successfully", func() {
		roleName := "test-role"
		roleArn := fmt.Sprintf("arn:aws:iam::123456789012:role/%s", roleName)

		mockClient.EXPECT().CreateRole(gomock.Eq(roleName), gomock.Any(), gomock.Eq(""),
			gomock.Any(), gomock.Eq("/")).Return(
			types.Role{Arn: &roleArn, RoleName: &roleName}, nil)

		mockClient.EXPECT().CreateIAMPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			&types.Policy{}, nil)

		mockClient.EXPECT().AttachIAMPolicy(gomock.Eq(roleName), gomock.Any()).Return(nil)

		resultARN, err := CreateRole(mockClient, creator, "test-prefix", roleName, "/", "",
			"org-123", "production", ProfileStandard, policies, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(resultARN).To(Equal(roleArn))
	})

	It("should create admin OCM role with two policies and tag", func() {
		roleName := "test-role"
		roleArn := fmt.Sprintf("arn:aws:iam::123456789012:role/%s", roleName)

		mockClient.EXPECT().CreateRole(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return(
			types.Role{Arn: &roleArn, RoleName: &roleName}, nil)

		// Standard policy
		mockClient.EXPECT().CreateIAMPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			&types.Policy{}, nil)
		mockClient.EXPECT().AttachIAMPolicy(gomock.Any(), gomock.Any()).Return(nil)

		// Admin policy
		mockClient.EXPECT().CreateIAMPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			&types.Policy{}, nil)
		mockClient.EXPECT().AttachIAMPolicy(gomock.Any(), gomock.Any()).Return(nil)

		// Admin tag
		mockClient.EXPECT().TagRole(roleName, gomock.Any()).DoAndReturn(
			func(roleName string, tags map[string]string) error {
				Expect(tags[AdminRole]).To(Equal("true"))
				return nil
			})

		resultARN, err := CreateRole(mockClient, creator, "test-prefix", roleName, "/", "",
			"org-123", "production", ProfileAdmin, policies, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(resultARN).To(Equal(roleArn))
	})

	It("should create no-console OCM role with policy and tag", func() {
		roleName := "test-role"
		roleArn := fmt.Sprintf("arn:aws:iam::123456789012:role/%s", roleName)

		mockClient.EXPECT().CreateRole(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return(
			types.Role{Arn: &roleArn, RoleName: &roleName}, nil)

		mockClient.EXPECT().CreateIAMPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			&types.Policy{}, nil)
		mockClient.EXPECT().AttachIAMPolicy(gomock.Any(), gomock.Any()).Return(nil)

		mockClient.EXPECT().TagRole(roleName, gomock.Any()).DoAndReturn(
			func(roleName string, tags map[string]string) error {
				Expect(tags[NoConsoleRole]).To(Equal("true"))
				return nil
			})

		resultARN, err := CreateRole(mockClient, creator, "test-prefix", roleName, "/", "",
			"org-123", "production", ProfileNoConsole, policies, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(resultARN).To(Equal(roleArn))
	})

	It("should fail when role creation fails", func() {
		mockClient.EXPECT().CreateRole(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return(types.Role{}, fmt.Errorf("role creation failed"))

		_, err := CreateRole(mockClient, creator, "test-prefix", "test-role", "/", "",
			"org-123", "production", ProfileStandard, policies, false)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("role creation failed"))
	})
})
