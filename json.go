package notionapi

import "encoding/json"

func jsonUnmarshalFromMap(m map[string]interface{}, v interface{}) error {
	d, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, v)
}

func jsonGetMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

/*
func jsonGetArray(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key]; ok {
		if a, ok := v.([]interface{}); ok {
			return a
		}
	}
	return nil
}
*/
