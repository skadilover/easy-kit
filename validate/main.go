package validate

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

type Property map[string]interface{}

func (p *Property) Property(key string, sub *Property) *Property {
	ps, ok := (*p)["properties"]
	if !ok {
		ps = map[string]interface{}{}
	}
	dict := ps.(map[string]interface{})
	dict[key] = sub
	(*p)["properties"] = dict
	return p
}
func (p *Property) Properties(m map[string]*Property) *Property {
	ps, ok := (*p)["properties"]
	if !ok {
		ps = map[string]interface{}{}
	}
	dict := ps.(map[string]interface{})
	for k, v := range m {
		dict[k] = v
	}
	(*p)["properties"] = dict
	return p
}

func (p *Property) Pattern(pattern string) *Property {
	(*p)["pattern"] = pattern
	return p
}

func (p *Property) Items(m *Property) *Property {
	(*p)["items"] = m
	return p
}

func (p *Property) MinItems(m int) *Property {
	(*p)["minItems"] = m
	return p
}

func (p *Property) MaxItems(m int) *Property {
	(*p)["maxItems"] = m
	return p
}

func (p *Property) Required(keys ...string) *Property {
	array, ok := (*p)["required"]
	var required []string
	if ok {
		required = array.([]string)
	}
	for _, v := range keys {
		required = append(required, v)
	}
	(*p)["required"] = required
	return p
}
func (p *Property) Type(t string) *Property {
	(*p)["type"] = t
	return p
}
func (p *Property) Min(min int) *Property {
	(*p)["minimum"] = min
	return p
}

func (p *Property) Max(max int) *Property {
	(*p)["maximum"] = max
	return p
}

func (p *Property) MinLength(min int) *Property {
	(*p)["minLength"] = min
	return p
}
func (p *Property) MaxLength(max int) *Property {
	(*p)["maxLength"] = max
	return p
}

func (p *Property) Validate(obj interface{}) error {
	sl := gojsonschema.NewGoLoader(p)
	ol := gojsonschema.NewGoLoader(obj)
	schema, err := gojsonschema.NewSchema(sl)
	if err != nil {
		return err
	}
	result, err := schema.Validate(ol)
	if err != nil {
		return err
	}
	if !result.Valid() {
		return fmt.Errorf("valid failed:%s", result.Errors())
	}
	return nil
}

func (p *Property) ValidateString(text string) error {
	sl := gojsonschema.NewGoLoader(p)
	ol := gojsonschema.NewStringLoader(text)
	schema, err := gojsonschema.NewSchema(sl)
	if err != nil {
		return err
	}
	result, err := schema.Validate(ol)
	if err != nil {
		return err
	}
	if !result.Valid() {
		return fmt.Errorf("valid failed:%s", result.Errors())
	}
	return nil
}

func NewProperty(t string) *Property {
	a := make(Property)
	a["type"] = t
	return &a
}

const (
	TypeString     = "string"
	TypeNumber     = "number"
	TypeBoolean    = "boolean"
	TypeArray      = "array"
	TypeValue      = "value"
	TypeObject     = "object"
	TypeWhitespace = "whitespace"
	TypeNull       = "null"
	TypeInteger    = "integer"
)
