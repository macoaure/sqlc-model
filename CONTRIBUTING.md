# Contributing to sqlc-gen-richmodel

Thanks for considering a contribution! This project is a Go module, and the workflow is standard Go tooling.

## Local setup

1. Install Go 1.24 or newer (`go version`).
2. Clone your fork and enter the directory:

   ```sh
   git clone https://github.com/<you>/sqlc-gen-richmodel.git
   cd sqlc-gen-richmodel
   ```

3. Download dependencies:

   ```sh
   go mod download
   ```

## Running tests

```sh
go test ./...
```

## Running lint

Install [`golangci-lint`](https://golangci-lint.run/welcome/install/), then:

```sh
golangci-lint run
```

CI runs the same build, vet, test, and lint checks on every pull request (see `.github/workflows/ci.yml`).

## Submitting a change

1. Open a pull request against `main`. The [pull request template](.github/pull_request_template.md) will pre-fill a description, related-issue link, and checklist.
2. Make sure `go build ./...`, `go test ./...`, and `golangci-lint run` all pass locally before requesting review.

## Reporting issues

Use the [Bug Report](.github/ISSUE_TEMPLATE/bug_report.yml) or [Feature Request](.github/ISSUE_TEMPLATE/feature_request.yml) templates when opening a new issue, or open a blank issue if neither fits.
