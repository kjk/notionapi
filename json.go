package notionapi

import "encoding/json"

func jsonUnmarshalFromMap(m map[string]interface{}, v interface{}) error {
	d, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, v)
}

func jsonGetString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func jsonGetMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if s, ok := v.(map[string]interface{}); ok {
			return s
		}
	}
	return nil
}
