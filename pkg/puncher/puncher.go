package puncher

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	EasyEasyAcceptorMs    = 250
	EasyEasyReadTimeoutMs = 1000
)

const (
	EasyHardAcceptorMs     = 250
	EasyHardChunkTimeoutMs = 500
	EasyHardChunkProbes    = 128
	EasyHardTotalProbes    = 8192
)

const (
	HardEasyReadTimeoutMs = 30000
	HardEasyTotalProbes   = 256
)

var ErrNoExpected = fmt.Errorf("failed to receive expected")

func EasyEasy(conn *net.UDPConn, peerAddr *net.UDPAddr, wasFirst bool) error {
	if wasFirst {
		// opening port
		err := sendProbeMsg(conn, peerAddr, "EasyEasy: open")
		if err != nil {
			return err
		}
		fmt.Println("EasyEasy: opened port")

		// catching probe
		_, err = receiveProbe(conn, EasyEasyReadTimeoutMs) // todo: ignore probes from unexpected addresses
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return ErrNoExpected
			}
			return err
		}
		fmt.Println("EasyEasy: received probe")

		// sending confirmation
		err = sendProbeMsg(conn, peerAddr, "EasyEasy: confirm")
		if err != nil {
			return err
		}
		fmt.Println("EasyEasy: sent confirmation")
	} else {
		// waiting for peer to open port
		time.Sleep(EasyEasyAcceptorMs * time.Millisecond)

		t0 := time.Now()

		// probing peer
		err := sendProbeMsg(conn, peerAddr, "EasyEasy: probe")
		if err != nil {
			return err
		}
		fmt.Println("EasyEasy: sent probe")

		// catching confirmation
		_, err = receiveProbe(conn, EasyEasyReadTimeoutMs) // todo: ignore probes from unexpected addresses
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return ErrNoExpected
			}
			return err
		}
		fmt.Println("EasyEasy: received confirmation", "in", time.Since(t0))
	}

	return nil
}

func EasyHard(conn *net.UDPConn, peerIP net.IP) (*net.UDPAddr, error) {
	// waiting for peer to open ports
	time.Sleep(EasyHardAcceptorMs * time.Millisecond)

	used := make([]bool, math.MaxUint16)
	for j := 0; j*EasyHardChunkProbes < EasyHardTotalProbes; j++ {
		t0 := time.Now()
		for i := 0; i < EasyHardChunkProbes; i++ {
			var port int
			for {
				port = rand.Intn(math.MaxUint16)
				if !used[port] && port > 1024 {
					used[port] = true
					break
				}
			}
			// probing peer
			dst := &net.UDPAddr{IP: peerIP, Port: port}
			err := sendProbeMsg(conn, dst, fmt.Sprintf("EasyHard: probe %d", i))
			if err != nil {
				return nil, err
			}
		}
		fmt.Println("EasyHard: sent", EasyHardChunkProbes, "probes")

		// catching confirmation
		addr, err := receiveProbe(conn, EasyHardChunkTimeoutMs) // todo: ignore probes from unexpected addresses
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			return nil, err
		}
		fmt.Println("EasyHard: received confirmation from", addr, "in", time.Since(t0))

		return addr, nil
	}

	return nil, ErrNoExpected
}

type responseHardEasy struct {
	conn *net.UDPConn
	err  error
}

func HardEasy(peerAddr *net.UDPAddr) (*net.UDPConn, error) {
	respCh := make(chan responseHardEasy)
	for i := 0; i < HardEasyTotalProbes; i++ {
		go func() {
			conn, err := net.ListenUDP("udp4", nil)
			if err != nil {
				respCh <- responseHardEasy{err: err}
				return
			}

			// opening port
			err = sendProbeMsg(conn, peerAddr, fmt.Sprintf("HardEasy: open %d", i))
			if err != nil {
				respCh <- responseHardEasy{err: err}
				_ = conn.Close()
				return
			}

			// catching probe
			_, err = receiveProbe(conn, HardEasyReadTimeoutMs) // todo: ignore probes from unexpected addresses
			if err != nil {
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					respCh <- responseHardEasy{err: err}
				}
				_ = conn.Close()
				return
			}

			select {
			case respCh <- responseHardEasy{conn: conn}:
				fmt.Println("HardEasy: received probe")
			default:
				_ = conn.Close()
			}
		}()
	}

	fmt.Println("HardEasy: opened", HardEasyTotalProbes, "ports")

	select {
	case resp := <-respCh:
		if resp.err != nil {
			return nil, resp.err
		}
		// sending confirmation
		err := sendProbeMsg(resp.conn, peerAddr, "HardEasy: confirm")
		if err != nil {
			return nil, err
		}
		fmt.Println("EasyEasy: sent confirmation")
		return resp.conn, resp.err
	case <-time.After(HardEasyReadTimeoutMs * time.Millisecond):
		return nil, ErrNoExpected
	}
}

func sendProbeMsg(conn *net.UDPConn, peerAddr *net.UDPAddr, msg string) error {
	_, err := conn.WriteToUDP([]byte("probe "+msg), peerAddr)
	return err
}

func receiveProbe(conn *net.UDPConn, timeout int) (*net.UDPAddr, error) {
	buf := make([]byte, 1536)
	err := conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(timeout)))
	if err != nil {
		return nil, err
	}
	_, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		return nil, err
	}
	return addr, nil
}
