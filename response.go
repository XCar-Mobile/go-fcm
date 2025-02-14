package fcm

import (
	"encoding/json"
	"errors"
	"fmt"
)

// FCM HTTP v1 error variables.
var (
	ErrInvalidParameters   = errors.New("invalid parameters")
	ErrAuthentication      = errors.New("authentication error")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrNotFound            = errors.New("not found")
	ErrInternalServerError = errors.New("internal server error")
	ErrUnavailable         = errors.New("service unavailable")
	ErrUnknown             = errors.New("unknown error")
)

const (
	// No more information is available about this error.
	ErrorCodeUnspecifiedError = "UNSPECIFIED_ERROR"

	// HTTP error code = 400
	//
	// Request parameters were invalid. An extension of type
	// google.rpc.BadRequest is returned to specify which field was invalid.
	ErrorCodeInvalidArgument = "INVALID_ARGUMENT"

	// HTTP error code = 404
	//
	// App instance was unregistered from FCM. This usually means that the token
	// used is no longer valid and a new one must be used.
	ErrorCodeUnregistered = "UNREGISTERED"

	// HTTP error code = 403
	//
	// The authenticated sender ID is different from the sender ID
	// for the registration token.
	ErrorCodeSenderIdMismatch = "SENDER_ID_MISMATCH"

	// HTTP error code = 429
	//
	// Sending limit exceeded for the message target. An extension of type
	// google.rpc.QuotaFailure is returned to specify which quota was exceeded.
	ErrorCodeQuotaExceeded = "QUOTA_EXCEEDED"

	// HTTP error code = 503
	//
	// The server is overloaded.
	ErrorCodeUnavailable = "UNAVAILABLE"

	// HTTP error code = 500
	//
	// An unknown internal error occurred.
	ErrorCodeInternal = "INTERNAL"

	// HTTP error code = 401
	//
	// APNs certificate or web push auth key was invalid or missing.
	ErrorCodeThirdPartyAuthError = "THIRD_PARTY_AUTH_ERROR"
)

const errTypeFCMError = "type.googleapis.com/google.firebase.fcm.v1.FcmError"

// connectionError represents connection errors such as timeout error, etc.
// Implements `net.Error` interface.
type connectionError string

func (err connectionError) Error() string {
	return string(err)
}

func (err connectionError) Temporary() bool {
	return true
}

func (err connectionError) Timeout() bool {
	return true
}

// serverError represents internal server errors.
// Implements `net.Error` interface.
type serverError string

func (err serverError) Error() string {
	return string(err)
}

func (serverError) Temporary() bool {
	return true
}

func (serverError) Timeout() bool {
	return false
}

type (
	// Response represents the FCM HTTP v1 server response.
	// On success, the response contains the "name" field (a string like "projects/myproject/messages/123").
	// On failure, the response contains an "error" field following the google.rpc.Status format.
	Response struct {
		// Success response: the server returns the fully qualified message name.
		Name string `json:"name,omitempty"`
		// Error response: see google.rpc.Status for details.
		Error *ResponseError `json:"error,omitempty"`
	}

	// ResponseError represents the error structure returned by the FCM HTTP v1 API.
	ResponseError struct {
		Code    int                   `json:"code"`
		Message string                `json:"message"`
		Status  string                `json:"status"`
		Details []ResponseErrorDetail `json:"details,omitempty"`
	}

	// ResponseErrorDetail represents additional details that may be provided with an error.
	ResponseErrorDetail struct {
		// The fully qualified type of the error detail.
		Type string `json:"@type"`
		// An error code specific to the FCM API (if provided).
		ErrorCode string `json:"errorCode,omitempty"`
		// For some errors, details about which field(s) are invalid.
		FieldViolations []ResponseErrorFieldViolation `json:"fieldViolations,omitempty"`
	}

	// ResponseErrorFieldViolation provides details about an individual field error.
	ResponseErrorFieldViolation struct {
		Field       string `json:"field"`
		Description string `json:"description"`
	}
)

// UnmarshalJSON implements custom unmarshalling for Response.
// It simply decodes the JSON into the Response struct.
func (r *Response) UnmarshalJSON(data []byte) error {
	type respAlias Response
	var temp respAlias
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	*r = Response(temp)
	return nil
}

func (r *Response) Err() error {
	if r.Error == nil {
		return nil
	}

	errCode := ErrorCodeUnspecifiedError
	for _, detail := range r.Error.Details {
		errCode = detail.ErrorCode

		if detail.Type == errTypeFCMError {
			errCode = detail.ErrorCode
			break
		}
	}

	return fmt.Errorf("FCM error (%s | %s): %s",
		r.Error.Status, errCode, r.Error.Message)
}
