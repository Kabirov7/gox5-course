package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"lesson8/httpserver/internal/domain"
	"lesson8/httpserver/internal/users"
)

var uuidV7Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func testHandler() http.Handler {
	return NewHandler(users.NewService(), slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func request(t *testing.T, h http.Handler, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	h.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("%s %s: status = %d, want %d; body = %s", method, path, response.Code, status, response.Body)
	}
	if !uuidV7Pattern.MatchString(response.Header().Get("X-Request-ID")) {
		t.Fatalf("invalid request ID: %q", response.Header().Get("X-Request-ID"))
	}
	return response
}

func decodeResponse[T any](t *testing.T, response *httptest.ResponseRecorder) T {
	t.Helper()
	if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected JSON content type, got %q", response.Header().Get("Content-Type"))
	}
	var result T
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v; body = %s", err, response.Body)
	}
	return result
}

func TestUserLifecycle(t *testing.T) {
	h := testHandler()
	health := request(t, h, http.MethodGet, "/healthz", "", http.StatusOK)
	if health.Body.String() != "ok" {
		t.Fatalf("health response = %q", health.Body.String())
	}
	empty := request(t, h, http.MethodGet, "/users", "", http.StatusOK)
	if strings.TrimSpace(empty.Body.String()) != "[]" {
		t.Fatalf("empty list = %s", empty.Body)
	}

	created := request(t, h, http.MethodPost, "/users", `{"name":" Alice ","age":30}`, http.StatusOK)
	result := decodeResponse[map[string]string](t, created)
	id := result["id"]
	if len(result) != 1 || !uuidV7Pattern.MatchString(id) {
		t.Fatalf("create response = %v", result)
	}
	want := domain.User{ID: id, Name: "Alice", Age: 30}
	got := decodeResponse[domain.User](t, request(t, h, http.MethodGet, "/users/"+id, "", http.StatusOK))
	if got != want {
		t.Fatalf("user = %+v, want %+v", got, want)
	}
	list := decodeResponse[[]domain.User](t, request(t, h, http.MethodGet, "/users", "", http.StatusOK))
	if len(list) != 1 || list[0] != want {
		t.Fatalf("list = %+v, want [%+v]", list, want)
	}

	want.Name, want.Age = "Bob", 0
	updated := request(t, h, http.MethodPut, "/users/"+id, `{"name":"Bob","age":0}`, http.StatusOK)
	if got := decodeResponse[domain.User](t, updated); got != want {
		t.Fatalf("updated user = %+v, want %+v", got, want)
	}
	if got := decodeResponse[domain.User](t, request(t, h, http.MethodGet, "/users/"+id, "", http.StatusOK)); got != want {
		t.Fatalf("persisted user = %+v, want %+v", got, want)
	}
	deleted := request(t, h, http.MethodDelete, "/users/"+id, "", http.StatusNoContent)
	if deleted.Body.Len() != 0 {
		t.Fatalf("delete response should be empty: %s", deleted.Body)
	}
	request(t, h, http.MethodGet, "/users/"+id, "", http.StatusNotFound)
	list = decodeResponse[[]domain.User](t, request(t, h, http.MethodGet, "/users", "", http.StatusOK))
	if len(list) != 0 {
		t.Fatalf("list after deletion = %+v", list)
	}
}

func TestInvalidUserInput(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"malformed JSON", `{"name":`},
		{"missing fields", `{}`},
		{"missing name", `{"age":30}`},
		{"missing age", `{"name":"Alice"}`},
		{"blank name", `{"name":" \t ","age":30}`},
		{"null name", `{"name":null,"age":30}`},
		{"null age", `{"name":"Alice","age":null}`},
		{"negative age", `{"name":"Alice","age":-1}`},
		{"fractional age", `{"name":"Alice","age":1.5}`},
		{"string age", `{"name":"Alice","age":"30"}`},
		{"number name", `{"name":12,"age":30}`},
		{"age overflow", `{"name":"Alice","age":999999999999999999999999}`},
		{"unknown field", `{"name":"Alice","age":30,"admin":true}`},
		{"null object", `null`},
		{"array", `[]`},
		{"multiple objects", `{"name":"Alice","age":30} {}`},
		{"trailing garbage", `{"name":"Alice","age":30} !`},
		{"oversized body", `{"name":"` + strings.Repeat("a", 1<<20) + `","age":30}`},
	}
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					h := testHandler()
					created := request(t, h, http.MethodPost, "/users", `{"name":"Alice","age":30}`, http.StatusOK)
					id := decodeResponse[map[string]string](t, created)["id"]
					path := "/users"
					if method == http.MethodPut {
						path += "/" + id
					}
					response := request(t, h, method, path, tc.body, http.StatusBadRequest)
					if result := decodeResponse[map[string]string](t, response); result["error"] == "" {
						t.Fatalf("missing error description: %v", result)
					}
					list := decodeResponse[[]domain.User](t, request(t, h, http.MethodGet, "/users", "", http.StatusOK))
					if len(list) != 1 || list[0] != (domain.User{ID: id, Name: "Alice", Age: 30}) {
						t.Fatalf("invalid request changed stored users: %+v", list)
					}
				})
			}
		})
	}
}

func TestMissingUserAndRouting(t *testing.T) {
	h := testHandler()
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			response := request(t, h, method, "/users/missing", `{"name":"Alice","age":30}`, http.StatusNotFound)
			if got := decodeResponse[map[string]string](t, response)["error"]; got != "user not found" {
				t.Fatalf("error = %q", got)
			}
		})
	}
	request(t, h, http.MethodGet, "/unknown", "", http.StatusNotFound)
	request(t, h, http.MethodPatch, "/users/missing", "", http.StatusMethodNotAllowed)
	request(t, h, http.MethodPost, "/healthz", "", http.StatusMethodNotAllowed)
}

type serviceStub struct {
	UserService
	get func(context.Context) ([]domain.User, error)
}

func (s serviceStub) Get(ctx context.Context) ([]domain.User, error) {
	return s.get(ctx)
}

func TestServiceFailureAndPanic(t *testing.T) {
	for _, panicRequest := range []bool{false, true} {
		t.Run(fmt.Sprintf("panic=%t", panicRequest), func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			service := serviceStub{get: func(ctx context.Context) ([]domain.User, error) {
				Logger(ctx).InfoContext(ctx, "service called")
				if panicRequest {
					panic("private service details")
				}
				return nil, errors.New("private service details")
			}}
			h := NewHandler(service, logger)
			response := request(t, h, http.MethodGet, "/users", "", http.StatusInternalServerError)
			if got := decodeResponse[map[string]string](t, response)["error"]; got != "internal server error" {
				t.Fatalf("internal error exposed: %q", got)
			}
			requestID := response.Header().Get("X-Request-ID")
			decoder := json.NewDecoder(&logs)
			var entries []map[string]any
			for decoder.More() {
				var entry map[string]any
				if err := decoder.Decode(&entry); err != nil {
					t.Fatal(err)
				}
				if entry["request_id"] != requestID {
					t.Fatalf("request ID missing from log: %v", entry)
				}
				entries = append(entries, entry)
			}
			if len(entries) != 3 {
				t.Fatalf("expected service, error and access logs, got %v", entries)
			}
			access := entries[2]
			if access["method"] != "GET" || access["path"] != "/users" || access["status"] != float64(500) {
				t.Fatalf("invalid access log: %v", access)
			}
			if duration, ok := access["duration_ms"].(float64); !ok || duration < 0 {
				t.Fatalf("invalid duration: %v", access)
			}
			if panicRequest && entries[1]["stack"] == "" {
				t.Fatal("panic stack trace missing")
			}
			health := request(t, h, http.MethodGet, "/healthz", "", http.StatusOK)
			if health.Header().Get("X-Request-ID") == requestID {
				t.Fatal("request ID reused")
			}
		})
	}
}

func TestRequestLogging(t *testing.T) {
	var logs bytes.Buffer
	h := NewHandler(users.NewService(), slog.New(slog.NewJSONHandler(&logs, nil)))
	cases := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{http.MethodGet, "/healthz", "", http.StatusOK},
		{http.MethodPost, "/users", `{}`, http.StatusBadRequest},
		{http.MethodGet, "/unknown", "", http.StatusNotFound},
		{http.MethodPatch, "/users", "", http.StatusMethodNotAllowed},
	}
	for _, tc := range cases {
		logs.Reset()
		response := request(t, h, tc.method, tc.path, tc.body, tc.status)
		var entry map[string]any
		if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
			t.Fatal(err)
		}
		if entry["request_id"] != response.Header().Get("X-Request-ID") ||
			entry["method"] != tc.method || entry["path"] != tc.path || entry["status"] != float64(tc.status) {
			t.Fatalf("invalid access log: %v", entry)
		}
		if duration, ok := entry["duration_ms"].(float64); !ok || duration < 0 {
			t.Fatalf("invalid duration: %v", entry)
		}
	}
}

func TestConcurrentRequests(t *testing.T) {
	h := testHandler()
	t.Run("clients", func(t *testing.T) {
		for i := 0; i < 32; i++ {
			t.Run(fmt.Sprintf("client-%d", i), func(t *testing.T) {
				t.Parallel()
				created := request(t, h, http.MethodPost, "/users", `{"name":"Alice","age":30}`, http.StatusOK)
				id := decodeResponse[map[string]string](t, created)["id"]
				request(t, h, http.MethodGet, "/users/"+id, "", http.StatusOK)
				request(t, h, http.MethodGet, "/users", "", http.StatusOK)
				request(t, h, http.MethodPut, "/users/"+id, `{"name":"Bob","age":31}`, http.StatusOK)
				request(t, h, http.MethodDelete, "/users/"+id, "", http.StatusNoContent)
			})
		}
	})
	list := decodeResponse[[]domain.User](t, request(t, h, http.MethodGet, "/users", "", http.StatusOK))
	if len(list) != 0 {
		t.Fatalf("remaining users: %+v", list)
	}
}
