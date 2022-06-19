package util

import (
	"github.com/iancoleman/strcase"
)

func ConvertMapKeysCaseToSnake(mapV map[string]interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{})
	for k, v := range mapV {
		resultMap[strcase.ToSnake(k)] = v
	}

	return resultMap
}
