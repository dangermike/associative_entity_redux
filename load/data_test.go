package load

import (
	"slices"
	"strconv"
	"testing"
)

func TestJoiner(t *testing.T) {
	j := joiner(10000, 10000)
	trg := make([][2]int, 1_000_000)
	j(trg)
	slices.SortFunc(trg, ComparePair)
	for i := 1; i < len(trg); i++ {
		if ComparePair(trg[i-1], trg[i]) == 0 {
			t.Fatal("collision")
		}
	}
}

func TestFillWords(t *testing.T) {
	trg := make([]string, 100000)
	fillWords(trg)
	slices.Sort(trg)
	for i := 1; i < len(trg); i++ {
		if trg[i] == trg[i-1] {
			t.Fatal("collision")
		}
	}
}

func BenchmarkJoiner(b *testing.B) {
	for _, test := range []int{1000, 1000000, 5000000} {
		const batchSize = 1000
		trg := make([][2]int, 1000)
		b.Run(strconv.Itoa(test), func(b *testing.B) {
			j := joiner(1000000, 1000000)
			for b.Loop() {
				for i := 0; i < test; i += batchSize {
					j(trg)
				}
			}
		})
	}
}
