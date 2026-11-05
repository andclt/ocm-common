# Contributing to OCM-Common
Welcome, and thank you for considering to contribute to OCM-Common.
Before you begin, or have more questions reach out to us on [Slack](https://redhat.enterprise.slack.com/archives/C0AKBJST8QH)

## Contributing Code
To contribute bug fixes or features to OCM-Common:

- Communicate your intent.
- Make your changes.
- Test your changes.
- Open a Pull Request (PR).

Communicate your intent in the form of a JIRA ticket on the [ROSAENG](https://redhat.atlassian.net/browse/ROSAENG) project.

Be sure to practice good git commit hygiene as you make your changes. All but the smallest changes should be broken up
into a few commits that tell a story. Use your git commits to provide context for the folks who will review PR. We strive
to follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/#summary).

The commit message should follow this template:
```shell
<type>[JIRA-TICKET] | [TYPE]: <MESSAGE>

[optional BODY]

[optional FOOTER(s)]
```

The commit contains the following structural types, to communicate your intent:

- `fix:` a commit of the type fix patches a bug in your codebase (this correlates with PATCH in Semantic Versioning).
- `feat:` a commit of the type feat introduces a new feature to the codebase (this correlates with MINOR in Semantic
  Versioning).

Types other than `fix:` and `feat:` are allowed:
- `build`: Changes that affect the build system or external dependencies
- `ci`: Changes to our CI configuration files and scripts
- `docs`: Documentation only changes
- `perf`: A code change that improves performance
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `style`: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- `test`: Adding missing tests or correcting existing tests

All code should be covered by tests. We use [Ginkgo](https://onsi.github.io/ginkgo/). Other third party testing packages
will be rejected.

Once you made and tested your changes, create a pull request (PR). In the PR `overview` please link the
jira ticket associated with your change. This should follow the format `JIRA: ROSAENG-xxxx`. Note the key word `JIRA`,
use of any other key word may result in the bot performing unwanted action to the ticket in JIRA. Please also include in the
`overview` any additional information not in the JIRA that may help set context around your intent. Also include any extra
validation steps which may help reviews to validate the changes.

We work on a Sprint basis, all changes should be tracked by a JIRA. When being worked on these changes should be added to
the current SDA sprint. These sprints are denoted as `SDA - Sprint xxx`. The workflow should be as follows:
- `Todo` will complete in this current sprint
- `In Progress` ticket is currently being worked on
- `Code Review` PR has been created and is being reviewed by the team.
- `Review` Once the changes have been merged, move ticket to `Review` a QE person will be assigned to the ticket and tested.
- `Done` Once QE are satisfied with the change and bugs have been fixed the QE person assigned to your ticket will mark it
  as done

During `Review`, remain assigned to the ticket, so that QE knows who to assign any follow-up bugs to during testing. You
will also be asked to review test cases for this change. QE will supply you with a link to the test cases, if you approve
the test cases add `tc-approved` label to the JIRA, if you need changes to the test cases work with QE in the JIRA comments
to resolve these changes.

## GitHub Workflows

This repository uses GitHub actions which are configured at `./github/workflows`

## Releasing a new OCM-Common Version

Releases are handled via GitHub action and goreleaser. Simply push a version tag to create a release.

Use the Makefile `release` target to create and push a new version tag:

```shell
git checkout main
git pull
make release VERSION=v0.1.39
```

The `make release` target will:
- Validate that you're on the `main` branch
- Check that the version follows semantic versioning format (vX.Y.Z)
- Ensure the new version is higher than the latest existing tag
- Create an annotated tag with a proper release message
- Push the tag to the `upstream` remote

Note that a repository administrator may need to push the tag to the repository due to access restrictions.

## Boilerplate

This repository subscribes to the `openshift/rosa-services` boilerplate convention in
[`boilerplate/update.cfg`](boilerplate/update.cfg). It supplies the shared root
`OWNERS_ALIASES`; this repository maintains its own `OWNERS` files, which can
reference those aliases. See the [convention README](boilerplate/openshift/rosa-services/README.md)
for details.

To pull in changes to the convention (including alias membership), run
`make boilerplate-update` from a clean checkout, review the resulting changes,
and commit them. Update shared aliases in the upstream
[openshift/boilerplate](https://github.com/openshift/boilerplate) repository rather
than editing the copied `OWNERS_ALIASES` or managed boilerplate files here.

# Questions?

If you have any questions about the code or how to contribute, don't hesitate to
[open an issue](https://github.com/openshift-online/ocm-common/issues/new) in this repo
