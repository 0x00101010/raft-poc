package control

import (
	"net/http"

	"github.com/base-org/leader-election/leader/rpc"
	"github.com/pkg/errors"
)

const (
	StartBatcherMethod = "admin_startBatcher"
	StopBatcherMethod  = "admin_stopBatcher"
)

type BatcherRPC interface {
	StartBatcher() error
	StopBatcher() error
}

type BatcherRPCClient struct {
	serverAddr string
	client     *http.Client
}

var _ BatcherRPC = (*BatcherRPCClient)(nil)

func NewBatcherRPC(serverAddr string) BatcherRPC {
	return &BatcherRPCClient{
		serverAddr: serverAddr,
		client:     &http.Client{},
	}
}

// StartBatcher implements IBatcherAdmin.
func (b *BatcherRPCClient) StartBatcher() error {
	req := rpc.JSONRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  StartBatcherMethod,
		Params:  []any{},
		ID:      0,
	}

	if _, err := rpc.Post(b.client, b.serverAddr, req); err != nil {
		return errors.Wrap(err, "failed to send request")
	}

	return nil
}

// StopBatcher implements IBatcherAdmin.
func (b *BatcherRPCClient) StopBatcher() error {
	req := rpc.JSONRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  StopBatcherMethod,
		Params:  []any{},
		ID:      0,
	}

	if _, err := rpc.Post(b.client, b.serverAddr, req); err != nil {
		return errors.Wrap(err, "failed to send request")
	}

	return nil
}
