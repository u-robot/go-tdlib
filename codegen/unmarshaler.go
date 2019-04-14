package codegen

import (
	"bytes"
	"fmt"

	"github.com/u-robot/go-tdlib/tlparser"
)

// GenerateUnmarshalers generates source code from the Telegram API scheme.
func GenerateUnmarshalers(schema *tlparser.Schema, packageName string) []byte {
	buf := bytes.NewBufferString("")

	buf.WriteString(fmt.Sprintf("%s\n\npackage %s\n\n", header, packageName))

	buf.WriteString(`import (
    "encoding/json"
    "fmt"
)

`)

	for _, class := range schema.Classes {
		tdlibClass := NewTdlibClass(class.Name, schema)

		buf.WriteString(fmt.Sprintf(`// Unmarshal%s parses the JSON-encoded data and return it as %s object.
func Unmarshal%s(data json.RawMessage) (%s, error) {
    var meta meta

    err := json.Unmarshal(data, &meta)
    if err != nil {
        return nil, err
    }

    switch meta.Type {
`, tdlibClass.ToGoType(), tdlibClass.ToGoType(), tdlibClass.ToGoType(), tdlibClass.ToGoType()))

		for _, subType := range tdlibClass.GetSubTypes() {
			buf.WriteString(fmt.Sprintf(`    case %s:
        return Unmarshal%s(data)

`, subType.ToTypeConst(), subType.ToGoType()))

		}

		buf.WriteString(`    default:
        return nil, fmt.Errorf("Error unmarshaling. Unknown type: " +  meta.Type)
    }
}

`)
	}

	for _, typ := range schema.Types {
		tdlibType := NewTdlibType(typ.Name, schema)

		if tdlibType.IsList() || tdlibType.IsInternal() {
			continue
		}

		buf.WriteString(fmt.Sprintf(`// Unmarshal%s parses the JSON-encoded data and return it as %s object.
func Unmarshal%s(data json.RawMessage) (*%s, error) {
    var response %s

    err := json.Unmarshal(data, &response)

    return &response, err
}

`, tdlibType.ToGoType(), tdlibType.ToGoType(), tdlibType.ToGoType(), tdlibType.ToGoType(), tdlibType.ToGoType()))

	}

	buf.WriteString(`// UnmarshalType parses the JSON-encoded data and return it as %s object.
func UnmarshalType(data json.RawMessage) (Type, error) {
    var meta meta

    err := json.Unmarshal(data, &meta)
    if err != nil {
        return nil, err
    }

    switch meta.Type {
`)

	for _, typ := range schema.Types {
		tdlibType := NewTdlibType(typ.Name, schema)

		if tdlibType.IsList() || tdlibType.IsInternal() {
			continue
		}

		buf.WriteString(fmt.Sprintf(`    case %s:
        return Unmarshal%s(data)

`, tdlibType.ToTypeConst(), tdlibType.ToGoType()))

	}

	buf.WriteString(`    default:
        return nil, fmt.Errorf("Error unmarshaling. Unknown type: " +  meta.Type)
    }
}
`)

	return buf.Bytes()
}
