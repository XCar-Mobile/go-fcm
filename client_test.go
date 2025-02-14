package fcm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	t.Run("send=success", func(t *testing.T) {
		expectedName := "projects/test/messages/12345"
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected: Bearer token, got: %s", req.Header.Get("Authorization"))
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, `{"name": "%s"}`, expectedName)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithTimeout(10*time.Second))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := client.Send(&NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Name != expectedName {
			t.Fatalf("expected name: %s, got: %s", expectedName, resp.Name)
		}
	})

	t.Run("send=failure", func(t *testing.T) {
		// Simulate an error response with HTTP 400.
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected: Bearer token, got: %s", req.Header.Get("Authorization"))
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, `{
				"error": {
					"code": 400,
					"message": "Invalid argument: topic missing",
					"status": "INVALID_ARGUMENT"
				}
			}`)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := client.Send(&NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token")
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if resp != nil {
			t.Fatalf("expected nil response, got: %v", resp)
		}
	})

	t.Run("send=invalid_token", func(t *testing.T) {
		_, err := NewClient("test", WithEndpoint(""))
		if err == nil {
			t.Fatal("expected error due to empty endpoint, got nil")
		}
	})

	t.Run("send=invalid_message", func(t *testing.T) {
		c, err := NewClient("test", WithEndpoint("http://example.com"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Assuming NewMessage.Validate() returns an error for an invalid message.
		_, err = c.Send(&NewMessage{Message: Message{}}, "token")
		if err == nil {
			t.Fatal("expected error for invalid message, got nil")
		}
	})

	t.Run("send=invalid-response", func(t *testing.T) {
		// Simulate a malformed JSON response.
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			// "name" is expected to be a string but here it is a number.
			fmt.Fprint(rw, `{"name": 12345}`)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithHTTPClient(&http.Client{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = client.SendWithRetry(&NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token", 3)
		if err == nil {
			t.Fatal("expected error due to invalid response JSON, got nil")
		}
	})
}

func TestSendWithRetry(t *testing.T) {
	t.Run("send_with_retry=success", func(t *testing.T) {
		expectedName := "projects/test/messages/12345"
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected: Bearer token, got: %s", req.Header.Get("Authorization"))
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, `{"name": "%s"}`, expectedName)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithHTTPClient(&http.Client{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := client.SendWithRetry(&NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token", 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Name != expectedName {
			t.Fatalf("expected name: %s, got: %s", expectedName, resp.Name)
		}
	})

	t.Run("send_with_retry=failure", func(t *testing.T) {
		// Simulate an error response.
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Header().Set("Content-Type", "application/json")
			fmt.Fprint(rw, `{
				"error": {
					"code": 400,
					"message": "Bad Request",
					"status": "INVALID_ARGUMENT"
				}
			}`)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithHTTPClient(&http.Client{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := client.SendWithRetry(&NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token", 2)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if resp != nil {
			t.Fatalf("expected nil response, got: %v", resp)
		}
	})

	t.Run("send_with_retry=success_retry", func(t *testing.T) {
		var attempts int
		expectedName := "projects/test/messages/12345"
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			attempts++
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected: Bearer token, got: %s", req.Header.Get("Authorization"))
			}
			rw.Header().Set("Content-Type", "application/json")
			if attempts < 3 {
				rw.WriteHeader(http.StatusInternalServerError)
				// Return an error response for temporary errors.
				fmt.Fprint(rw, `{"error": {"code": 500, "message": "Internal Server Error", "status": "INTERNAL"}}`)
			} else {
				rw.WriteHeader(http.StatusOK)
				fmt.Fprintf(rw, `{"name": "%s"}`, expectedName)
			}
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithHTTPClient(&http.Client{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := client.SendWithRetry(&NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token", 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attempts != 3 {
			t.Fatalf("expected 3 attempts, got: %d attempts", attempts)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Name != expectedName {
			t.Fatalf("expected name: %s, got: %s", expectedName, resp.Name)
		}
	})

	t.Run("send_with_retry=failure_retry", func(t *testing.T) {
		// Use a client with a very short timeout to force a connection error.
		client, err := NewClient("test",
			WithEndpoint("http://127.0.0.1:80"),
			WithHTTPClient(&http.Client{Timeout: time.Nanosecond}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := client.SendWithRetry(&NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token", 3)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if resp != nil {
			t.Fatalf("expected nil response, got: %v", resp)
		}
	})
}

func TestSendWithRetryWithContext(t *testing.T) {
	t.Run("send_with_retry_with_context=success", func(t *testing.T) {
		expectedName := "projects/test/messages/12345"
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected: Bearer token, got: %s", req.Header.Get("Authorization"))
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, `{"name": "%s"}`, expectedName)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithHTTPClient(&http.Client{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ctx := context.Background()
		resp, err := client.SendWithRetryWithContext(ctx, &NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token", 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Name != expectedName {
			t.Fatalf("expected name: %s, got: %s", expectedName, resp.Name)
		}
	})

	t.Run("send_with_retry_with_context=failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected: Bearer token, got: %s", req.Header.Get("Authorization"))
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, `{
				"error": {
					"code": 400,
					"message": "Bad Request",
					"status": "INVALID_ARGUMENT"
				}
			}`)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithHTTPClient(&http.Client{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ctx := context.Background()
		resp, err := client.SendWithRetryWithContext(ctx, &NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token", 2)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if resp != nil {
			t.Fatalf("expected nil response, got: %v", resp)
		}
	})

	t.Run("send_with_retry_with_context=success_retry", func(t *testing.T) {
		var attempts int
		expectedName := "projects/test/messages/12345"
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			attempts++
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected: Bearer token, got: %s", req.Header.Get("Authorization"))
			}
			rw.Header().Set("Content-Type", "application/json")
			if attempts < 3 {
				rw.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(rw, `{"error": {"code": 500, "message": "Internal Error", "status": "INTERNAL"}}`)
			} else {
				rw.WriteHeader(http.StatusOK)
				fmt.Fprintf(rw, `{"name": "%s"}`, expectedName)
			}
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithHTTPClient(&http.Client{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ctx := context.Background()
		resp, err := client.SendWithRetryWithContext(ctx, &NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token", 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attempts != 3 {
			t.Fatalf("expected 3 attempts, got: %d attempts", attempts)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Name != expectedName {
			t.Fatalf("expected name: %s, got: %s", expectedName, resp.Name)
		}
	})

	t.Run("send_with_retry_with_context=failure_retry", func(t *testing.T) {
		client, err := NewClient("test",
			WithEndpoint("http://127.0.0.1:80"),
			WithHTTPClient(&http.Client{Timeout: time.Nanosecond}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ctx := context.Background()
		resp, err := client.SendWithRetryWithContext(ctx, &NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token", 3)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if resp != nil {
			t.Fatalf("expected nil response, got: %v", resp)
		}
	})

	t.Run("send_with_retry_with_context=failure_timeout", func(t *testing.T) {
		var attempts int
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			attempts++
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected: Bearer token, got: %s", req.Header.Get("Authorization"))
			}
			if attempts < 3 {
				rw.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(rw, `{"error": {"code": 500, "message": "Internal Error", "status": "INTERNAL"}}`)
			} else {
				rw.WriteHeader(http.StatusOK)
				fmt.Fprint(rw, `{"name": "projects/test/messages/12345"}`)
			}
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithTimeout(10*time.Second))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_, err = client.SendWithRetryWithContext(ctx, &NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}}, "token", 4)
		if err == nil {
			t.Fatalf("expected context timeout error")
		}
		if attempts != 1 {
			t.Fatalf("expected 1 attempt due to context timeout, got: %d attempts", attempts)
		}
		_, ok := err.(connectionError)
		if !ok {
			t.Fatalf("error is not of type connectionError, got: %T", err)
		}
	})
}

func TestSendWithContext(t *testing.T) {
	t.Run("send_context=success", func(t *testing.T) {
		expectedName := "projects/test/messages/12345"
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected: Bearer token, got: %s", req.Header.Get("Authorization"))
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, `{"name": "%s"}`, expectedName)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithTimeout(10*time.Second))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ctx := context.Background()
		resp, err := client.SendWithContext(ctx, "token", &NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Name != expectedName {
			t.Fatalf("expected name: %s, got: %s", expectedName, resp.Name)
		}
	})

	t.Run("send_context=timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("expected: Bearer token, got: %s", req.Header.Get("Authorization"))
			}
			time.Sleep(100 * time.Millisecond)
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			fmt.Fprint(rw, `{"name": "projects/test/messages/12345"}`)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithTimeout(10*time.Second))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_, err = client.SendWithContext(ctx, "token", &NewMessage{Message{
			Topic: "test",
			Data:  map[string]interface{}{"foo": "bar"},
		}})
		if err == nil {
			t.Fatalf("expected context timeout error")
		}
		_, ok := err.(connectionError)
		if !ok {
			t.Fatalf("error is not of type connectionError, got: %T", err)
		}
	})
}
