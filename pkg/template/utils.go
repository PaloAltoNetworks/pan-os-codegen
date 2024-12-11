package template

import "fmt"

func TemplateMap(elts ...any) (map[string]any, error) {
	if len(elts)%2 != 0 {
		return nil, fmt.Errorf("templateMap: number of arguments must be divisible by 2")
	}

	mapped := make(map[string]any, len(elts)/2)
	for i := 0; i < len(elts); i += 2 {
		mapKey, ok := elts[i].(string)
		if !ok {
			return nil, fmt.Errorf("templateMap: keys must be strings")
		}
		mapValue := elts[i+1]

		mapped[mapKey] = mapValue
	}
	return mapped, nil
}
