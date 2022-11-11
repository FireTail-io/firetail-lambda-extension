ARCH := amd64
VERSION := latest
LAYER_NAME := firetail-extension-${ARCH}-${VERSION}
REGION := eu-west-1
AWS_amd64 := x86_64
AWS_arm64 := arm64
AWS_ARCH := $(AWS_$(ARCH))

.PHONY: test
test:
	go test ./... -race -coverprofile coverage.out -covermode atomic

.PHONY: build
build:
	rm -rf build
	GOOS=linux GOARCH=${ARCH} go build -o build/extensions/firetail-extension-${ARCH}
	chmod +x build/extensions/firetail-extension-${ARCH}

.PHONY: package
package: build
	cd build && zip -r ../build/firetail-extension-${ARCH}-${VERSION}.zip extensions/

.PHONY: publish
publish:
	@aws lambda publish-layer-version --layer-name "${LAYER_NAME}" --compatible-architectures "${AWS_ARCH}" --region "${REGION}" --zip-file  "fileb://build/firetail-extension-${ARCH}-${VERSION}.zip" | jq -r '.LayerVersionArn'
	
.PHONY: add
add: 
	aws lambda update-function-configuration --region ${REGION} --function-name ${FUNCTION_NAME} --layers ${LAYER_ARN}