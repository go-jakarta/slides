package main

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

type Table[T ~string, K constraints.Ordered] struct {
	ID      T
	Amounts []K
}

func NewTable[T ~string, K constraints.Ordered](id T, amounts ...K) *Table[T, K] {
	return &Table[T, K]{
		ID:      id,
		Amounts: amounts,
	}
}

func (t Table[T, K]) Sum() K {
	var i K
	for _, a := range t.Amounts {
		i = i + a
	}
	return i
}

func (t Table[T, K]) Print() {
	sum := t.Sum()
	fmt.Printf("%q (%T): %v\n", t.ID, sum, sum)
}

// END OMIT

func main() {
	t := NewTable("table1", 15, 20)
	t.Print()
	u := NewTable("table2", 15.0, 25.5)
	u.Print()
}
