package logging

/**
 * To create a log entry with equivalent of monolog CRITICAL or EMERGENCY without
 * exiting the program do:
 * log.WithLevel(zerolog.FatalLevel).Msg("...") //CRITICAL
 * log.WithLevel(zerolog.PanicLevel).Msg("...") //EMERGENCY
 *
 * There is no current way to do NOTICE or ALERT. @see https://github.com/rs/zerolog/issues/220
 */

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rs/zerolog"
	"os"
)

// ZERO_LOG_TIME_FORMAT See MAR-62
const ZERO_LOG_TIME_FORMAT = "2006-01-02T15:04:05.999Z"

type LogContext struct {
	User string `json:"user"`
}

// add monolog equivalent levels to log entry. NB no NOTICE or ALERT equivalents
type MonologHook struct{}

func (h MonologHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level != zerolog.NoLevel {
		switch level {
		case zerolog.DebugLevel:
			e.Int("monolog_level", 100) //Monolog DEBUG
		case zerolog.InfoLevel:
			e.Int("monolog_level", 200) //Monolog INFO
		case zerolog.WarnLevel:
			e.Int("monolog_level", 300) //Monolog WARNING
		case zerolog.ErrorLevel:
			e.Int("monolog_level", 400) //Monolog ERROR
		//The following will exit the program
		case zerolog.FatalLevel:
			e.Int("monolog_level", 500) //Monolog CRITICAL
		case zerolog.PanicLevel:
			e.Int("monolog_level", 600) //Monolog EMERGENCY
		}
	}
}

// InitLoggerFromAPIGWRequest initialises the logger from an API Gateway request
func InitLoggerFromAPIGWRequest(request events.APIGatewayV2HTTPRequest, channelName string, contextId string, contextName string, extra *map[string]interface{}) zerolog.Logger {
	initZeroLog()

	host := request.Headers["Host"]
	if request.Headers["X-Forwarded-For"] != "" {
		host = request.Headers["X-Forwarded-For"]
	}

	extraContext := make(map[string]interface{})
	extraContext["ContextId"] = contextId
	extraContext["ContextName"] = contextName
	extraContext["RequestId"] = request.RequestContext.RequestID
	extraContext["Host"] = host
	if extra != nil {
		for key, val := range *extra {
			extraContext[key] = val
		}
	}

	return zerolog.New(os.Stdout).With().
		Timestamp().
		Interface("extra", extraContext).
		Str("channel", channelName).
		Str("environment", os.Getenv("ENVIRONMENT")).
		Caller().
		Logger().
		Hook(MonologHook{})
}

func InitLoggerFromS3Event(event events.S3EventRecord, channelName string, contextId string, contextName string, extra *map[string]interface{}) zerolog.Logger {
	initZeroLog()
	extraContext := make(map[string]interface{})
	extraContext["ContextId"] = contextId
	extraContext["ContextName"] = contextName
	extraContext["RequestId"] = event.ResponseElements["z-amz-request-id"]
	if extra != nil {
		for key, val := range *extra {
			extraContext[key] = val
		}
	}
	return zerolog.New(os.Stdout).With().
		Timestamp().
		Interface("extra", extraContext).
		Interface("context", LogContext{User: event.PrincipalID.PrincipalID}).
		Str("channel", channelName).
		Str("environment", os.Getenv("ENVIRONMENT")).
		Caller().
		Logger().
		Hook(MonologHook{})
}

// init logger from nothing - aka not known environment
func InitLoggerFromNothing(channelName string, contextId string, contextName string, extra *map[string]interface{}) zerolog.Logger {
	initZeroLog()
	extraContext := make(map[string]interface{})
	extraContext["ContextId"] = contextId
	extraContext["ContextName"] = contextName
	if extra != nil {
		for key, val := range *extra {
			extraContext[key] = val
		}
	}
	return zerolog.New(os.Stdout).With().
		Timestamp().
		Interface("extra", extraContext).
		Str("channel", channelName).
		Str("environment", os.Getenv("ENVIRONMENT")).
		Caller().
		Logger().
		Hook(MonologHook{})
}

func InitLoggerFromLambdaContext(ctx context.Context, channelName string, contextName string, extra *map[string]interface{}) zerolog.Logger {
	lctx, _ := lambdacontext.FromContext(ctx)
	initZeroLog()
	extraContext := make(map[string]interface{})
	extraContext["ContextId"] = lctx.AwsRequestID
	extraContext["ContextName"] = contextName
	if extra != nil {
		for key, val := range *extra {
			extraContext[key] = val
		}
	}
	return zerolog.New(os.Stdout).With().
		Timestamp().
		Interface("extra", extraContext).
		Interface("context", LogContext{User: lctx.Identity.CognitoIdentityID}).
		Str("channel", channelName).
		Str("environment", os.Getenv("ENVIRONMENT")).
		Caller().
		Logger().
		Hook(MonologHook{})

}

// Basic logger initialisation for our lambda environments
func initZeroLog() {
	zerolog.TimeFieldFormat = ZERO_LOG_TIME_FORMAT
	zerolog.TimestampFieldName = "@timestamp"
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	env := os.Getenv("ENVIRONMENT")
	if env == "sandbox" || env == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	if env == "unit-test" {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}
}

// AddOrg  add org id to all subsequent log entries
func AddOrg(logger zerolog.Logger, orgId int64) zerolog.Logger {
	return logger.With().Int64("orgId", orgId).Logger()
}
