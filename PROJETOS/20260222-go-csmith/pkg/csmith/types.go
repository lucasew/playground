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
	// Mirrors Type::GenerateSimpleTypes order:
	// eChar, eSChar, eUChar, eShort, eUShort, eInt, eUInt,
	// eLong, eULong, eLongLong, eULongLong, eInt128, eUInt128.
	// Keep entries even when aliases collapse to same C type to preserve
	// upstream RNG selection cardinality.
	pool := make([]CType, 0, 13)
	pool = append(pool, CType{Name: "int8_t", Signed: true, Bits: 8})   // char
	pool = append(pool, CType{Name: "int8_t", Signed: true, Bits: 8})   // signed char
	pool = append(pool, CType{Name: "uint8_t", Signed: false, Bits: 8}) // unsigned char
	pool = append(pool, CType{Name: "int16_t", Signed: true, Bits: 16})
	pool = append(pool, CType{Name: "uint16_t", Signed: false, Bits: 16})
	pool = append(pool, hostIntType(opts))
	pool = append(pool, unsignedOf(hostIntType(opts).Bits))
	pool = append(pool, CType{Name: "int64_t", Signed: true, Bits: 64})   // long
	pool = append(pool, CType{Name: "uint64_t", Signed: false, Bits: 64}) // unsigned long
	if opts.LongLong {
		pool = append(pool, CType{Name: "int64_t", Signed: true, Bits: 64})
		pool = append(pool, CType{Name: "uint64_t", Signed: false, Bits: 64})
	}
	pool = append(pool, CType{Name: "__int128", Signed: true, Bits: 128})
	pool = append(pool, CType{Name: "unsigned __int128", Signed: false, Bits: 128})
	return pool
}

func pickType(r *rng, pool []CType) CType {
	return pool[int(r.upto(uint32(len(pool))))]
}

func castLiteral(t CType, expr string) string {
	return fmt.Sprintf("((%s)(%s))", t.Name, expr)
}
