package control

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/base-org/leader-election/leader/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	StartSequencerMethod = "admin_startSequencer"
	StopSequencerMethod  = "admin_stopSequencer"
)

type NodeRPC interface {
	StartSequencer(hsh common.Hash) error
	StopSequencer() (common.Hash, error)
	LatestBlock() (common.Hash, error)
}

type NodeRPCClient struct {
	serverAddr string
	client     *http.Client
}

var _ NodeRPC = (*NodeRPCClient)(nil)

func NewNodeRPC(serverAddr string) NodeRPC {
	return &NodeRPCClient{
		serverAddr: serverAddr,
		client:     &http.Client{},
	}
}

// StartSequencer implements INodeAdmin.
func (n *NodeRPCClient) StartSequencer(hsh common.Hash) error {
	req := rpc.JSONRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  StartSequencerMethod,
		Params:  []any{hsh.String()},
		ID:      0,
	}

	if _, err := rpc.Post(n.client, n.serverAddr, req); err != nil {
		return errors.Wrap(err, "failed to send request")
	}

	return nil
}

// StopSequencer implements INodeAdmin.
func (n *NodeRPCClient) StopSequencer() (common.Hash, error) {
	req := rpc.JSONRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  StopSequencerMethod,
		Params:  []any{},
		ID:      0,
	}

	resp, err := rpc.Post(n.client, n.serverAddr, req)
	if err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to read response body")
	}

	// TODO: place holder, need to change to correct unmarshalling code.
	var result common.Hash
	if err := json.Unmarshal(bytes, &result); err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to unmarshal response body")
	}

	return result, nil
}

// LatestBlock implements NodeRPC.
func (n *NodeRPCClient) LatestBlock() (common.Hash, error) {
	req := rpc.JSONRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  "eth_getBlockByNumber",
		Params:  []any{"latest", true},
		ID:      0,
	}

	resp, err := rpc.Post(n.client, n.serverAddr, req)
	if err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to send request")
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to read response body")
	}

	var result rpc.JSONRPCResponse
	if err := json.Unmarshal(bytes, &result); err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to unmarshal response body")
	}

	blockData, ok := result.Result.([]byte)
	if !ok {
		return common.Hash{}, errors.New("failed to convert result to bytes")
	}

	var block rpc.Block
	if err := json.Unmarshal(blockData, &block); err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to unmarshal response body")
	}

	return common.HexToHash(block.Hash), nil
}
