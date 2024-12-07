package peerlinker

import (
	"fmt"
	"net"
	"strconv"

	"github.com/qencept/gonut/pkg/nat"
	"github.com/qencept/gonut/pkg/puncher"
	"github.com/qencept/gonut/pkg/rendezvousclient"
	"github.com/qencept/gonut/pkg/stunclient"
)

type Linker struct {
	detector         *nat.Detector
	rendezvousClient *rendezvousclient.RendezvousClient
}

func New(stunServers [2]string, rendezvousServer string) *Linker {
	return &Linker{
		detector:         nat.NewDetector(stunclient.New(), stunServers),
		rendezvousClient: rendezvousclient.New(rendezvousServer),
	}
}

func (l *Linker) Link(id string) (*net.UDPConn, *net.UDPAddr, error) {
	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, nil, err
	}

	myNat, myAddr, err := l.detector.Detect(conn)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("Me:", myNat, omitZeroPort(myAddr))

	peerNat, peerAddr, wasFirst, err := l.rendezvousClient.PeerNatAddr(id, myNat, myAddr)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("Peer:", peerNat, omitZeroPort(peerAddr))

	if myNat == nat.EIM && peerNat == nat.EIM {
		err = puncher.EasyEasy(conn, peerAddr, wasFirst)
		if err != nil {
			return nil, nil, err
		}
	} else if myNat == nat.EIM && peerNat == nat.EDM {
		peerAddr, err = puncher.EasyHard(conn, peerAddr.IP) // new addr (peers NAT port found)
		if err != nil {
			return nil, nil, err
		}
	} else if myNat == nat.EDM && peerNat == nat.EIM {
		_ = conn.Close()
		conn, err = puncher.HardEasy(peerAddr) // new conn (my NAT port in use)
		if err != nil {
			return nil, nil, err
		}
	} else {
		return nil, nil, fmt.Errorf("can not handle")
	}

	return conn, peerAddr, nil
}

func omitZeroPort(addr *net.UDPAddr) string {
	str := addr.IP.String()
	if addr.Port != 0 {
		str += ":" + strconv.Itoa(addr.Port)
	}
	return str
}
