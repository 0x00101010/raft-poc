package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const (
	DefaultJsonRPCVersion      = "2.0"
	ContentTypeApplicationJSON = "application/json"
)

type JSONRPCRequest struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      int    `json:"id"`
}

type JSONRPCResponse struct {
	Version string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func (j JSONRPCError) Error() string {
	return fmt.Sprintf(
		`JSONRPCError code=%d message="%s"`,
		j.Code,
		j.Message,
	)
}

// Block represents the Ethereum block JSON structure returned by eth_getBlockByX.
type Block struct {
	// Number           string   `json:"number"`
	Hash string `json:"hash"`
	// ParentHash       string   `json:"parentHash"`
	// Nonce            string   `json:"nonce"`
	// Sha3Uncles       string   `json:"sha3Uncles"`
	// LogsBloom        string   `json:"logsBloom"`
	// TransactionsRoot string   `json:"transactionsRoot"`
	// StateRoot        string   `json:"stateRoot"`
	// ReceiptsRoot     string   `json:"receiptsRoot"`
	// Miner            string   `json:"miner"`
	// Difficulty       string   `json:"difficulty"`
	// TotalDifficulty  string   `json:"totalDifficulty"`
	// ExtraData        string   `json:"extraData"`
	// Size             string   `json:"size"`
	// GasLimit         string   `json:"gasLimit"`
	// GasUsed          string   `json:"gasUsed"`
	// Timestamp        string   `json:"timestamp"`
	// Transactions     []any    `json:"transactions"`
	// Uncles           []string `json:"uncles"`
}

func Post(c *http.Client, url string, req JSONRPCRequest) (*http.Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal json request")
	}

	resp, err := c.Post(url, ContentTypeApplicationJSON, bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}

	return resp, nil
}
