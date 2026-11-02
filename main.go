package main

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"text/tabwriter"
	"unsafe"
)

type noscan16 struct{ a, b uint64 }
type noscan24 struct{ a, b, c uint64 }
type noscan32 struct{ a, b, c, d uint64 }
type noscan48 struct{ v [6]uint64 }
type noscan64 struct{ v [8]uint64 }
type noscan80 struct{ v [10]uint64 }
type noscan96 struct{ v [12]uint64 }
type noscan256 struct{ v [32]uint64 }

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

var sink any

var classes = [...]uintptr{8, 16, 24, 32, 48, 64, 80}

type benchCase struct {
	name string
	size uintptr
	ptrs bool
	fn   func(b *testing.B)
}

var cases = []benchCase{
	{"noscan16", unsafe.Sizeof(noscan16{}), false, func(b *testing.B) {
		for range b.N {
			sink = new(noscan16)
		}
	}},
	{"noscan24", unsafe.Sizeof(noscan24{}), false, func(b *testing.B) {
		for range b.N {
			sink = new(noscan24)
		}
	}},
	{"noscan32", unsafe.Sizeof(noscan32{}), false, func(b *testing.B) {
		for range b.N {
			sink = new(noscan32)
		}
	}},
	{"noscan48", unsafe.Sizeof(noscan48{}), false, func(b *testing.B) {
		for range b.N {
			sink = new(noscan48)
		}
	}},
	{"noscan64", unsafe.Sizeof(noscan64{}), false, func(b *testing.B) {
		for range b.N {
			sink = new(noscan64)
		}
	}},
	{"noscan80", unsafe.Sizeof(noscan80{}), false, func(b *testing.B) {
		for range b.N {
			sink = new(noscan80)
		}
	}},
	{"noscan96", unsafe.Sizeof(noscan96{}), false, func(b *testing.B) {
		for range b.N {
			sink = new(noscan96)
		}
	}},
	{"noscan256", unsafe.Sizeof(noscan256{}), false, func(b *testing.B) {
		for range b.N {
			sink = new(noscan256)
		}
	}},
	{"scan16", unsafe.Sizeof(scan16{}), true, func(b *testing.B) {
		for range b.N {
			sink = new(scan16)
		}
	}},
	{"scan24", unsafe.Sizeof(scan24{}), true, func(b *testing.B) {
		for range b.N {
			sink = new(scan24)
		}
	}},
	{"scan32", unsafe.Sizeof(scan32{}), true, func(b *testing.B) {
		for range b.N {
			sink = new(scan32)
		}
	}},
	{"scan48", unsafe.Sizeof(scan48{}), true, func(b *testing.B) {
		for range b.N {
			sink = new(scan48)
		}
	}},
	{"scan64", unsafe.Sizeof(scan64{}), true, func(b *testing.B) {
		for range b.N {
			sink = new(scan64)
		}
	}},
	{"scan80", unsafe.Sizeof(scan80{}), true, func(b *testing.B) {
		for range b.N {
			sink = new(scan80)
		}
	}},
	{"scan96", unsafe.Sizeof(scan96{}), true, func(b *testing.B) {
		for range b.N {
			sink = new(scan96)
		}
	}},
}

func sizeClass(size uintptr) (int, bool) {
	for i, c := range classes {
		if size <= c {
			return i + 1, size > 0
		}
	}
	return 0, false
}

func main() {
	fmt.Printf("Go version : %s\nPlatform   : %s/%s\nGOMAXPROCS : %d\n\n",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TYPE\tSIZE\tSIZE CLASS\tELIGIBLE (<=80B)?\tNS/OP\tB/OP\tALLOCS/OP")
	fmt.Fprintln(w, "----\t----\t----------\t----------------\t-----\t----\t---------")

	for _, c := range cases {
		res := testing.Benchmark(c.fn)

		class, specialized := "-", "no"
		if sc, ok := sizeClass(c.size); ok {
			class, specialized = fmt.Sprint(sc), "yes"
		}
		kind := "noscan"
		if c.ptrs {
			kind = "scan"
		}

		fmt.Fprintf(w, "%s (%s)\t%dB\t%s\t%s\t%.1f\t%d\t%d\n",
			c.name, kind, c.size, class, specialized,
			float64(res.T.Nanoseconds())/float64(res.N), res.AllocedBytesPerOp(), res.AllocsPerOp())
	}
	w.Flush()
}
