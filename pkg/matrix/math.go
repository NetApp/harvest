/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

func (my *Matrix) InstanceWiseAdditionUint64(toInstance, fromInstance *Instance, fromData *Matrix) {
	for key, fromMetric := range fromData.GetMetrics() {
		if toMetric := my.GetMetric(key); toMetric != nil {
			fromValue, _ := fromMetric.GetValueUint64(fromInstance)
			toValue, _ := toMetric.GetValueUint64(toInstance)
			_ = toMetric.SetValueUint64(toInstance, fromValue+toValue)
		}
	}
}
