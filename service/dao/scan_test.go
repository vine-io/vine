package dao

import (
	"testing"
)

func TestFieldPatch(t *testing.T) {

	type M2 struct {
		AA string `json:"aa"`
	}

	type S1 struct {
		M2   *M2                          `json:"m2"`
		Name string                       `json:"name"`
		M1   map[int]*M2                  `json:"m1"`
		MM   map[string]map[string]string `json:"mm"`
		Age  *int64                       `json:"age"`
	}

	s1Ins := &S1{
		Name: "s1",
		M1: map[int]*M2{1: &M2{AA: "11"}},
		MM: map[string]map[string]string{"22": {"aa": "22"}},
		//Age:  new(int64),
		//M2:   &M2{},
	}
	//*s1Ins.Age = 22

	s1InsP := FieldPatch(s1Ins)

	printMap(t, s1InsP)
}

func printMap(t *testing.T, m map[string]interface{}) {
	for k, v := range m {
		t.Logf("%s = %v", k, v)
	}
}
