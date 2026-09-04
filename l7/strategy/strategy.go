package strategy

import "sync/atomic"

type Strategy interface {
	NextServer() uint32
}

type RoundRobin struct {
	Current          atomic.Uint32
	LengthofReplicas atomic.Uint32
}

func NewRoundRobin(length uint32) *RoundRobin {
	rr := &RoundRobin{
		Current:          atomic.Uint32{},
		LengthofReplicas: atomic.Uint32{},
	}
	rr.LengthofReplicas.Store(length)
	return rr
}

func (rr *RoundRobin) NextServer() uint32 {
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

type WeightedRoundRobin struct {
	Current          atomic.Uint32
	LengthofReplicas atomic.Uint32
	Weights          []uint32
	CurrentWeight    atomic.Uint32
	MaxWeight        uint32
	GCD              uint32
}

func (wrr *WeightedRoundRobin) NextServer() uint32 { return 0 }
