package maps

type Engine map[string]interface{}

func (that Engine) Fetch() (map[string]interface{}, error) {
	return that, nil
}
