package nat

import (
	"net"
	"reflect"
)

type Detector struct {
	stunClient StunClient
	servers    [2]string
}

type StunClient interface {
	Endpoint(conn *net.UDPConn, addrStr string) (*net.UDPAddr, error)
}

func NewDetector(stunClient StunClient, servers [2]string) *Detector {
	return &Detector{stunClient: stunClient, servers: servers}
}

func (d *Detector) Detect(conn *net.UDPConn) (Nat, *net.UDPAddr, error) {
	endpoint1, err := d.stunClient.Endpoint(conn, d.servers[0])
	if err != nil {
		return 0, nil, err
	}

	endpoint2, err := d.stunClient.Endpoint(conn, d.servers[1])
	if err != nil {
		return 0, nil, err
	}

	if reflect.DeepEqual(endpoint1, endpoint2) {
		return EIM, endpoint1, nil
	}

	return EDM, &net.UDPAddr{IP: endpoint1.IP}, nil
}
