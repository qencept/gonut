package rendezvousclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/qencept/gonut/pkg/nat"
)

type RendezvousClient struct {
	server string
}

func New(server string) *RendezvousClient {
	return &RendezvousClient{server: server}
}

func (c *RendezvousClient) PeerNatAddr(id string, nat nat.Nat, addr *net.UDPAddr) (nat.Nat, *net.UDPAddr, bool, error) {
	buf := &bytes.Buffer{}
	req := Request{
		Id:       id,
		Nat:      nat,
		Endpoint: fmt.Sprint(addr),
	}

	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return 0, nil, false, err
	}

	response, err := http.Post(c.server, "application/json", buf)
	if err != nil {
		return 0, nil, false, err
	}

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return 0, nil, false, err
		}
		return 0, nil, false, fmt.Errorf("rendezvousclient: (%d) %v", response.StatusCode, strings.TrimSuffix(string(body), "\n"))
	}

	var resp Response
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return 0, nil, false, err
	}

	peerAddr, err := net.ResolveUDPAddr("udp4", resp.Endpoint)
	if err != nil {
		return 0, nil, false, err
	}

	return resp.Nat, peerAddr, resp.WasFirst, nil
}
