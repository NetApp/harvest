//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package matrix

// Calculate M - N, such that M is a metric from
// our data, and n is from N
func (m *Matrix) Delta(PrevData *Matrix, metricIndex int) error {

    if len(m.Instances) != len(PrevData.Instances) {
        panic("invalid delta operation")
    }
    for k := 0; k < len(m.Instances); k += 1 {
        m.Data[metricIndex][k] -= PrevData.Data[metricIndex][k]
    }
    return nil
}

func (m *Matrix) Divide(numeratorIndex, denominatorIndex int, threshold float64) error {

    for k := 0; k < len(m.Instances); k += 1 {
        if m.Data[denominatorIndex][k] <= threshold {
            m.Data[numeratorIndex][k] = NAN
        } else {
            m.Data[numeratorIndex][k] /= m.Data[denominatorIndex][k]
        }
    }
    return nil
}

func (m *Matrix) MultByScalar(metricIndex int, scalarValue float64) {
    for k := 0; k < len(m.Instances); k += 1 {
        m.Data[metricIndex][k] *= scalarValue
    }
}

func (m *Matrix) InstanceWiseAddition(toInstance, fromInstance *Instance, fromData *Matrix) {

    for i := 0; i < len(m.Metrics); i += 1 {
        if m.Data[i][toInstance.Index] == m.Data[i][toInstance.Index] {
            m.Data[i][toInstance.Index] += fromData.Data[i][fromInstance.Index]
        } else {
            m.Data[i][toInstance.Index] = fromData.Data[i][fromInstance.Index]
        }
    }
}
