package test

import (
	"github.com/firmianavan/go-type-generator/common"
	"testing"
)

func TestToCamel(t *testing.T) {
	cases := []struct {
		ori string
		des string
	}{
		{"", ""},
		{"abc", "Abc"},
		{"abc_de", "AbcDe"},
		{"a_d", "AD"},
		{"你", "你"},
	}

	for _, c := range cases {
		if tmp := common.ChangeNameToCamel(c.ori, "_"); tmp != c.des {
			t.Errorf("change failed: ori is %s , des is %s, expect %s", c.ori, tmp, c.des)
		} /* else {
		    fmt.Println(string(tmp))
		}*/
	}
}
