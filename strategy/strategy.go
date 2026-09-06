package strategy

import (
	"fmt"
	"sync/atomic"
)

type StrategyType string

const (
	RoundRobin         StrategyType = "round-robin"
	WeightedRoundRobin StrategyType = "weighted-round-robin"
	LeastConnections   StrategyType = "least-connections"
)

func (s StrategyType) StrategyType() string {
	return string(s)
}

type Strategy interface {
	NextServer() uint32
	AddBackendCount(length uint32)
}

type RoundRobinAlgo struct {
	Current          atomic.Uint32
	LengthofReplicas atomic.Uint32
}

func NewRoundRobin(length uint32) *RoundRobinAlgo {
	rr := &RoundRobinAlgo{
		Current:          atomic.Uint32{},
		LengthofReplicas: atomic.Uint32{},
	}
	rr.LengthofReplicas.Store(length)
	return rr
}

func (rr *RoundRobinAlgo) NextServer() uint32 {
	length := uint32(rr.LengthofReplicas.Load())
	for {
		current := rr.Current.Load()
		next := current + 1

		if next >= length {
			next = 0
		}
		if rr.Current.CompareAndSwap(current, next) {
			return next
		}
	}
}

func (rr *RoundRobinAlgo) AddBackendCount(length uint32) {
	newLength := rr.LengthofReplicas.Load() + length
	rr.LengthofReplicas.Store(newLength)
}

type WeightedRoundRobinAlgo struct {
	Current          atomic.Uint32
	LengthofReplicas atomic.Uint32
	Weights          []uint32
	CurrentWeight    atomic.Uint32
	MaxWeight        uint32
	GCD              uint32
}

func NewWeightedRoundRobin(length uint32, weights []uint32) *WeightedRoundRobinAlgo { return nil }
func (wrr *WeightedRoundRobinAlgo) NextServer() uint32                              { return 0 }
func (wrr *WeightedRoundRobinAlgo) AddBackendCount(length uint32)                {}

type StrategyConfig struct {
	Type    StrategyType
	Weights []uint32
}

func NewStrategy(cfg StrategyConfig, length uint32) (Strategy, error) {

	switch StrategyType(cfg.Type) {
	case RoundRobin:
		return NewRoundRobin(length), nil

	case WeightedRoundRobin:
		return NewWeightedRoundRobin(length, cfg.Weights), nil

	// case LeastConnections:
	// return NewLeastConnections(), nil

	default:
		return nil, fmt.Errorf("unsupported strategy: %s", cfg.Type)
	}
}
