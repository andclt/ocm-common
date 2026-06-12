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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	. "github.com/openshift-online/ocm-common/pkg/rosa/ocmrole"
)

var _ = Describe("OCM Role Helpers", func() {
	Describe("GetOCMRoleName", func() {
		It("should construct correct role name", func() {
			name := GetOCMRoleName("ManagedOpenShift", "OCM", "12345678")
			Expect(name).To(Equal("ManagedOpenShift-OCM-Role-12345678"))
		})

		It("should truncate long role names", func() {
			// AWS role names have a 64 character limit
			longPrefix := "VeryLongPrefixNameThatExceedsTheMaximumLengthAllowedByAWS"
			name := GetOCMRoleName(longPrefix, "OCM", "12345678")
			Expect(len(name)).To(BeNumerically("<=", 64))
		})
	})

	Describe("Policy ARN builders", func() {
		It("should build standard policy ARN correctly", func() {
			arn := GetPolicyArnWithSuffix("aws", "123456789012", "test-role", "/")
			Expect(arn).To(Equal("arn:aws:iam::123456789012:policy/test-role-Policy"))
		})

		It("should build admin policy ARN correctly", func() {
			arn := GetAdminPolicyARN("aws", "123456789012", "test-role", "/")
			Expect(arn).To(Equal("arn:aws:iam::123456789012:policy/test-role-Admin-Policy"))
		})

		It("should build no-console policy ARN correctly", func() {
			arn := GetNoConsolePolicyARN("aws", "123456789012", "test-role", "/")
			Expect(arn).To(Equal("arn:aws:iam::123456789012:policy/test-role-NoConsole-Policy"))
		})

		It("should handle custom paths correctly", func() {
			arn := GetAdminPolicyARN("aws", "123456789012", "test-role", "/custom/path/")
			Expect(arn).To(Equal("arn:aws:iam::123456789012:policy/custom/path/test-role-Admin-Policy"))
		})

		It("should handle GovCloud partition", func() {
			arn := GetPolicyArnWithSuffix("aws-us-gov", "123456789012", "test-role", "/")
			Expect(arn).To(Equal("arn:aws-us-gov:iam::123456789012:policy/test-role-Policy"))
		})
	})

	Describe("GetJumpAccount", func() {
		It("should return correct account for production", func() {
			account := GetJumpAccount("production")
			Expect(account).To(Equal("710019948333"))
		})

		It("should return correct account for staging", func() {
			account := GetJumpAccount("staging")
			Expect(account).To(Equal("644306948063"))
		})

		It("should return correct account for integration", func() {
			account := GetJumpAccount("integration")
			Expect(account).To(Equal("896164604406"))
		})
	})

	Describe("DetermineProfile", func() {
		It("should return admin profile when isAdmin is true", func() {
			profile := DetermineProfile(true, false)
			Expect(profile).To(Equal(ProfileAdmin))
		})

		It("should return no-console profile when isNoConsole is true", func() {
			profile := DetermineProfile(false, true)
			Expect(profile).To(Equal(ProfileNoConsole))
		})

		It("should return standard profile when both are false", func() {
			profile := DetermineProfile(false, false)
			Expect(profile).To(Equal(ProfileStandard))
		})

		It("should prefer admin over no-console when both are true", func() {
			profile := DetermineProfile(true, true)
			Expect(profile).To(Equal(ProfileAdmin))
		})
	})

	Describe("GetPolicyDetails", func() {
		var policies map[string]*cmv1.AWSSTSPolicy

		BeforeEach(func() {
			policy1, _ := cmv1.NewAWSSTSPolicy().
				ID("test-policy-1").
				Details(`{"Version":"2012-10-17"}`).
				Build()
			policy2, _ := cmv1.NewAWSSTSPolicy().
				ID("test-policy-2").
				Details(`{"Statement":[]}`).
				Build()

			policies = map[string]*cmv1.AWSSTSPolicy{
				"test-policy-1": policy1,
				"test-policy-2": policy2,
			}
		})

		It("should return policy details when key exists", func() {
			details := GetPolicyDetails(policies, "test-policy-1")
			Expect(details).To(Equal(`{"Version":"2012-10-17"}`))
		})

		It("should return empty string when key doesn't exist", func() {
			details := GetPolicyDetails(policies, "non-existent")
			Expect(details).To(BeEmpty())
		})
	})

	Describe("GetManagedPolicyARN", func() {
		var policies map[string]*cmv1.AWSSTSPolicy

		BeforeEach(func() {
			managedPolicy, _ := cmv1.NewAWSSTSPolicy().
				ID("managed-policy").
				ARN("arn:aws:iam::aws:policy/managed-policy").
				Build()
			unmanagedPolicy, _ := cmv1.NewAWSSTSPolicy().
				ID("unmanaged-policy").
				Details(`{"Version":"2012-10-17"}`).
				Build()

			policies = map[string]*cmv1.AWSSTSPolicy{
				"managed-policy":   managedPolicy,
				"unmanaged-policy": unmanagedPolicy,
			}
		})

		It("should return ARN for managed policy", func() {
			arn, err := GetManagedPolicyARN(policies, "managed-policy")
			Expect(err).ToNot(HaveOccurred())
			Expect(arn).To(Equal("arn:aws:iam::aws:policy/managed-policy"))
		})

		It("should error when policy doesn't exist", func() {
			_, err := GetManagedPolicyARN(policies, "non-existent")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should error when policy has no ARN", func() {
			_, err := GetManagedPolicyARN(policies, "unmanaged-policy")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not have a managed ARN"))
		})
	})

	Describe("InterpolatePolicyDocument", func() {
		It("should replace placeholders with values", func() {
			doc := `{"Statement":[{"Resource":"%{bucket_arn}"}]}`
			result := InterpolatePolicyDocument("aws", doc, map[string]string{
				"bucket_arn": "arn:aws:s3:::my-bucket",
			})
			Expect(result).To(ContainSubstring("arn:aws:s3:::my-bucket"))
			Expect(result).ToNot(ContainSubstring("%{bucket_arn}"))
		})

		It("should replace multiple placeholders", func() {
			doc := `{"Statement":[{"Principal":"%{principal}","Resource":"%{resource}"}]}`
			result := InterpolatePolicyDocument("aws", doc, map[string]string{
				"principal": "arn:aws:iam::123456789012:role/test",
				"resource":  "arn:aws:s3:::my-bucket/*",
			})
			Expect(result).To(ContainSubstring("arn:aws:iam::123456789012:role/test"))
			Expect(result).To(ContainSubstring("arn:aws:s3:::my-bucket/*"))
		})

		It("should handle partition replacement for legacy policies", func() {
			doc := `{"Statement":[{"Resource":"arn:aws:s3:::bucket"}]}`
			result := InterpolatePolicyDocument("aws-us-gov", doc, map[string]string{})
			Expect(result).To(ContainSubstring("arn:aws-us-gov:s3:::bucket"))
			Expect(result).ToNot(ContainSubstring("arn:aws:s3:::bucket"))
		})

		It("should not replace partition when already correct", func() {
			doc := `{"Statement":[{"Resource":"arn:aws-us-gov:s3:::bucket"}]}`
			result := InterpolatePolicyDocument("aws-us-gov", doc, map[string]string{})
			Expect(result).To(Equal(doc))
		})
	})
})
