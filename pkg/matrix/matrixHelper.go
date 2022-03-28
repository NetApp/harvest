package matrix

func CreateMetric(key string, data *Matrix) error {
	var err error
	at := data.GetMetric(key)
	if at == nil {
		if _, err = data.NewMetricFloat64(key); err != nil {
			return err
		}
	}
	return nil
}
