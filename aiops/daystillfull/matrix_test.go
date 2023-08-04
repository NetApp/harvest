package daystillfull

import (
	"testing"
)

func TestMat_Multiple(t *testing.T) {
	a := NewMatrix(3, 3)
	a.Set(0, 0, 5)
	a.Set(0, 1, 1)
	a.Set(0, 2, 3)
	a.Set(1, 0, 1)
	a.Set(1, 1, 1)
	a.Set(1, 2, 1)
	a.Set(2, 0, 1)
	a.Set(2, 1, 2)
	a.Set(2, 2, 1)

	b := NewMatrix(3, 3)
	b.Set(0, 0, 1)
	b.Set(0, 1, 2)
	b.Set(0, 2, 3)
	b.Set(1, 0, 1)
	b.Set(1, 1, 2)
	b.Set(1, 2, 3)
	b.Set(2, 0, 2)
	b.Set(2, 1, 3)
	b.Set(2, 2, 4)
	dp := a.DotProduct(b)

	want := NewMatrix(3, 3)
	want.Set(0, 0, 12.0)
	want.Set(0, 1, 21.0)
	want.Set(0, 2, 30.0)
	want.Set(1, 0, 4.0)
	want.Set(1, 1, 7.0)
	want.Set(1, 2, 10.0)
	want.Set(2, 0, 5.0)
	want.Set(2, 1, 9.0)
	want.Set(2, 2, 13.0)

	r := dp.NumRows()
	c := dp.NumCols()
	if r != 3 {
		t.Errorf("got numRows=%d want=3", r)
	}
	if c != 3 {
		t.Errorf("got numCols=%d want=3", c)
	}

	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if (*dp)[i][j] != (*want)[i][j] {
				t.Errorf("dotProduct at i=%d j=%d got=%f want=%f", i, j, (*dp)[i][j], (*want)[i][j])

			}
		}
	}
}
