package control

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/base-org/leader-election/leader/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type GethRPC interface {
	LatestBlock() (common.Hash, error)
}

type GethRPCClient struct {
	serverAddr string
	client     *http.Client
}

var _ GethRPC = (*GethRPCClient)(nil)

func NewGethRPC(serverAddr string) GethRPC {
	fmt.Printf("NewGethRPC: %s\n", serverAddr)
	return &GethRPCClient{
		serverAddr: serverAddr,
		client:     &http.Client{},
	}
}

// LatestBlock implements NodeRPC.
func (g *GethRPCClient) LatestBlock() (common.Hash, error) {
	req := rpc.JSONRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  "eth_getBlockByNumber",
		Params:  []any{"latest", true},
		ID:      0,
	}

	resp, err := rpc.Post(g.client, g.serverAddr, req)
	if err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to send request")
	}

	bytes, err := io.ReadAll(resp.Body)
	fmt.Println(string(bytes))
	if err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to read response body")
	}

	var result rpc.JSONRPCResponse
	if err := json.Unmarshal(bytes, &result); err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to unmarshal response body")
	}
	fmt.Println(result)

	blockResult, ok := result.Result.(string)
	if !ok {
		return common.Hash{}, errors.New("failed to convert result to bytes")
	}

	var block rpc.Block
	if err := json.Unmarshal([]byte(blockResult), &block); err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to unmarshal response body")
	}

	return common.HexToHash(block.Hash), nil
}

type MockGethRPC struct{}

var _ GethRPC = (*MockGethRPC)(nil)

func NewMockGethRPC() GethRPC {
	return &MockGethRPC{}
}

// LatestBlock implements NodeRPC.
func (*MockGethRPC) LatestBlock() (common.Hash, error) {
	return common.Hash{}, nil
}
