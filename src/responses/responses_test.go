package responses_test

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"testing"

	"github.com/chippyash/gamedistributor/src/responses"
	"github.com/rs/zerolog"
)

func newResponder() *responses.ApiGwResponder {
	return responses.NewApiGwResponder(zerolog.Nop(), "ctx-1")
}

func TestRespondReturnsJSONWithStandardHeaders(t *testing.T) {
	resp, err := newResponder().Respond(200, responses.BodyResponse{"msg": "hi"}, responses.Headers{}, false, false)
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode: got %d, want 200", resp.StatusCode)
	}
	if resp.Body != `{"msg":"hi"}` {
		t.Errorf("Body: got %q", resp.Body)
	}
	if resp.IsBase64Encoded {
		t.Errorf("IsBase64Encoded: got true, want false")
	}
	want := map[string]string{
		"Content-Type":                  "application/json",
		"Access-Control-Allow-Origin":   "*",
		"x-context-id":                  "ctx-1",
		"Access-Control-Expose-Headers": "x-context-id",
	}
	for k, v := range want {
		if got := resp.Headers[k]; got != v {
			t.Errorf("header %q: got %q, want %q", k, got, v)
		}
	}
}

func TestRespondCallerHeadersAreAdded(t *testing.T) {
	resp, err := newResponder().Respond(404, responses.BodyResponse{"msg": "nope"}, responses.Headers{"X-Custom": "yes"}, false, false)
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if got := resp.Headers["X-Custom"]; got != "yes" {
		t.Errorf("X-Custom: got %q, want %q", got, "yes")
	}
	if resp.StatusCode != 404 {
		t.Errorf("StatusCode: got %d, want 404", resp.StatusCode)
	}
}

func TestRespondBase64Encodes(t *testing.T) {
	resp, err := newResponder().Respond(200, responses.BodyResponse{"msg": "hi"}, responses.Headers{}, true, false)
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if !resp.IsBase64Encoded {
		t.Fatalf("IsBase64Encoded: got false, want true")
	}
	decoded, err := base64.StdEncoding.DecodeString(resp.Body)
	if err != nil {
		t.Fatalf("body is not base64: %v", err)
	}
	if string(decoded) != `{"msg":"hi"}` {
		t.Errorf("decoded body: got %q", decoded)
	}
}

func TestRespondGzipRoundTrip(t *testing.T) {
	t.Setenv("DISABLE_GZIP", "false")
	resp, err := newResponder().Respond(200, responses.BodyResponse{"msg": "hi"}, responses.Headers{}, false, true)
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if !resp.IsBase64Encoded {
		t.Fatalf("IsBase64Encoded: got false, want true when gzipped")
	}
	if got := resp.Headers["Content-Encoding"]; got != "gzip" {
		t.Errorf("Content-Encoding: got %q, want gzip", got)
	}
	compressed, err := base64.StdEncoding.DecodeString(resp.Body)
	if err != nil {
		t.Fatalf("body is not base64: %v", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("body is not gzip: %v", err)
	}
	plain, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("gunzip: %v", err)
	}
	if string(plain) != `{"msg":"hi"}` {
		t.Errorf("gunzipped body: got %q", plain)
	}
}

func TestRespondGzipDisabledByEnv(t *testing.T) {
	t.Setenv("DISABLE_GZIP", "true")
	resp, err := newResponder().Respond(200, responses.BodyResponse{"msg": "hi"}, responses.Headers{}, false, true)
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if resp.IsBase64Encoded {
		t.Errorf("IsBase64Encoded: got true, want false when DISABLE_GZIP=true")
	}
	if _, ok := resp.Headers["Content-Encoding"]; ok {
		t.Errorf("Content-Encoding set when DISABLE_GZIP=true")
	}
	if resp.Body != `{"msg":"hi"}` {
		t.Errorf("Body: got %q", resp.Body)
	}
}

func TestRespondRawSerialisesAnyValue(t *testing.T) {
	type payload struct {
		Count int `json:"count"`
	}
	resp, err := newResponder().RespondRaw(201, payload{Count: 3}, responses.Headers{}, false, false)
	if err != nil {
		t.Fatalf("RespondRaw: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("StatusCode: got %d, want 201", resp.StatusCode)
	}
	if resp.Body != `{"count":3}` {
		t.Errorf("Body: got %q", resp.Body)
	}
}
