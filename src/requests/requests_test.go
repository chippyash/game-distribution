package requests_test

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/chippyash/gamedistributor/src/requests"
)

func TestModifyRequestSetsContextIdAndKeepsHeaders(t *testing.T) {
	in := events.APIGatewayProxyRequest{
		Headers:         map[string]string{"Accept": "application/json"},
		IsBase64Encoded: true,
	}
	out := requests.ModifyRequest(in, "ctx-9")
	if got := out.Headers["x-context-id"]; got != "ctx-9" {
		t.Errorf("x-context-id: got %q, want %q", got, "ctx-9")
	}
	if got := out.Headers["Accept"]; got != "application/json" {
		t.Errorf("Accept header: got %q, want %q", got, "application/json")
	}
	if out.IsBase64Encoded {
		t.Errorf("IsBase64Encoded: got true, want false")
	}
}
