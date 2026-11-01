// main.go
//
// A small, self-contained demo for the Go blog post "Size-Specialized Memory
// Allocation" (https://go.dev/blog/size-specialized-allocations).
//
// It allocates heap objects of many different sizes, with and without
// pointers, and reports ns/op, B/op and allocs/op for each one.
//
// Run it with:
//
//	go run main.go
//
// To compare with and without size-specialized malloc (Go 1.27+):
//
//	go build -o with_ssm .
//	GOEXPERIMENT=nosizespecializedmalloc go build -o without_ssm .
//	./with_ssm && ./without_ssm
//
// Note: numbers are machine dependent and noisy. Run each binary a few
// times (or use benchstat) before drawing conclusions.
package main

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"text/tabwriter"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Pointer-free ("noscan") types, one per size class boundary.
// Size classes up to 80 bytes: 8, 16, 24, 32, 48, 64, 80.
// ---------------------------------------------------------------------------

type noscan16 struct{ a, b uint64 }
type noscan24 struct{ a, b, c uint64 }
type noscan32 struct{ a, b, c, d uint64 }
type noscan48 struct{ v [6]uint64 }
type noscan64 struct{ v [8]uint64 }
type noscan80 struct{ v [10]uint64 }
type noscan96 struct{ v [12]uint64 } // just above the specialized range
type noscan256 struct{ v [32]uint64 }

// ---------------------------------------------------------------------------
// Types that contain pointers ("scan"). The GC must track where the pointers
// are, which is part of the bookkeeping that specialized functions speed up.
// ---------------------------------------------------------------------------

type scan16 struct {
	p *int
	x uint64
}
type scan24 struct {
	p *int
	x [2]uint64
}
type scan32 struct {
	p *int
	x [3]uint64
}
type scan48 struct {
	p *int
	x [5]uint64
}
type scan64 struct {
	p *int
	x [7]uint64
}
type scan80 struct {
	p *int
	x [9]uint64
}
type scan96 struct {
	p *int
	x [11]uint64
}

// sink forces allocations to escape to the heap. Without this, the compiler's
// escape analysis would put the objects on the stack and nothing would be
// measured.
var sink any

// benchCase describes one allocation benchmark.
type benchCase struct {
	name string
	size uintptr
	ptrs bool
	fn   func(b *testing.B)
}

// NOTE: each benchmark is written out by hand, with a concrete type, on
// purpose. If we used a generic helper, the compiler might not know the size
// of the allocation at compile time, which is exactly the situation where it
// can't call a size-specialized function directly.
func cases() []benchCase {
	return []benchCase{
		{"noscan16", unsafe.Sizeof(noscan16{}), false, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(noscan16)
			}
		}},
		{"noscan24", unsafe.Sizeof(noscan24{}), false, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(noscan24)
			}
		}},
		{"noscan32", unsafe.Sizeof(noscan32{}), false, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(noscan32)
			}
		}},
		{"noscan48", unsafe.Sizeof(noscan48{}), false, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(noscan48)
			}
		}},
		{"noscan64", unsafe.Sizeof(noscan64{}), false, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(noscan64)
			}
		}},
		{"noscan80", unsafe.Sizeof(noscan80{}), false, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(noscan80)
			}
		}},
		{"noscan96", unsafe.Sizeof(noscan96{}), false, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(noscan96)
			}
		}},
		{"noscan256", unsafe.Sizeof(noscan256{}), false, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(noscan256)
			}
		}},

		{"scan16", unsafe.Sizeof(scan16{}), true, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(scan16)
			}
		}},
		{"scan24", unsafe.Sizeof(scan24{}), true, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(scan24)
			}
		}},
		{"scan32", unsafe.Sizeof(scan32{}), true, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(scan32)
			}
		}},
		{"scan48", unsafe.Sizeof(scan48{}), true, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(scan48)
			}
		}},
		{"scan64", unsafe.Sizeof(scan64{}), true, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(scan64)
			}
		}},
		{"scan80", unsafe.Sizeof(scan80{}), true, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(scan80)
			}
		}},
		{"scan96", unsafe.Sizeof(scan96{}), true, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = new(scan96)
			}
		}},
	}
}

// sizeClass returns the (approximate) runtime size class for sizes up to 80
// bytes, matching the table in the blog post. It returns 0 if the size is
// outside the range covered by size-specialized allocation.
func sizeClass(size uintptr) int {
	switch {
	case size == 0:
		return 0
	case size <= 8:
		return 1
	case size <= 16:
		return 2
	case size <= 24:
		return 3
	case size <= 32:
		return 4
	case size <= 48:
		return 5
	case size <= 64:
		return 6
	case size <= 80:
		return 7
	default:
		return 0
	}
}

func main() {
	fmt.Printf("Go version : %s\n", runtime.Version())
	fmt.Printf("Platform   : %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("GOMAXPROCS : %d\n\n", runtime.GOMAXPROCS(0))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TYPE\tSIZE\tSIZE CLASS\tELIGIBLE (<=80B)?\tNS/OP\tB/OP\tALLOCS/OP")
	fmt.Fprintln(w, "----\t----\t----------\t----------------\t-----\t----\t---------")

	for _, c := range cases() {
		res := testing.Benchmark(c.fn)

		class := "-"
		specialized := "no"
		if sc := sizeClass(c.size); sc != 0 {
			class = fmt.Sprintf("%d", sc)
			specialized = "yes"
		}
		kind := "noscan"
		if c.ptrs {
			kind = "scan"
		}

		fmt.Fprintf(w, "%s (%s)\t%dB\t%s\t%s\t%.1f\t%d\t%d\n",
			c.name, kind, c.size, class, specialized,
			float64(res.T.Nanoseconds())/float64(res.N),
			res.AllocedBytesPerOp(),
			res.AllocsPerOp(),
		)
	}
	w.Flush()

	fmt.Println()
	fmt.Println("Tip: build once normally and once with")
	fmt.Println("  GOEXPERIMENT=nosizespecializedmalloc")
	fmt.Println("(Go 1.27+) and compare the 16B and 24B rows.")
}
