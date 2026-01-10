package server

import (
	contextpkg "context"
	"io"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	wsjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/tliron/commonlog"
)

func (self *Server) newStreamConnection(stream io.ReadWriteCloser) *jsonrpc2.Conn {
	handler := self.newHandler()
	connectionOptions := self.newConnectionOptions()

	// Use background context for connection lifetime - LSP connections should persist
	// for the duration of the editor session, not be limited by a timeout
	context := contextpkg.Background()

	return jsonrpc2.NewConn(context, jsonrpc2.NewBufferedStream(stream, jsonrpc2.VSCodeObjectCodec{}), handler, connectionOptions...)
}

func (self *Server) newWebSocketConnection(socket *websocket.Conn) *jsonrpc2.Conn {
	handler := self.newHandler()
	connectionOptions := self.newConnectionOptions()

	// Use background context for connection lifetime - LSP connections should persist
	// for the duration of the editor session, not be limited by a timeout
	context := contextpkg.Background()

	return jsonrpc2.NewConn(context, wsjsonrpc2.NewObjectStream(socket), handler, connectionOptions...)
}

func (self *Server) newConnectionOptions() []jsonrpc2.ConnOpt {
	if self.Debug {
		log := commonlog.NewScopeLogger(self.Log, "rpc")
		return []jsonrpc2.ConnOpt{jsonrpc2.LogMessages(&JSONRPCLogger{log})}
	} else {
		return nil
	}
}
