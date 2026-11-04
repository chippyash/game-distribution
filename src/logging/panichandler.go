/*
*
Global panic handler

usage: At the top of your code (usually the first lines of the lambda Handler method):
ph := logging.NewPanicHandler(&log)
defer ph.DoRecovery()
*/
package logging

import (
	"github.com/rs/zerolog"
	"runtime/debug"
)

// our Panic Handler
type PanicHandler struct {
	Logger *zerolog.Logger
}

// do a deferred call to this to handle unexpected errors
func (p *PanicHandler) DoRecovery() {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			//recover returned some error, so log it
			p.Logger.WithLevel(zerolog.FatalLevel).
				Err(err).
				Stack().
				Msg("Unexpected error")
		} else {
			//recover returned something else
			p.Logger.WithLevel(zerolog.FatalLevel).
				Interface("Recovered exception object", r).
				Stack().
				Msg("Unexpected exception")
		}
		p.Logger.WithLevel(zerolog.FatalLevel).Bytes("stack-trace", debug.Stack()).Send()
	}
}

// create and return a panic handler
func NewPanicHandler(logger *zerolog.Logger) PanicHandler {
	return PanicHandler{Logger: logger}
}
