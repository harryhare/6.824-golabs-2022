package util

func Format(x interface{}) []byte {
	data, ok := x.([]byte)
	if ok {
		return data
	}
	a, ok := x.([][]byte)
	if ok {
		result := []byte("[")
		for _, data := range a {
			result = append(result, data...)
			result = append(result, ',')
		}
		n := len(result)
		result[n-1] = ']'
		return result
	}
	return nil
}
