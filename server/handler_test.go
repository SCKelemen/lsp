package server

import (
	contextpkg "context"
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/SCKelemen/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type stubHandler struct {
	result      any
	validMethod bool
	validParams bool
	err         error
	calls       int
	lastContext *lsp.Context
}

func (s *stubHandler) Handle(context *lsp.Context) (any, bool, bool, error) {
	s.calls++
	s.lastContext = context
	return s.result, s.validMethod, s.validParams, s.err
}

func TestHandleMethodNotFound(t *testing.T) {
	handler := &stubHandler{
		validMethod: false,
		validParams: false,
	}
	server := NewServer(handler, "server-test-method-not-found", false)

	result, err := server.handle(contextpkg.Background(), nil, &jsonrpc2.Request{
		Method: "custom/unsupported",
	})
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	var rpcErr *jsonrpc2.Error
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected jsonrpc2.Error, got %T", err)
	}
	if rpcErr.Code != jsonrpc2.CodeMethodNotFound {
		t.Fatalf("expected method-not-found code, got %d", rpcErr.Code)
	}
	if rpcErr.Message == "" {
		t.Fatal("expected non-empty method-not-found message")
	}
}

func TestHandleInvalidParamsWithoutErrorMessage(t *testing.T) {
	handler := &stubHandler{
		validMethod: true,
		validParams: false,
	}
	server := NewServer(handler, "server-test-invalid-params-empty", false)

	result, err := server.handle(contextpkg.Background(), nil, &jsonrpc2.Request{
		Method: "textDocument/hover",
	})
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	var rpcErr *jsonrpc2.Error
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected jsonrpc2.Error, got %T", err)
	}
	if rpcErr.Code != jsonrpc2.CodeInvalidParams {
		t.Fatalf("expected invalid-params code, got %d", rpcErr.Code)
	}
	if rpcErr.Message != "" {
		t.Fatalf("expected empty message, got %q", rpcErr.Message)
	}
}

func TestHandleInvalidParamsWithErrorMessage(t *testing.T) {
	handler := &stubHandler{
		validMethod: true,
		validParams: false,
		err:         errors.New("bad params"),
	}
	server := NewServer(handler, "server-test-invalid-params-message", false)

	result, err := server.handle(contextpkg.Background(), nil, &jsonrpc2.Request{
		Method: "textDocument/hover",
	})
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	var rpcErr *jsonrpc2.Error
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected jsonrpc2.Error, got %T", err)
	}
	if rpcErr.Code != jsonrpc2.CodeInvalidParams {
		t.Fatalf("expected invalid-params code, got %d", rpcErr.Code)
	}
	if rpcErr.Message != "bad params" {
		t.Fatalf("expected propagated message, got %q", rpcErr.Message)
	}
}

func TestHandleInvalidRequestFromHandlerError(t *testing.T) {
	handler := &stubHandler{
		validMethod: true,
		validParams: true,
		err:         errors.New("handler failed"),
	}
	server := NewServer(handler, "server-test-invalid-request", false)

	result, err := server.handle(contextpkg.Background(), nil, &jsonrpc2.Request{
		Method: "workspace/symbol",
	})
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	var rpcErr *jsonrpc2.Error
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected jsonrpc2.Error, got %T", err)
	}
	if rpcErr.Code != jsonrpc2.CodeInvalidRequest {
		t.Fatalf("expected invalid-request code, got %d", rpcErr.Code)
	}
	if rpcErr.Message != "handler failed" {
		t.Fatalf("expected propagated message, got %q", rpcErr.Message)
	}
}

func TestHandleSuccessAndContextForwarding(t *testing.T) {
	handler := &stubHandler{
		result:      map[string]string{"ok": "yes"},
		validMethod: true,
		validParams: true,
	}
	server := NewServer(handler, "server-test-success", false)

	params := json.RawMessage(`{"line":1}`)
	result, err := server.handle(contextpkg.Background(), nil, &jsonrpc2.Request{
		Method: "textDocument/definition",
		Params: &params,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if handler.calls != 1 {
		t.Fatalf("expected handler to be called once, got %d", handler.calls)
	}
	if handler.lastContext == nil {
		t.Fatal("expected forwarded context")
	}
	if handler.lastContext.Method != "textDocument/definition" {
		t.Fatalf("expected forwarded method, got %q", handler.lastContext.Method)
	}
	if string(handler.lastContext.Params) != string(params) {
		t.Fatalf("expected forwarded params %s, got %s", string(params), string(handler.lastContext.Params))
	}
	if handler.lastContext.Context == nil {
		t.Fatal("expected non-nil context.Context")
	}
}

func TestHandleExitClosesConnection(t *testing.T) {
	handler := &stubHandler{
		validMethod: true,
		validParams: true,
	}
	server := NewServer(handler, "server-test-exit", false)

	serverConn, clientConn := newJSONRPCConnPair()
	defer serverConn.Close()
	defer clientConn.Close()

	result, err := server.handle(contextpkg.Background(), serverConn, &jsonrpc2.Request{
		Method: "exit",
		Notif:  true,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	if handler.calls != 1 {
		t.Fatalf("expected exit handler call, got %d", handler.calls)
	}

	select {
	case <-serverConn.DisconnectNotify():
	case <-time.After(2 * time.Second):
		t.Fatal("server connection was not closed after exit")
	}
}

func newJSONRPCConnPair() (*jsonrpc2.Conn, *jsonrpc2.Conn) {
	serverSide, clientSide := net.Pipe()
	noop := jsonrpc2.HandlerWithError(func(contextpkg.Context, *jsonrpc2.Conn, *jsonrpc2.Request) (any, error) {
		return nil, nil
	})
	serverConn := jsonrpc2.NewConn(
		contextpkg.Background(),
		jsonrpc2.NewBufferedStream(serverSide, jsonrpc2.VSCodeObjectCodec{}),
		noop,
	)
	clientConn := jsonrpc2.NewConn(
		contextpkg.Background(),
		jsonrpc2.NewBufferedStream(clientSide, jsonrpc2.VSCodeObjectCodec{}),
		noop,
	)
	return serverConn, clientConn
}
