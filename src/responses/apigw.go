package responses

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
	"os"
	"strconv"
)

// Our basic struct for handling api gateway responses
type ApiGwResponder struct {
	Logger    zerolog.Logger
	ContextId string
}

// proxy type for the body of a response
type BodyResponse map[string]interface{}

// proxy type for a set of headers
type Headers map[string]string

// constructor for ApiGwResponder
func NewApiGwResponder(logger zerolog.Logger, contextId string) *ApiGwResponder {
	return &ApiGwResponder{
		Logger:    logger,
		ContextId: contextId,
	}
}

// Respond Create a response using a BodyResponse to be sent back to API Gateway
func (a *ApiGwResponder) Respond(status int, body BodyResponse, headers Headers, isbase64Encoded bool, gzipResponse bool) (events.APIGatewayProxyResponse, error) {
	return a.RespondRaw(status, body, headers, isbase64Encoded, gzipResponse)
}

// RespondRaw json serialize whatever body you are give
func (a *ApiGwResponder) RespondRaw(status int, body interface{}, headers Headers, isbase64Encoded bool, gzipResponse bool) (events.APIGatewayProxyResponse, error) {
	var buf bytes.Buffer
	bodyJson, err := json.Marshal(body)
	if err != nil {
		a.Logger.WithLevel(zerolog.FatalLevel).
			Err(err).
			Msg("Json error")

		return events.APIGatewayProxyResponse{StatusCode: 404}, err
	}

	if gzipResponse {
		flag, err := strconv.ParseBool(os.Getenv("DISABLE_GZIP"))
		if err != nil {
			a.Logger.Error().Err(err).Msg("failed to parse DISABLE_GZIP env value. Setting default (false)")
			flag = false
		}
		if flag {
			gzipResponse = false
		}
	}

	var zipFailed = false
	if gzipResponse {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		if _, err := gz.Write([]byte(bodyJson)); err != nil {
			a.Logger.Error().Interface("gzip write failed", err).Send()
			zipFailed = true
		}
		if err := gz.Close(); err != nil {
			a.Logger.Error().Interface("gzip close failed", err).Send()
			zipFailed = true
		}
		if !zipFailed {
			bodyJson = b.Bytes()
			headers["Content-Encoding"] = "gzip"
			isbase64Encoded = true
		}
	}

	if isbase64Encoded {
		bodyJson = []byte(base64.StdEncoding.EncodeToString(bodyJson))
	}
	json.HTMLEscape(&buf, bodyJson)

	stdHeaders := Headers{
		"Content-Type":                  "application/json",
		"Access-Control-Allow-Origin":   "*",
		"x-context-id":                  a.ContextId,
		"Access-Control-Expose-Headers": "x-context-id",
	}
	for key, value := range headers {
		stdHeaders[key] = value
	}

	resp := events.APIGatewayProxyResponse{
		StatusCode:      status,
		IsBase64Encoded: isbase64Encoded,
		Body:            buf.String(),
		Headers:         stdHeaders,
	}

	loggableBody := bodyJson
	if gzipResponse && !zipFailed {
		// escape binary characters for sane logging
		loggableBody = []byte(strconv.QuoteToASCII(string(bodyJson)))
	}

	a.Logger.Debug().
		Int("status", status).
		RawJSON("body", loggableBody).
		Msg("Response to gateway")

	return resp, nil
}
