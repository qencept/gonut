package rendezvousclient

import "github.com/qencept/gonut/pkg/nat"

type Request struct {
	Id       string  `json:"id"`
	Nat      nat.Nat `json:"nat"`
	Endpoint string  `json:"endpoint"`
}

type Response struct {
	Nat      nat.Nat `json:"nat"`
	Endpoint string  `json:"endpoint"`
	WasFirst bool    `json:"was_first"`
}
