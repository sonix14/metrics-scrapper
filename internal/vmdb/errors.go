package vmdb

import "errors"

var (
	ErrSendingRequest               = errors.New("sending request")
	ErrUnexpectedResponseStatusCode = errors.New("unexpected http response status code")
	ErrReadingRequestBody           = errors.New("reading request body")
	ErrUnmarshalingRequestBody      = errors.New("unmarshaling request body")
	ErrFailedConvertExecTimestamp   = errors.New("converting last exec timestamp to string failed")
	ErrParsingVMURL                 = errors.New("parsing vm url")
)
