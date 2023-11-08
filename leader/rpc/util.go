package rpc

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

const (
	DefaultJsonRPCVersion      = "2.0"
	ContentTypeApplicationJSON = "application/json"
)

type JsonRPCRequest struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	Id      int    `json:"id"`
}

func Post(c *http.Client, url string, req JsonRPCRequest) (*http.Response, error) {
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
