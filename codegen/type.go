package codegen

import (
	"bytes"
	"fmt"

	"github.com/u-robot/go-tdlib/tlparser"
)

// GenerateTypes generates source code from the Telegram API scheme.
func GenerateTypes(schema *tlparser.Schema, packageName string) []byte {
	buf := bytes.NewBufferString("")

	buf.WriteString(fmt.Sprintf("%s\n\npackage %s\n\n", header, packageName))

	buf.WriteString(`import (
    "encoding/json"
)

`)

	buf.WriteString("// Class constants.\nconst (\n")
	for _, entity := range schema.Classes {
		tdlibClass := NewTdlibClass(entity.Name, schema)
		buf.WriteString(fmt.Sprintf("    %s = %q\n", tdlibClass.ToClassConst(), entity.Name))
	}
	for _, entity := range schema.Types {
		tdlibType := NewTdlibType(entity.Name, schema)
		if tdlibType.IsInternal() || tdlibType.HasClass() {
			continue
		}
		buf.WriteString(fmt.Sprintf("    %s = %q\n", tdlibType.ToClassConst(), entity.Class))
	}
	buf.WriteString(")")

	buf.WriteString("\n\n")

	buf.WriteString("// Type constants.\nconst (\n")
	for _, entity := range schema.Types {
		tdlibType := NewTdlibType(entity.Name, schema)
		if tdlibType.IsInternal() {
			continue
		}
		buf.WriteString(fmt.Sprintf("    %s = %q\n", tdlibType.ToTypeConst(), entity.Name))
	}
	buf.WriteString(")")

	buf.WriteString("\n\n")

	for _, class := range schema.Classes {
		tdlibClass := NewTdlibClass(class.Name, schema)

		buf.WriteString(fmt.Sprintf(`// %s %s
type %s interface {
    %sType() string
}

`, tdlibClass.ToGoType(), firstLower(class.Description), tdlibClass.ToGoType(), tdlibClass.ToGoType()))
	}

	for _, typ := range schema.Types {
		tdlibType := NewTdlibType(typ.Name, schema)
		if tdlibType.IsInternal() {
			continue
		}

		buf.WriteString(fmt.Sprintf("// %s %s\n", tdlibType.ToGoType(), firstLower(typ.Description)))

		if len(typ.Properties) > 0 {
			buf.WriteString(`type ` + tdlibType.ToGoType() + ` struct {
    meta
`)
			for _, property := range typ.Properties {
				tdlibTypeProperty := NewTdlibTypeProperty(property.Name, property.Type, schema)

				buf.WriteString(fmt.Sprintf("    // %s\n", property.Description))
				buf.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n", tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoType(), property.Name))
			}

			buf.WriteString("}\n\n")
		} else {
			buf.WriteString(`type ` + tdlibType.ToGoType() + ` struct{
    meta
}

`)
		}

		buf.WriteString(fmt.Sprintf(`// MarshalJSON returns %s object as the JSON encoding of %s.
func (entity *%s) MarshalJSON() ([]byte, error) {
    entity.meta.Type = entity.GetType()

    type stub %s

    return json.Marshal((*stub)(entity))
}

`, tdlibType.ToGoType(), tdlibType.ToGoType(), tdlibType.ToGoType(), tdlibType.ToGoType()))

		buf.WriteString(fmt.Sprintf(`// GetClass returns constant class string of the class.
func (*%s) GetClass() string {
    return %s
}

// GetType returns constant class type string of the class.
func (*%s) GetType() string {
    return %s
}

`, tdlibType.ToGoType(), tdlibType.ToClassConst(), tdlibType.ToGoType(), tdlibType.ToTypeConst()))

		if tdlibType.HasClass() {
			tdlibClass := NewTdlibClass(tdlibType.GetClass().Name, schema)

			buf.WriteString(fmt.Sprintf(`// %sType returns constant class type string of the class.
func (*%s) %sType() string {
    return %s
}

`, tdlibClass.ToGoType(), tdlibType.ToGoType(), tdlibClass.ToGoType(), tdlibType.ToTypeConst()))
		}

		if tdlibType.HasClassProperties() {
			buf.WriteString(fmt.Sprintf(`// UnmarshalJSON sets %s object to a copy of JSON encoding of %s.
func (entity *%s) UnmarshalJSON(data []byte) error {
    var tmp struct {
`, tdlibType.ToGoType(), tdlibType.ToGoType(), tdlibType.ToGoType()))

			var countSimpleProperties int

			for _, property := range typ.Properties {
				tdlibTypeProperty := NewTdlibTypeProperty(property.Name, property.Type, schema)

				if !tdlibTypeProperty.IsClass() || tdlibTypeProperty.IsList() {
					buf.WriteString(fmt.Sprintf("        %s %s `json:\"%s\"`\n", tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoType(), property.Name))
					countSimpleProperties++
				} else {
					buf.WriteString(fmt.Sprintf("        %s %s `json:\"%s\"`\n", tdlibTypeProperty.ToGoName(), "json.RawMessage", property.Name))
				}
			}

			buf.WriteString(`    }

    err := json.Unmarshal(data, &tmp)
    if err != nil {
        return err
    }

`)

			for _, property := range typ.Properties {
				tdlibTypeProperty := NewTdlibTypeProperty(property.Name, property.Type, schema)

				if !tdlibTypeProperty.IsClass() || tdlibTypeProperty.IsList() {
					buf.WriteString(fmt.Sprintf("    entity.%s = tmp.%s\n", tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoName()))
				}
			}

			if countSimpleProperties > 0 {
				buf.WriteString("\n")
			}

			for _, property := range typ.Properties {
				tdlibTypeProperty := NewTdlibTypeProperty(property.Name, property.Type, schema)

				if tdlibTypeProperty.IsClass() && !tdlibTypeProperty.IsList() {
					buf.WriteString(fmt.Sprintf(`    field%s, _ := Unmarshal%s(tmp.%s)
    entity.%s = field%s

`, tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoType(), tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoName()))
				}
			}

			buf.WriteString(`    return nil
}

`)
		}
	}

	return buf.Bytes()
}
