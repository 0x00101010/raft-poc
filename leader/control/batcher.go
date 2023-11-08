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

type BatcherAdmin interface {
	StartBatcher() error
	StopBatcher() error
}

type BatcherAdminRPC struct {
	serverAddr string
	client     *http.Client
}

var _ BatcherAdmin = (*BatcherAdminRPC)(nil)

func NewBatcherAdmin(serverAddr string) BatcherAdmin {
	return &BatcherAdminRPC{
		serverAddr: serverAddr,
		client:     &http.Client{},
	}
}

// StartBatcher implements IBatcherAdmin.
func (b *BatcherAdminRPC) StartBatcher() error {
	req := rpc.JsonRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  StartBatcherMethod,
		Params:  []any{},
		Id:      0,
	}

	if _, err := rpc.Post(b.client, b.serverAddr, req); err != nil {
		return errors.Wrap(err, "failed to send request")
	}

	return nil
}

// StopBatcher implements IBatcherAdmin.
func (b *BatcherAdminRPC) StopBatcher() error {
	req := rpc.JsonRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  StopBatcherMethod,
		Params:  []any{},
		Id:      0,
	}

	if _, err := rpc.Post(b.client, b.serverAddr, req); err != nil {
		return errors.Wrap(err, "failed to send request")
	}

	return nil
}
