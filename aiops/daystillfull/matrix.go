package daystillfull

import (
	"fmt"
	"strings"
)

type Mat [][]float32

func (m *Mat) Set(i int, j int, f float32) {
	r := (*m)[i]
	r[j] = f
}

// T Transposes the receiver
func (m *Mat) T() *Mat {
	r1 := m.NumRows()
	c1 := m.NumCols()
	t := NewMatrix(c1, r1)
	for i := 0; i < r1; i++ {
		for j := 0; j < c1; j++ {
			t.Set(j, i, (*m)[i][j])
		}
	}
	return t
}

func (m *Mat) Reciprocal() *Mat {
	r1 := m.NumRows()
	c1 := m.NumCols()
	s := NewMatrix(r1, c1)
	for i := 0; i < r1; i++ {
		for j := 0; j < c1; j++ {
			s.Set(i, j, 1/(*m)[i][j])
		}
	}
	return s
}

func (m *Mat) Scale(f float32) *Mat {
	r1 := m.NumRows()
	c1 := m.NumCols()
	s := NewMatrix(r1, c1)
	for i := 0; i < r1; i++ {
		for j := 0; j < c1; j++ {
			s.Set(i, j, f*(*m)[i][j])
		}
	}
	return s
}

func (m *Mat) DotProduct(b *Mat) *Mat {
	r1 := m.NumRows()
	r2 := b.NumRows()
	c2 := b.NumCols()

	n := NewMatrix(r1, c2)

	for i := 0; i < r1; i++ {
		for j := 0; j < c2; j++ {
			for k := 0; k < r2; k++ {
				(*n)[i][j] += (*m)[i][k] * (*b)[k][j]
			}
		}
	}

	return n
}

func (m *Mat) NumRows() int {
	return len(*m)
}

func (m *Mat) NumCols() int {
	r := (*m)[0]
	return len(r)
}

func (m *Mat) Add(f float32) *Mat {
	r1 := m.NumRows()
	c1 := m.NumCols()

	n := NewMatrix(r1, c1)

	for i := 0; i < r1; i++ {
		for j := 0; j < c1; j++ {
			n.Set(i, j, (*m)[i][j]+f)
		}
	}

	return n
}

func (m *Mat) Plus(b *Mat) *Mat {
	r1 := m.NumRows()
	c1 := m.NumCols()

	n := NewMatrix(r1, c1)

	for i := 0; i < r1; i++ {
		for j := 0; j < c1; j++ {
			n.Set(i, j, (*m)[i][j]+(*b)[i][j])
		}
	}

	return n
}

func (m *Mat) Minus(b *Mat) *Mat {
	r1 := m.NumRows()
	c1 := m.NumCols()

	n := NewMatrix(r1, c1)

	for i := 0; i < r1; i++ {
		for j := 0; j < c1; j++ {
			n.Set(i, j, (*m)[i][j]-(*b)[i][j])
		}
	}

	return n
}

func (m *Mat) flt() float32 {
	return (*m)[0][0]
}

func (m *Mat) String() string {
	r1 := m.NumRows()
	c1 := m.NumCols()
	b := strings.Builder{}

	for i := 0; i < r1; i++ {
		for j := 0; j < c1; j++ {
			b.WriteString(fmt.Sprintf("%4f ", (*m)[i][j]))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func NewMatrix(n, m int) *Mat {
	x := make(Mat, n)
	rows := make([]float32, n*m) // ensure contiguous memory
	for i, startRow := 0, 0; i < n; i, startRow = i+1, startRow+m {
		endRow := startRow + m
		x[i] = rows[startRow:endRow:endRow]
	}
	return &x
}

func Identity(n int) *Mat {
	matrix := NewMatrix(n, n)
	for i, r := range *matrix {
		r[i] = 1.0
	}
	return matrix
}
