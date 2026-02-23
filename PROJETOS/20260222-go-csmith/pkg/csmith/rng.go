package csmith

const (
	lcgA    uint64 = 0x5DEECE66D
	lcgC    uint64 = 0xB
	lcgMask uint64 = (1 << 48) - 1
)

// rng is compatible with the libc srand48/lrand48 core recurrence used by Csmith.
type rng struct {
	state uint64
}

func newRNG(seed uint64) *rng {
	// srand48 semantics.
	return &rng{state: ((seed << 16) + 0x330E) & lcgMask}
}

func (r *rng) next31() uint32 {
	r.state = (lcgA*r.state + lcgC) & lcgMask
	return uint32(r.state >> 17)
}

func (r *rng) upto(n uint32) uint32 {
	if n == 0 {
		return 0
	}
	return r.next31() % n
}
