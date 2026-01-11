package server

import (
	"time"

	"github.com/tliron/commonlog"
	"github.com/SCKelemen/lsp"
)

var DefaultTimeout = time.Minute

//
// Server
//

type Server struct {
	Handler     lsp.Handler
	LogBaseName string
	Debug       bool

	Log              commonlog.Logger
	Timeout          time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	StreamTimeout    time.Duration
	WebSocketTimeout time.Duration
}

func NewServer(handler lsp.Handler, logName string, debug bool) *Server {
	return &Server{
		Handler:          handler,
		LogBaseName:      logName,
		Debug:            debug,
		Log:              commonlog.GetLogger(logName),
		Timeout:          DefaultTimeout,
		ReadTimeout:      DefaultTimeout,
		WriteTimeout:     DefaultTimeout,
		StreamTimeout:    DefaultTimeout,
		WebSocketTimeout: DefaultTimeout,
	}
}
