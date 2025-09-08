package models

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

func newPair(symbol, network string) NetworkPair {
	return NetworkPair{
		Symbol:  strings.ToLower(symbol),
		Network: strings.ToLower(network),
	}
}

type NetworkPair struct {
	Symbol  string `json:"symbol"`
	Network string `json:"network"`
}

func (n NetworkPair) String() string {
	return fmt.Sprintf("%s-%s", n.Symbol, n.Network)
}

func newNetworks() Networks {
	return Networks{
		lookup: map[NetworkPair]bool{},
	}
}

type Networks struct {
	lookup map[NetworkPair]bool
	first  NetworkPair
}

func (n *Networks) ensureInitialized() {
	if n.lookup == nil {
		n.lookup = make(map[NetworkPair]bool)
	}
}

func (n *Networks) Add(symbol, network string) *Networks {
	n.ensureInitialized()
	if len(n.lookup) == 0 {
		n.first = newPair(symbol, network)
	}
	n.lookup[newPair(symbol, network)] = true
	return n
}

func (n *Networks) Has(symbol, network string) bool {
	n.ensureInitialized()
	_, ok := n.lookup[newPair(symbol, network)]
	return ok
}

func (n *Networks) GetAll() []NetworkPair {
	return lo.Keys(n.lookup)
}
