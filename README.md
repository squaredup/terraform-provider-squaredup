# SquaredUp Terraform Provider

[![Tests](https://github.com/squaredup/terraform-provider-squaredup/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/squaredup/terraform-provider-squaredup/actions/workflows/test.yml) [![Registry](https://img.shields.io/badge/squaredup-Terraform%20Registry-purple)](https://registry.terraform.io/providers/squaredup/squaredup/latest)

___
A Terraform provider for managing [SquaredUp](https://app.squaredup.com/)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.2
- [Go](https://golang.org/doc/install) >= 1.21
- [GoReleaser](https://goreleaser.com/) >= 0.153.x

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

Usage documentation is available on [Terraform Registry](https://registry.terraform.io/providers/squaredup/squaredup/latest/docs)

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources.

```shell
make testacc
```
