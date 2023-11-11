package control

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/base-org/leader-election/leader/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
)

const (
	StartSequencerMethod = "admin_startSequencer"
	StopSequencerMethod  = "admin_stopSequencer"
)

type NodeRPC interface {
	StartSequencer(hsh common.Hash) error
	StopSequencer() (common.Hash, error)
	SequencerActive() (bool, error)
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
	fmt.Printf("Starting sequencer at %s", hsh.String())

	req := rpc.JSONRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  StartSequencerMethod,
		Params:  []any{hsh.String()},
		ID:      0,
	}

	resp, err := rpc.Post(n.client, n.serverAddr, req)
	if err != nil {
		return errors.Wrap(err, "failed to send request")
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	var r rpc.JSONRPCResponse
	if err := json.Unmarshal(bytes, &r); err != nil {
		return errors.Wrap(err, "failed to unmarshal response body")
	}

	fmt.Println(r.Result)
	if r.Error != nil {
		fmt.Println(r.Error.Error())
	}

	fmt.Println("Sequencer started...")

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
	var r rpc.JSONRPCResponse
	if err := json.Unmarshal(bytes, &r); err != nil {
		return common.Hash{}, errors.Wrap(err, "failed to unmarshal response body")
	}

	hsh, ok := r.Result.(string)
	if !ok {
		fmt.Println("failed to convert result to hash string")
	}
	fmt.Printf("Sequencer stopped at %s\n", hsh)

	return common.HexToHash(hsh), nil
}

// SequencerActive implements NodeRPC.
func (n *NodeRPCClient) SequencerActive() (bool, error) {
	req := rpc.JSONRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  "admin_sequencerActive",
		Params:  []any{},
		ID:      0,
	}

	resp, err := rpc.Post(n.client, n.serverAddr, req)
	if err != nil {
		return false, errors.Wrap(err, "failed to send request")
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, errors.Wrap(err, "failed to read response body")
	}

	var r rpc.JSONRPCResponse
	if err := json.Unmarshal(bytes, &r); err != nil {
		fmt.Println(string(bytes))
		return false, errors.Wrap(err, "failed to unmarshal response body")
	}

	active, ok := r.Result.(bool)
	if !ok {
		return false, errors.New("failed to convert result to bool")
	}

	return active, nil
}

type MockNodeRPC struct{}

var _ NodeRPC = (*MockNodeRPC)(nil)

func NewMockNodeRPC() NodeRPC {
	return &MockNodeRPC{}
}

// StartSequencer implements NodeRPC.
func (*MockNodeRPC) StartSequencer(hsh common.Hash) error {
	log.Info("MockNodeRPC: StartSequencer")
	return nil
}

// StopSequencer implements NodeRPC.
func (*MockNodeRPC) StopSequencer() (common.Hash, error) {
	log.Info("MockNodeRPC: StopSequencer")
	return common.Hash{}, nil
}

// SequencerActive implements NodeRPC.
func (*MockNodeRPC) SequencerActive() (bool, error) {
	log.Info("MockNodeRPC: SequencerActive")
	return true, nil
}
