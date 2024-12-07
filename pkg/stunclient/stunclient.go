package stunclient

import (
	"log"
	"net"

	"github.com/pion/stun/v3"
)

type StunClient struct {
}

func New() *StunClient {
	return &StunClient{}
}

func (c *StunClient) Endpoint(conn *net.UDPConn, addrStr string) (*net.UDPAddr, error) {
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return nil, err
	}

	request := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	_, err = conn.WriteTo(request.Raw, addr)

	buf := make([]byte, 1536)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]

	m := new(stun.Message)
	m.Raw = buf
	err = m.Decode()
	if err != nil {
		log.Fatal(err)
	}

	mappedAddr := &stun.MappedAddress{}
	err = mappedAddr.GetFrom(m)
	if err != nil {
		return nil, err
	}

	externalAddr := &net.UDPAddr{
		IP:   mappedAddr.IP,
		Port: mappedAddr.Port,
	}

	return externalAddr, nil
}
