# Firetail Logging Extension

This extension is a proof of concept using the Lambda Logs API to extract request & responses and forward them to the Firetail SaaS.



## Tests

Automated testing is setup with the `testing` package, using [github.com/stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify) for shorthand assertions. You can run them with `go test`, or use the provided [Makefile](./Makefile)'s `build` target.



## Deployment

You will need to install the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html), [Golang](https://go.dev/doc/install) and [JQ](https://stedolan.github.io/jq/).

A [makefile](./makefile) is provided. To build the extension, package it into a layer and then publish that layer simply do:

```bash
make publish REGION=your-region
```

The command above will output the ARN of your layer.

You can then use the layer's ARN to add it to a function as follows:

```bash
make add REGION=your-region LAYER_ARN=your:extension:arn FUNCTION_NAME=your-function-name
```