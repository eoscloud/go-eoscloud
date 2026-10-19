# Contributing

Thanks for your interest in go-eoscloud.

## Before you start

- **Bugs and feature requests:** open an issue using one of the forms.
  Please search existing issues first.
- **Security issues:** never open a public issue. See [SECURITY.md](SECURITY.md).
- **Questions and support:** see
  [SUPPORT.md](https://github.com/eoscloud/.github/blob/main/SUPPORT.md).
- **Larger changes:** open an issue first so we can agree on the approach
  before you invest time in a pull request.

## Development

You need Go 1.27 or newer and [pre-commit](https://pre-commit.com/).

```sh
pre-commit install --install-hooks
go test ./...
```

The pre-commit hooks run `go mod tidy` and golangci-lint. CI runs the same
checks on every pull request.

The SDK is organized by API: one package per service under `eoscloud/`,
with shared types in `eoscloud/common/model` and the HTTP client in
`pkg/api`. Keep request and response types close to the API schema.

## Pull requests

- Keep pull requests small and focused on one change.
- Add or update tests for the behavior you change.
- Update the documentation and examples if users will notice the change.
- Reference the issue in the description, e.g. `Fixes #123`.
- A maintainer reviews every pull request. We may ask for changes or
  squash commits when merging.

## Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>
```

| Type    | Use for                                 |
| ------- | --------------------------------------- |
| `feat`  | A new feature or user-facing capability |
| `fix`   | A bug fix                               |
| `docs`  | Documentation only                      |
| `test`  | Tests only                              |
| `ci`    | CI configuration                        |
| `chore` | Dependency updates and maintenance      |

The scope is optional. Release notes are generated from these messages, so
write the description for users.

## Developer Certificate of Origin

By contributing, you certify that you wrote the change or otherwise have
the right to submit it under the project's license, as described in the
[Developer Certificate of Origin](https://developercertificate.org/).
Sign off every commit:

```sh
git commit -s
```

This adds a `Signed-off-by: Your Name <you@example.com>` line. A CI check
rejects pull requests with commits that are not signed off.

## License

By contributing, you agree that your contributions are licensed under the
Apache License, Version 2.0, the license of the repository.

## Code of Conduct

This project follows our
[Code of Conduct](https://github.com/eoscloud/.github/blob/main/CODE_OF_CONDUCT.md). By
participating you agree to uphold it.
