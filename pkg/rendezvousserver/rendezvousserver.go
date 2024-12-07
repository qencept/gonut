package rendezvousserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/qencept/gonut/pkg/nat"
	"github.com/qencept/gonut/pkg/rendezvousclient"
)

const requestTimeoutSeconds = 60

type RendezvousServer struct {
	addr      string
	cert      string
	key       string
	awaitings map[string]*awaitingEntry
}

type awaitingEntry struct {
	ch1 chan *peerInfo
	ch2 chan *peerInfo
	ep1 string
}

type peerInfo struct {
	Nat      nat.Nat
	Endpoint string
}

func New(addr, cert, key string) *RendezvousServer {
	return &RendezvousServer{
		addr:      addr,
		cert:      cert,
		key:       key,
		awaitings: make(map[string]*awaitingEntry),
	}
}

func (s *RendezvousServer) Serve() error {
	if s.cert == "" && s.key == "" {
		log.Println("Listening HTTP on", s.addr)
		return http.ListenAndServe(s.addr, s)
	}
	log.Println("Listening HTTPS on", s.addr)
	return http.ListenAndServeTLS(s.addr, s.cert, s.key, s)
}

func (s *RendezvousServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req rendezvousclient.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[%v] Request: %+v\n", r.RemoteAddr, req)

	var chr <-chan *peerInfo
	var chw chan<- *peerInfo
	awaiting, ok := s.awaitings[req.Id]
	if !ok {
		awaiting = &awaitingEntry{
			ch1: make(chan *peerInfo),
			ch2: make(chan *peerInfo),
		}
		s.awaitings[req.Id] = awaiting
	}
	if awaiting.ep1 == "" {
		awaiting.ep1 = req.Endpoint
		chr, chw = awaiting.ch1, awaiting.ch2
	} else if awaiting.ep1 != req.Endpoint {
		chr, chw = awaiting.ch2, awaiting.ch1
	} else {
		var str string
		if awaiting.ep1 == req.Endpoint {
			str = fmt.Sprintf("already have Id:%v Endpoint:%v", req.Id, req.Endpoint)
		}
		http.Error(w, str, http.StatusBadRequest)
		log.Printf("[%v] Error: %+v\n", r.RemoteAddr, str)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		select {
		case chw <- &peerInfo{Nat: req.Nat, Endpoint: req.Endpoint}:
		case <-time.After(time.Second * requestTimeoutSeconds):
		}
		close(chw)
		wg.Done()
	}()
	var info *peerInfo
	go func() {
		select {
		case info = <-chr:
		case <-time.After(time.Second * requestTimeoutSeconds):
		}
		wg.Done()
	}()
	wg.Wait()

	if !ok {
		delete(s.awaitings, req.Id)
	}

	if info == nil {
		str := "no peer within timeout"
		http.Error(w, str, http.StatusRequestTimeout)
		log.Printf("[%v] Error: %+v\n", r.RemoteAddr, str)
		return
	}

	resp := rendezvousclient.Response{
		Nat:      info.Nat,
		Endpoint: info.Endpoint,
		WasFirst: !ok,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[%v] Response: %+v\n", r.RemoteAddr, resp)
}
