package logging_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/chippyash/gamedistributor/src/logging"
	"github.com/rs/zerolog"
)

func TestAddOrgAddsOrgIdToEntries(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	withOrg := logging.AddOrg(logger, 42)
	withOrg.Info().Msg("hello")
	if !strings.Contains(buf.String(), `"orgId":42`) {
		t.Errorf("log entry %q does not contain orgId", buf.String())
	}
}

func TestPanicHandlerRecoversErrorPanic(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	ph := logging.NewPanicHandler(&logger)
	func() {
		defer ph.DoRecovery()
		panic(errors.New("boom"))
	}()
	out := buf.String()
	if !strings.Contains(out, "Unexpected error") || !strings.Contains(out, "boom") {
		t.Errorf("log output %q does not report the recovered error", out)
	}
}

func TestPanicHandlerRecoversNonErrorPanic(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	ph := logging.NewPanicHandler(&logger)
	func() {
		defer ph.DoRecovery()
		panic("plain string")
	}()
	out := buf.String()
	if !strings.Contains(out, "Unexpected exception") || !strings.Contains(out, "plain string") {
		t.Errorf("log output %q does not report the recovered value", out)
	}
}

func TestPanicHandlerNoPanicLogsNothing(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	ph := logging.NewPanicHandler(&logger)
	func() {
		defer ph.DoRecovery()
	}()
	if buf.Len() != 0 {
		t.Errorf("expected no log output, got %q", buf.String())
	}
}

func TestInitLoggersReturnUsableLoggers(t *testing.T) {
	t.Setenv("ENVIRONMENT", "unit-test")
	t.Cleanup(func() { zerolog.SetGlobalLevel(zerolog.DebugLevel) })
	extra := map[string]interface{}{"k": "v"}

	apigw := logging.InitLoggerFromAPIGWRequest(events.APIGatewayV2HTTPRequest{
		Headers:        map[string]string{"Host": "example"},
		RequestContext: events.APIGatewayV2HTTPRequestContext{RequestID: "req-1"},
	}, "chan", "ctx-id", "name", &extra)
	apigw.Info().Msg("apigw")

	nothing := logging.InitLoggerFromNothing("chan", "ctx-id", "name", nil)
	nothing.Info().Msg("nothing")

	s3 := logging.InitLoggerFromS3Event(events.S3EventRecord{}, "chan", "ctx-id", "name", nil)
	s3.Info().Msg("s3")
}
