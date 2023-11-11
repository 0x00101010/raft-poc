package health

import (
	"fmt"
	"net/http"

	"github.com/base-org/leader-election/leader/rpc"
)

const ContentTypeApplicationJSON = "application/json"

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
	req := rpc.JSONRPCRequest{
		Version: "2.0",
		Method:  "",
		Params:  []any{},
		ID:      0,
	}

	resp, err := rpc.Post(c.client, fmt.Sprintf("%s/healthz", c.serverAddr), req)
	if err != nil {
		return false, err
	}

	fmt.Printf("%s response code is %d\n", c.serverAddr, resp.StatusCode)

	return true, nil
}
