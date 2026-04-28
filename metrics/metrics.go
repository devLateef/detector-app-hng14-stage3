package metrics

var data = map[string]interface{}{}

func Set(key string, val interface{}) {
	data[key] = val
}

func Get() map[string]interface{} {
	return data
}
