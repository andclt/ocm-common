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

// Tag constants used for OCM IAM roles
const (
	// Prefix used by all the tag names
	tagPrefix = "rosa_"

	// RoleType is the name of the tag that will contain the purpose of the role
	RoleType = tagPrefix + "role_type"

	// RolePrefix is the name of the tag that will contain the user-set prefix of the role
	RolePrefix = tagPrefix + "role_prefix"

	// Environment is the name of the tag that will contain the environment of the role
	Environment = tagPrefix + "environment"

	// AdminRole tags the role as admin (true/false)
	AdminRole = tagPrefix + "admin_role"

	// NoConsoleRole tags the role as no-console role
	NoConsoleRole = tagPrefix + "no_console_role"

	// RedHatManaged tags the role as red_hat_managed
	RedHatManaged = "red-hat-managed"

	// True is the string value for boolean tags
	True = "true"
)
