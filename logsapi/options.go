package logsapi

import "log"

type Options struct {
	ExtensionID         string           // The ID of the extension
	RecordsBufferSize   int              // The size of the records channel buffer
	LogServerAddress    string           // The address that the log server should assume
	AwsLambdaRuntimeAPI string           // The URL of the Lambda Runtime API
	ErrCallback         *func(err error) // A callback used for any errs raised when handling requests from the Lambda Logs API
}

var defaultErrCallback = func(err error) {
	log.Println(err.Error())
}

func (o *Options) setDefaults() {
	if o.ErrCallback == nil {
		o.ErrCallback = &defaultErrCallback
	}
}
