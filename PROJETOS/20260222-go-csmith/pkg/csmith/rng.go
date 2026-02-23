package csmith

import (
	"fmt"
	"os"
)

const (
	lcgA    uint64 = 0x5DEECE66D
	lcgC    uint64 = 0xB
	lcgMask uint64 = (1 << 48) - 1
)

// rng is compatible with the libc srand48/lrand48 core recurrence used by Csmith.
type rng struct {
	state     uint64
	trace     bool
	traceFile string
	tracePos  uint64
}

func newRNG(seed uint64) *rng {
	// srand48 semantics.
	r := &rng{state: ((seed << 16) + 0x330E) & lcgMask}
	if os.Getenv("CSMITH_TRACE_RNG") != "" {
		r.trace = true
		r.traceFile = os.Getenv("CSMITH_TRACE_RNG_FILE")
		if r.traceFile == "" {
			r.traceFile = "/tmp/csmith-go-rng.trace"
		}
		_ = os.WriteFile(r.traceFile, []byte(fmt.Sprintf("# seed=%d\n", seed)), 0644)
	}
	return r
}

func (r *rng) next31() uint32 {
	r.state = (lcgA*r.state + lcgC) & lcgMask
	return uint32(r.state >> 17)
}

func (r *rng) upto(n uint32) uint32 {
	if n == 0 {
		return 0
	}
	x := r.next31() % n
	r.traceU(n, x)
	return x
}

func (r *rng) uptoWithFilter(n uint32, reject func(uint32) bool) uint32 {
	if n == 0 {
		return 0
	}
	x := r.next31() % n
	for reject != nil && reject(x) {
		x = r.next31() % n
	}
	r.traceU(n, x)
	return x
}

func (r *rng) traceU(n uint32, x uint32) {
	if r.trace {
		r.tracePos++
		f, err := os.OpenFile(r.traceFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err == nil {
			_, _ = fmt.Fprintf(f, "%d U %d -> %d\n", r.tracePos, n, x)
			_ = f.Close()
		}
	}
}

func (r *rng) flipcoin(p uint32) bool {
	if p > 100 {
		p = 100
	}
	v := r.next31() % 100
	ok := v < p
	if r.trace {
		r.tracePos++
		f, err := os.OpenFile(r.traceFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err == nil {
			var b uint32
			if ok {
				b = 1
			}
			_, _ = fmt.Fprintf(f, "%d F %d -> %d\n", r.tracePos, p, b)
			_ = f.Close()
		}
	}
	return ok
}
