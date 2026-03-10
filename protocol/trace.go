package protocol

import (
	"sync"

	"github.com/SCKelemen/lsp"
)

var traceValue TraceValue = TraceValueOff
var traceValueLock sync.Mutex

func GetTraceValue() TraceValue {
	traceValueLock.Lock()
	defer traceValueLock.Unlock()
	return traceValue
}

func SetTraceValue(value TraceValue) {
	traceValueLock.Lock()
	defer traceValueLock.Unlock()

	// The spec clearly says "message", but some implementations use "messages" instead
	if value == "messages" {
		value = TraceValueMessage
	}
	switch value {
	case TraceValueOff, TraceValueMessage, TraceValueVerbose:
	default:
		value = TraceValueOff
	}

	traceValue = value
}

func HasTraceLevel(value TraceValue) bool {
	value_ := GetTraceValue()
	switch value_ {
	case TraceValueOff:
		return false

	case TraceValueMessage:
		return value == TraceValueMessage

	case TraceValueVerbose:
		return true

	default:
		return false
	}
}

func HasTraceMessageType(type_ MessageType) bool {
	switch type_ {
	case MessageTypeError, MessageTypeWarning, MessageTypeInfo:
		return HasTraceLevel(TraceValueMessage)

	case MessageTypeLog:
		return HasTraceLevel(TraceValueVerbose)

	default:
		return false
	}
}

func Trace(context *lsp.Context, type_ MessageType, message string) error {
	if HasTraceMessageType(type_) {
		go context.Notify(ServerWindowLogMessage, &LogMessageParams{
			Type:    type_,
			Message: message,
		})
	}
	return nil
}
