package main

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

func Sum[T constraints.Ordered](v ...T) T {
	var sum T
	for _, i := range v {
		sum += i
	}
	return sum
}

type SumFn[T constraints.Ordered] func(...T) T

type Ledger[T ~string, K constraints.Ordered] struct {
	ID      T
	Amounts []K
	SumFn   SumFn[K]
}

func (l Ledger[T, K]) Print() {
	sum := l.SumFn(l.Amounts...)
	fmt.Printf("%s (%T): %v\n", l.ID, sum, sum)
}

// END OMIT

type Ledgerish[T ~string, K constraints.Ordered] interface {
	~struct {
		ID      T
		Amounts []K
		SumFn   SumFn[K]
	}
	Print()
}

func Print[T ~string, K constraints.Ordered, L Ledgerish[T, K]](l L) {
	l.Print()
}

func main() {
	Print(Ledger[string, float64]{
		ID:      "ledger1",
		Amounts: []float64{1, 2, 3},
		SumFn:   Sum[float64],
	})
	Print(Ledger[string, int]{
		ID:      "ledger2",
		Amounts: []int{-1, -2, -3},
		SumFn:   Sum[int],
	})
}
