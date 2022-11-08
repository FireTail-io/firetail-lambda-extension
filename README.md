# Firetail Logging Extension

This extension is a proof of concept using the Lambda Logs API to extract request & responses and forward them to the Firetail SaaS.



## Deployment

You will need to install the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html), [Golang](https://go.dev/doc/install) and [JQ](https://stedolan.github.io/jq/).

A [makefile](./makefile) is provided. To build the extension, package it into a layer and then publish that layer simply do:

```bash
make publish REGION=your-region
```

This will output the ARN of your layer which you can then use to add it to a function:

```bash
make add REGION=your-region LAYER_ARN=your:extension:arn FUNCTION_NAME=your-function-name
```