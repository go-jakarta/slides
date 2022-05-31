package main

import (
	"fmt"

	"github.com/makiuchi-d/linq/v2"
)

type Student struct {
	Name  string
	Class string
	Score int
}

func main() {
	students := []Student{
		{"one", "1-A", 953},
		{"two", "1-B", 559},
		{"three", "1-B", 1136},
	}
	e1 := linq.FromSlice(students)
	e2 := linq.Where(e1, func(s Student) (bool, error) { return s.Class == "1-B", nil })
	e3 := linq.OrderByDescending(e2, func(s Student) (int, error) { return s.Score, nil })
	linq.ForEach(e3, func(s Student) error {
		fmt.Printf("%d %s\n", s.Score, s.Name)
		return nil
	})
}
