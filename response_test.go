package fcm

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestUnmarshalSuccessResponse(t *testing.T) {
	data := []byte(`{
		"name": "projects/myproject/messages/12345"
	}`)

	var response Response
	err := json.Unmarshal(data, &response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := Response{
		Name:  "projects/myproject/messages/12345",
		Error: nil,
	}
	if !reflect.DeepEqual(response, expected) {
		t.Fatalf("expected: %+v, got: %+v", expected, response)
	}

	if response.Err() != nil {
		t.Fatalf("expected no error, got: %v", response.Err())
	}
}

func TestUnmarshalErrorResponse(t *testing.T) {
	data := []byte(`{
		"error": {
			"code": 400,
			"message": "Invalid argument: registration token missing",
			"status": "INVALID_ARGUMENT",
			"details": [{
				"@type": "type.googleapis.com/google.firebase.fcm.v1.FcmError",
				"errorCode": "UNREGISTERED"
			}]
		}
	}`)

	var response Response
	err := json.Unmarshal(data, &response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ensure the success field is empty.
	if response.Name != "" {
		t.Fatalf("expected empty name in error response, got: %s", response.Name)
	}

	if response.Error == nil {
		t.Fatal("expected error in response, got nil")
	}

	if response.Error.Code != 400 {
		t.Errorf("expected error code 400, got: %d", response.Error.Code)
	}

	if response.Error.Message != "Invalid argument: registration token missing" {
		t.Errorf("unexpected error message: %s", response.Error.Message)
	}

	if response.Error.Status != "INVALID_ARGUMENT" {
		t.Errorf("unexpected error status: %s", response.Error.Status)
	}

	// Check mapping from status to our custom error.
	respErr := response.Err()
	if !strings.Contains(respErr.Error(), ErrorCodeUnregistered) {
		t.Errorf("expected mapped error to contain %v, got %v", ErrorCodeUnregistered, respErr)
	}
}

func TestUnmarshalUnknownErrorResponse(t *testing.T) {
	// Test error response with a status not mapped to a predefined error.
	data := []byte(`{
		"error": {
			"code": 500,
			"message": "Some unexpected error occurred",
			"status": "SOME_UNKNOWN_STATUS"
		}
	}`)

	var response Response
	err := json.Unmarshal(data, &response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.Error == nil {
		t.Fatal("expected error in response, got nil")
	}

	errMapped := response.Err()
	expectedErrMsg := "FCM error (SOME_UNKNOWN_STATUS): UNSPECIFIED_ERROR, Some unexpected error occurred"
	if errMapped.Error() != expectedErrMsg {
		t.Errorf("expected mapped error to be %q, got %q", expectedErrMsg, errMapped.Error())
	}
}

func TestUnmarshalMalformedResponse(t *testing.T) {
	// Malformed JSON: "name" is not a string.
	data := []byte(`{
		"name": ["projects/myproject/messages/12345"]
	}`)

	var response Response
	err := json.Unmarshal(data, &response)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}
