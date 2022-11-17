# Firetail Lambda Extension

[![License: LGPL v3](https://img.shields.io/badge/License-LGPL_v3-blue.svg)](https://www.gnu.org/licenses/lgpl-3.0) [![Test and coverage](https://github.com/FireTail-io/firetail-lambda-extension/actions/workflows/codecov.yml/badge.svg?branch=defaults)](https://github.com/FireTail-io/firetail-lambda-extension/actions/workflows/codecov.yml) [![codecov](https://codecov.io/gh/FireTail-io/firetail-lambda-extension/branch/main/graph/badge.svg?token=QNWMOGA31B)](https://codecov.io/gh/FireTail-io/firetail-lambda-extension)



## Overview

The Firetail Logging Extension receives AWS Lambda events & response payloads and sends them to the Firetail Logging API.

The extension receives these events and response payloads via a runtime-specific Firetail library which you will need to use in your Function code. The Firetail library outputs specifically formatted logs which the extension then receives via the [Lambda Logs API](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-logs-api.html). You can find a table of Firetail function libraries which correspond with a Lambda runtime in the [Function Libraries](#function-libraries) section.



## Function Libraries

| Supported Runtimes   | Library                                                      |
| -------------------- | ------------------------------------------------------------ |
| Python 3.7, 3.8, 3.9 | [github.com/FireTail-io/firetail-py-lambda](https://github.com/FireTail-io/firetail-py-lambda) |



## Tests

Automated testing is setup with the `testing` package, using [github.com/stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify) for shorthand assertions. You can run them with `go test`, or use the provided [Makefile](./Makefile)'s `test` target, which is as simple as:

```bash
make test
```

This will output a coverage report which you may use to view the test coverage in your browser by using the `go tool` command:

```bash
go tool cover -html coverage.out
```



## Deployment

The Firetail Logging Extension is an external Lambda extension, published as a Lambda Layer. Deploying it is a three step process:

- The first step is to [build the extension binary](#building-the-extension-binary).
- The second step is to [package the extension binary](#packaging-the-extension-binary).
- The third step is to [publish the package as a Lambda Layer](#publishing-the-package).
- An optional step is to [make the layer public](#making-the-layer-public).
- The final step is to [add the layer to a Lambda Function](#adding-the-layer-to-a-lambda-function).

This process has been partially automated in the provided [Makefile](./Makefile). In order to use this makefile you will need to install the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html), [Golang](https://go.dev/doc/install) and [JQ](https://stedolan.github.io/jq/). You may observe how this Makefile is used by us in the Github action named "[build & publish](./.github/workflows/release.yaml)".



### Building The Extension Binary

The logging extension is a standard Go project and can be built by [installing Go](https://go.dev/doc/install) and using the `go build` command from the root directory of this repository. You will need to set the `GOOS` and `GOARCH` environment variables appropriately for your target lambda runtime's operating system and architecture. See the [Environment variables](https://pkg.go.dev/cmd/go#hdr-Environment_variables) section of the [go command docs](https://pkg.go.dev/cmd/go) for more information.

```bash
GOOS=linux GOARCH=amd64 go build
```

This will yield a binary with the same name as the root directory, which if you have just cloned this repository will be `firetail-lambda-extension`.

The target in the provided makefile that corresponds to this step is `build`. It requires a target architecture (`ARCH`) which defaults to `amd64`. For example, you may wish to do:

```bash
make build ARCH=arm64
```

This will yield a `build` and `build/extensions` directory, and a binary within `build/extensions` named `firetail-extension-${ARCH}`.



### Packaging The Extension Binary

To package the extension binary, it must be placed into a directory named `extensions` and then zipped.

> During the `Init` phase, Lambda extracts layers containing extensions into the `/opt` directory in the execution environment. Lambda looks for extensions in the `/opt/extensions/` directory, interprets each file as an executable bootstrap for launching the extension, and starts all extensions in parallel.
>
> [Source](https://docs.aws.amazon.com/lambda/latest/dg/lambda-extensions.html)

The target in the provided makefile that corresponds to this step is `package`, and it depends upon the `build` step. It requries a target architecture (`ARCH`), and extension version (`VERSION`) which defaults to `latest`. For example, you may wish to do:

```bash
make package ARCH=arm64 VERSION=v1.0.0
```

This will yield a `.zip` file in the `build` directory named `firetail-extension-${ARCH}-${VERSION}.zip`, which contains the `extensions` directory and the binary within it such that when it is extracted into `/opt`, the extension binary will be found in the `/opt/extensions/` directory as per the AWS documentation.



### Publishing The Package

To publish the package, you may use the AWS CLI's [publish-layer-version](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/lambda/publish-layer-version.html) command. You will need to repeat this process for every region in which you wish to use the layer. You will also need to specify the compatible architectures, and give the layer a name. The output of the command will provide you with the layer's ARN and layer version, which you may use to add it to your Lambdas.

If you reuse the same layer name multiple times, the layer version will be incremented. The approach taken in the provided makefile is to publish each extension version with a new layer name, so the layer version will almost always be `1`.

The target in the provided makefile that corresponds to this step is `publish`. You must make the `build` target before the `publish` target. The `publish` target requires a target architecture (`ARCH`) and extension version (`VERSION`), which match that used when you made the `package` target; and a region in which to publish the layer (`AWS_REGION`). For example, you may wish to do:

```bash
make publish ARCH=arm64 VERSION=v1.0.0 AWS_REGION=eu-west-1
```



### Making The Layer Public

ℹ️ In this step, we make the layer publically available for anyone to use. You may wish to omit this step. 

To make the layer public, you may use the AWS CLI's [add-layer-version-permission](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/lambda/add-layer-version-permission.html) command. You will need to repeat this process for every layer you publish in every region. You will need to provide the layer name & layer version, a statement ID and region; and to make the layer public an action of `lambda:GetLayerVersion` and principal of `*`.

The target in the provided makefile corresponding to this step is `public`. You must make the `publish` target before the `public` target. The `public` target requires a target architecture (`ARCH`), extension version (`VERSION`) and AWS region (`AWS_REGION`) which match that used when you made the `publish` target, as well as the layer version created when you made the `publish` target (`AWS_LAYER_VERSION`). For example, you may wish to do:

```bash
make public ARCH=arm64 VERSION=v1.0.0 AWS_REGION=eu-west-1 AWS_LAYER_VERSION=1
```



### Adding The Layer To A Lambda Function

To add the Lambda Layer to a Function, you may use the AWS CLI's [update-function-configuration](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/lambda/update-function-configuration.html) command. You will need to provide a region, the layer ARN and the name of the Function to which the Layer is to be added.

The target in the provided makefile corresponding to this step is `add`. The `add` target requires the Layer ARN (`LAYER_ARN`), the name of the Function to add the Layer to (`FUNCTION_NAME`), and the AWS region in which both the Layer and the Function must be found (`AWS_REGION`). For example, you may wish to do:

```bash
make add AWS_REGION=eu-west-1 LAYER_ARN=your-layer-arn FUNCTION_NAME=your-function-name
```

