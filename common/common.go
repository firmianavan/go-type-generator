package common

import (
	"strings"
)

func ChangeNameToCamel(oriName, spliter string) string {
	strs := strings.Split(oriName, spliter)
	for i, v := range strs {
		strs[i] = strings.Title(v)
	}
	return strings.Join(strs, "")
}
