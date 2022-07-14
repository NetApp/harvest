/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

func (me *Matrix) InstanceWiseAdditionUint64(toInstance, fromInstance *Instance, fromData *Matrix) {
	for key, fromMetric := range fromData.GetMetrics() {
		if toMetric := me.GetMetric(key); toMetric != nil {
			fromValue, _ := fromMetric.GetValueUint64(fromInstance)
			toValue, _ := toMetric.GetValueUint64(toInstance)
			_ = toMetric.SetValueUint64(toInstance, fromValue+toValue)
		}
	}
}
