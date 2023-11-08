package health

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const ContentTypeApplicationJSON = "application/json"

type JsonRPCRequest struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	Id      int    `json:"id"`
}

type HealthzResponse struct {
	Version string `json:"version"`
}

type Client struct {
	serverAddr string
	client     *http.Client
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
		client:     &http.Client{},
	}
}

func (c *Client) Healthy() (bool, error) {
	hq := JsonRPCRequest{
		Version: "2.0",
		Method:  "",
		Params:  []any{},
		Id:      0,
	}

	data, err := json.Marshal(hq)
	if err != nil {
		return false, err
	}

	url := fmt.Sprintf("%s/healthz", c.serverAddr)
	_, err = c.client.Post(url, ContentTypeApplicationJSON, bytes.NewBuffer(data))
	if err != nil {
		return false, err
	}

	return true, nil
}
