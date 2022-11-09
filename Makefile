.PHONY: test
test:
	go test ./... -race -coverprofile coverage.out -covermode atomic

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o bin/extensions/go-example-logs-api-extension
	chmod +x bin/extensions/go-example-logs-api-extension

.PHONY: package
package: build
	cd bin && zip -r ../extension.zip extensions/

.PHONY: publish
publish: package
	aws lambda publish-layer-version --layer-name "go-example-logs-api-extension" --region ${REGION} --zip-file  "fileb://extension.zip" | jq -r '.LayerVersionArn'
	
.PHONY: add
add: 
	aws lambda update-function-configuration --region ${REGION} --function-name ${FUNCTION_NAME} --layers ${LAYER_ARN}