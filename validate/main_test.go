package validate

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_main(t *testing.T) {
	var sd *Property
	//validate data
	var obj = map[string]interface{}{
		"name": "123",
		"age":  29,
		"friends": []map[string]interface{}{
			map[string]interface{}{
				"name": "123",
			},
		},
	}
	//validate schema
	dict := map[string]*Property{
		"name": NewProperty(TypeString).MaxLength(3),
		"age":  NewProperty(TypeNumber).Max(99),
		"friends": NewProperty(TypeArray).Items(NewProperty(TypeObject).Properties(
			map[string]*Property{
				"name": NewProperty(TypeString).MaxLength(255).Pattern("^[0-9]*$"),
			},
		)).MaxItems(3).MinItems(1),
	}
	sd = NewProperty(TypeObject).Properties(dict).Required("name", "age")
	if err := sd.Validate(obj); err != nil {
		t.Error("valid object failed:", err.Error())
	}
	jData, _ := json.Marshal(obj)
	fmt.Printf("%s\n", jData)
	if err := sd.ValidateString(string(jData)); err != nil {
		t.Error("valid string failed:", err.Error())
	}
}
