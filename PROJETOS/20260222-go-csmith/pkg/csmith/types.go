package csmith

import "fmt"

// CType is a lightweight C scalar type descriptor used by the current generator.
type CType struct {
	Name   string
	Signed bool
	Bits   int
}

func hostIntType(opts Options) CType {
	switch opts.IntSize {
	case 1:
		return CType{Name: "int8_t", Signed: true, Bits: 8}
	case 2:
		return CType{Name: "int16_t", Signed: true, Bits: 16}
	case 8:
		return CType{Name: "int64_t", Signed: true, Bits: 64}
	default:
		return CType{Name: "int32_t", Signed: true, Bits: 32}
	}
}

func unsignedOf(bits int) CType {
	switch bits {
	case 8:
		return CType{Name: "uint8_t", Signed: false, Bits: 8}
	case 16:
		return CType{Name: "uint16_t", Signed: false, Bits: 16}
	case 64:
		return CType{Name: "uint64_t", Signed: false, Bits: 64}
	default:
		return CType{Name: "uint32_t", Signed: false, Bits: 32}
	}
}

func typePool(opts Options) []CType {
	pool := []CType{
		hostIntType(opts),
		unsignedOf(hostIntType(opts).Bits),
		CType{Name: "int16_t", Signed: true, Bits: 16},
		CType{Name: "uint32_t", Signed: false, Bits: 32},
	}

	if opts.Int8 {
		pool = append(pool, CType{Name: "int8_t", Signed: true, Bits: 8})
	}
	if opts.UInt8 {
		pool = append(pool, CType{Name: "uint8_t", Signed: false, Bits: 8})
	}
	if opts.LongLong && opts.Math64 {
		pool = append(pool,
			CType{Name: "int64_t", Signed: true, Bits: 64},
			CType{Name: "uint64_t", Signed: false, Bits: 64},
		)
	}
	return pool
}

func pickType(r *rng, pool []CType) CType {
	return pool[int(r.upto(uint32(len(pool))))]
}

func castLiteral(t CType, expr string) string {
	return fmt.Sprintf("((%s)(%s))", t.Name, expr)
}
