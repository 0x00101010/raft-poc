package control

import (
	"bytes"
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

type INodeAdmin interface {
	StartSequencer(hsh common.Hash) error
	StopSequencer() (common.Hash, error)
}

type NodeAdmin struct {
	serverAddr string
	client     *http.Client
}

var _ INodeAdmin = (*NodeAdmin)(nil)

// StartSequencer implements INodeAdmin.
func (n *NodeAdmin) StartSequencer(hsh common.Hash) error {
	req := rpc.JsonRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  StartSequencerMethod,
		Params:  []any{hsh.String()},
		Id:      0,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal json request")
	}

	if _, err = n.client.Post(n.serverAddr, rpc.ContentTypeApplicationJSON, bytes.NewBuffer(data)); err != nil {
		return errors.Wrap(err, "failed to send request")
	}

	return nil
}

// StopSequencer implements INodeAdmin.
func (n *NodeAdmin) StopSequencer() (common.Hash, error) {
	req := rpc.JsonRPCRequest{
		Version: rpc.DefaultJsonRPCVersion,
		Method:  StopSequencerMethod,
		Params:  []any{},
		Id:      0,
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
