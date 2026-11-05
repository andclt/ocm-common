#
# Copyright (c) 2023 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
include boilerplate/generated-includes.mk

# Ensure go modules are enabled:
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

# Disable CGO so that we always generate static binaries:
export CGO_ENABLED=0

# Unset GOFLAG for CI and ensure we've got nothing accidently set
unexport GOFLAGS

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test -coverprofile=cover.out  ./...

.PHONY: fmt
fmt:
	gofmt -s -l -w pkg

.PHONY: lint
lint:
	golangci-lint run --timeout 10m0s

.PHONY: clean
clean:
	rm -rf \
		ocm-common \
		$(NULL)

.PHONY: validate-commits
validate-commits:
	@echo "Validating commit messages..."
	@git log --pretty=format:%s origin/main..HEAD | while IFS= read -r message; do \
		echo "Validating: $$message"; \
		if echo "$$message" | grep -qE '^Merge (commit|pull request|branch)'; then \
			echo "  Skipping merge commit"; \
			continue; \
		fi; \
		if ! echo "$$message" | grep -qE '^[A-Z]+-[0-9]+ \| (feat|fix|refactor|chore|ci|docs|test|perf|style): .+$$'; then \
			echo "ERROR: Invalid commit message format"; \
			echo "Expected: JIRA-TICKET | TYPE: MESSAGE"; \
			echo "Example: OCMCOMM-123 | feat: add new feature"; \
			echo ""; \
			echo "Valid types: feat, fix, refactor, chore, ci, docs, test, perf, style"; \
			exit 1; \
		fi; \
		echo "  ✓ Valid"; \
	done
	@echo "All commits validated successfully"

.PHONY: validate-version
validate-version:
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=v0.1.0)
endif
	@# Validate that we're on the main branch
	@CURRENT_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$CURRENT_BRANCH" != "main" ]; then \
		echo "Error: Tags must be created from the main branch. Currently on $$CURRENT_BRANCH"; \
		exit 1; \
	fi
	@# Validate version format (vX.Y.Z)
	@if ! echo "$(VERSION)" | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' > /dev/null; then \
		echo "Error: VERSION must follow the format vX.Y.Z (e.g., v0.1.0)"; \
		exit 1; \
	fi
	@# Check if tag already exists
	@if git tag -l | grep -q "^$(VERSION)$$"; then \
		echo "Error: Tag $(VERSION) already exists"; \
		exit 1; \
	fi
	@# Get the latest tag and compare versions
	@LATEST_TAG=$$(git tag -l 'v*.*.*' | sort -V | tail -1); \
	if [ -n "$$LATEST_TAG" ]; then \
		if ! printf '%s\n%s\n' "$$LATEST_TAG" "$(VERSION)" | sort -V -C; then \
			echo "Error: VERSION $(VERSION) must be higher than the latest tag $$LATEST_TAG"; \
			exit 1; \
		fi; \
	fi

.PHONY: release
release: validate-version
	@echo "Creating and pushing release $(VERSION)..."
	git tag -a -m 'Release $(VERSION)' $(VERSION)
	git push upstream $(VERSION)
	@echo "Release $(VERSION) created and pushed successfully!"

.PHONY: boilerplate-update
boilerplate-update:
	@boilerplate/update
