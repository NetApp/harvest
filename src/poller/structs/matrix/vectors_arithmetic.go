package main

import "errors"

// Calculate M - N, such that M is a metric from
// our data, and n is from N
func (m *Matrix) Delta(x *Matrix, i, j int) error {

	if len(m.Instances) != len(x.Instances) {
		return errors.New("invalid delta operation")
	}
	for k:=0; k<len(m.Instances); k+=1 {
		m.Data[i][k] -= x.Data[j][k]
	}
	return nil
}

func (m *Matrix) Divide(x *Matrix, i, j int) error {

	if len(m.Instances) != len(x.Instances) {
		return errors.New("invalid delta operation")
	}
	for k:=0; k<len(m.Instances); k+=1 {
		m.Data[i][k] /= x.Data[i][k]
	}
	return nil
}

func (m *Matrix) MultByScalar(i int, a float64) {
	for k:=0; k<len(m.Instances); k+=1 {
		m.Data[i][k] *= x
	}	
}