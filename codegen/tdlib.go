package codegen

import (
	"log"
	"strings"

	"github.com/u-robot/go-tdlib/tlparser"
)

// TdlibFunction contains TDLib function info.
type TdlibFunction struct {
	name   string
	schema *tlparser.Schema
}

// NewTdlibFunction returns TDLib function info.
func NewTdlibFunction(name string, schema *tlparser.Schema) *TdlibFunction {
	return &TdlibFunction{
		name:   name,
		schema: schema,
	}
}

// ToGoName returns string representation of name of TDLib function name in Go style.
func (entity *TdlibFunction) ToGoName() string {
	return firstUpper(entity.name)
}

// TdlibFunctionReturn contains TDLib function return info.
type TdlibFunctionReturn struct {
	name   string
	schema *tlparser.Schema
}

// NewTdlibFunctionReturn returns TDLib function return info.
func NewTdlibFunctionReturn(name string, schema *tlparser.Schema) *TdlibFunctionReturn {
	return &TdlibFunctionReturn{
		name:   name,
		schema: schema,
	}
}

// IsType returns true if TDLib function return is type.
func (entity *TdlibFunctionReturn) IsType() bool {
	return isType(entity.name, func(entity *tlparser.Type) string {
		return entity.Class
	}, entity.schema)
}

// GetType returns info about type of TDLib function return.
func (entity *TdlibFunctionReturn) GetType() *TdlibType {
	return getType(entity.name, func(entity *tlparser.Type) string {
		return entity.Class
	}, entity.schema)
}

// IsClass returns true if TDLib function return is class.
func (entity *TdlibFunctionReturn) IsClass() bool {
	return isClass(entity.name, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

// GetClass returns info about class of TDLib function return.
func (entity *TdlibFunctionReturn) GetClass() *TdlibClass {
	return getClass(entity.name, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

// ToGoReturn returns string representation of TDLib function return in Go style.
func (entity *TdlibFunctionReturn) ToGoReturn() string {
	if strings.HasPrefix(entity.name, "vector<") {
		log.Fatal("vectors are not supported")
	}

	if entity.IsClass() {
		return entity.GetClass().ToGoType()
	}

	if entity.GetType().IsInternal() {
		return entity.GetType().ToGoType()
	}

	return "*" + entity.GetType().ToGoType()
}

// ToGoType returns string representation of TDLib function return type in Go style.
func (entity *TdlibFunctionReturn) ToGoType() string {
	if strings.HasPrefix(entity.name, "vector<") {
		log.Fatal("vectors are not supported")
	}

	if entity.IsClass() {
		return entity.GetClass().ToGoType()
	}

	return entity.GetType().ToGoType()
}

// TdlibFunctionProperty contains TDLib function property info.
type TdlibFunctionProperty struct {
	name         string
	propertyType string
	schema       *tlparser.Schema
}

// NewTdlibFunctionProperty returns TDLib function property info.
func NewTdlibFunctionProperty(name string, propertyType string, schema *tlparser.Schema) *TdlibFunctionProperty {
	return &TdlibFunctionProperty{
		name:         name,
		propertyType: propertyType,
		schema:       schema,
	}
}

// GetPrimitive returns info about privitive type of TDLib function property.
func (entity *TdlibFunctionProperty) GetPrimitive() string {
	primitive := entity.propertyType

	for strings.HasPrefix(primitive, "vector<") {
		primitive = strings.TrimSuffix(strings.TrimPrefix(primitive, "vector<"), ">")
	}

	return primitive
}

// IsType returns true if TDLib function property is type.
func (entity *TdlibFunctionProperty) IsType() bool {
	primitive := entity.GetPrimitive()
	return isType(primitive, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

// GetType returns info about type of TDLib function property.
func (entity *TdlibFunctionProperty) GetType() *TdlibType {
	primitive := entity.GetPrimitive()
	return getType(primitive, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

// IsClass returns true if TDLib function property is class.
func (entity *TdlibFunctionProperty) IsClass() bool {
	primitive := entity.GetPrimitive()
	return isClass(primitive, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

// GetClass returns info about class of TDLib function property.
func (entity *TdlibFunctionProperty) GetClass() *TdlibClass {
	primitive := entity.GetPrimitive()
	return getClass(primitive, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

// ToGoName returns string representation of TDLib function property name in Go style.
func (entity *TdlibFunctionProperty) ToGoName() string {
	name := firstLower(underscoreToCamelCase(entity.name))
	if name == "type" {
		name += "Param"
	}

	return name
}

// ToGoType returns string representation of TDLib function property type in Go style.
func (entity *TdlibFunctionProperty) ToGoType() string {
	TdlibType := entity.propertyType
	goType := ""

	for strings.HasPrefix(TdlibType, "vector<") {
		goType = goType + "[]"
		TdlibType = strings.TrimSuffix(strings.TrimPrefix(TdlibType, "vector<"), ">")
	}

	if entity.IsClass() {
		return goType + entity.GetClass().ToGoType()
	}

	if entity.GetType().IsInternal() {
		return goType + entity.GetType().ToGoType()
	}

	return goType + "*" + entity.GetType().ToGoType()
}

// TdlibType contains TDLib type info.
type TdlibType struct {
	name   string
	schema *tlparser.Schema
}

// NewTdlibType returns TDLib type info.
func NewTdlibType(name string, schema *tlparser.Schema) *TdlibType {
	return &TdlibType{
		name:   name,
		schema: schema,
	}
}

// IsInternal returns true if TDLib type is internal Go type.
func (entity *TdlibType) IsInternal() bool {
	switch entity.name {
	case "double":
		return true

	case "string":
		return true

	case "int32":
		return true

	case "int53":
		return true

	case "int64":
		return true

	case "bytes":
		return true

	case "boolFalse":
		return true

	case "boolTrue":
		return true

	case "vector<t>":
		return true
	}

	return false
}

// GetType returns info about type of TDLib type.
func (entity *TdlibType) GetType() *tlparser.Type {
	name := normalizeEntityName(entity.name)
	for _, typ := range entity.schema.Types {
		if typ.Name == name {
			return typ
		}
	}
	return nil
}

// ToGoType returns string representation of TDLib type in Go style.
func (entity *TdlibType) ToGoType() string {
	if strings.HasPrefix(entity.name, "vector<") {
		log.Fatal("vectors are not supported")
	}

	switch entity.name {
	case "double":
		return "float64"

	case "string":
		return "string"

	case "int32":
		return "int32"

	case "int53":
		return "int64"

	case "int64":
		return "Int64JSON"

	case "bytes":
		return "[]byte"

	case "boolFalse":
		return "bool"

	case "boolTrue":
		return "bool"
	}

	return firstUpper(entity.name)
}

// ToType returns string representation of TDLib type in Go style.
func (entity *TdlibType) ToType() string {
	return entity.ToGoType() + "Type"
}

// HasClass returns true if TDLib type has class.
func (entity *TdlibType) HasClass() bool {
	className := entity.GetType().Class
	for _, class := range entity.schema.Classes {
		if class.Name == className {
			return true
		}
	}

	return false
}

// GetClass returns info about class of TDLib type.
func (entity *TdlibType) GetClass() *tlparser.Class {
	className := entity.GetType().Class
	for _, class := range entity.schema.Classes {
		if class.Name == className {
			return class
		}
	}

	return nil
}

// HasClassProperties returns true if TDLib type has class properties.
func (entity *TdlibType) HasClassProperties() bool {
	for _, prop := range entity.GetType().Properties {
		TdlibTypeProperty := NewTdlibTypeProperty(prop.Name, prop.Type, entity.schema)
		if TdlibTypeProperty.IsClass() && !TdlibTypeProperty.IsList() {
			return true
		}

	}

	return false
}

// IsList returns true if TDLib type is list.
func (entity *TdlibType) IsList() bool {
	return strings.HasPrefix(entity.name, "vector<")
}

// ToClassConst returns string representation of TDLib type class in Go style.
func (entity *TdlibType) ToClassConst() string {
	if entity.HasClass() {
		return "Class" + NewTdlibClass(entity.GetType().Class, entity.schema).ToGoType()
	}
	return "Class" + entity.ToGoType()
}

// ToTypeConst returns string representation of TDLib type class in Go style.
func (entity *TdlibType) ToTypeConst() string {
	return "Type" + entity.ToGoType()
}

// TdlibClass contains TDLib class info.
type TdlibClass struct {
	name   string
	schema *tlparser.Schema
}

// NewTdlibClass returns TDLib class info.
func NewTdlibClass(name string, schema *tlparser.Schema) *TdlibClass {
	return &TdlibClass{
		name:   name,
		schema: schema,
	}
}

// ToGoType returns string representation of TDLib class name in Go style.
func (entity *TdlibClass) ToGoType() string {
	return firstUpper(entity.name)
}

// ToType returns string representation of TDLib class name in Go style.
func (entity *TdlibClass) ToType() string {
	return entity.ToGoType() + "Type"
}

// GetSubTypes returns list of subtype of TDLib class.
func (entity *TdlibClass) GetSubTypes() []*TdlibType {
	types := []*TdlibType{}

	for _, t := range entity.schema.Types {
		if t.Class == entity.name {
			types = append(types, NewTdlibType(t.Name, entity.schema))
		}
	}

	return types
}

// ToClassConst returns string representation of TDLib class in Go style.
func (entity *TdlibClass) ToClassConst() string {
	return "Class" + entity.ToGoType()
}

// TdlibTypeProperty contains TDLib type property info.
type TdlibTypeProperty struct {
	name         string
	propertyType string
	schema       *tlparser.Schema
}

// NewTdlibTypeProperty returns TDLib type property info.
func NewTdlibTypeProperty(name string, propertyType string, schema *tlparser.Schema) *TdlibTypeProperty {
	return &TdlibTypeProperty{
		name:         name,
		propertyType: propertyType,
		schema:       schema,
	}
}

// IsList returns true if TDLib type property is list.
func (entity *TdlibTypeProperty) IsList() bool {
	return strings.HasPrefix(entity.propertyType, "vector<")
}

// GetPrimitive returns info about privitive type of TDLib type property.
func (entity *TdlibTypeProperty) GetPrimitive() string {
	primitive := entity.propertyType

	for strings.HasPrefix(primitive, "vector<") {
		primitive = strings.TrimSuffix(strings.TrimPrefix(primitive, "vector<"), ">")
	}

	return primitive
}

// IsType returns true if TDLib type property is type.
func (entity *TdlibTypeProperty) IsType() bool {
	primitive := entity.GetPrimitive()
	return isType(primitive, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

// GetType returns info about type of TDLib type property.
func (entity *TdlibTypeProperty) GetType() *TdlibType {
	primitive := entity.GetPrimitive()
	return getType(primitive, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

// IsClass returns true if TDLib type property is class.
func (entity *TdlibTypeProperty) IsClass() bool {
	primitive := entity.GetPrimitive()
	return isClass(primitive, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

// GetClass returns info about class of TDLib type property.
func (entity *TdlibTypeProperty) GetClass() *TdlibClass {
	primitive := entity.GetPrimitive()
	return getClass(primitive, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

// ToGoName returns string representation of TDLib type property name in Go style.
func (entity *TdlibTypeProperty) ToGoName() string {
	return firstUpper(underscoreToCamelCase(entity.name))
}

// ToGoFunctionPropertyName returns string representation of TDLib type property name in Go style.
func (entity *TdlibTypeProperty) ToGoFunctionPropertyName() string {
	name := firstLower(underscoreToCamelCase(entity.name))
	if name == "type" {
		name += "Param"
	}

	return name
}

// ToGoType returns string representation of TDLib type property type in Go style.
func (entity *TdlibTypeProperty) ToGoType() string {
	TdlibType := entity.propertyType
	goType := ""

	for strings.HasPrefix(TdlibType, "vector<") {
		goType = goType + "[]"
		TdlibType = strings.TrimSuffix(strings.TrimPrefix(TdlibType, "vector<"), ">")
	}

	if entity.IsClass() {
		return goType + entity.GetClass().ToGoType()
	}

	if entity.GetType().IsInternal() {
		return goType + entity.GetType().ToGoType()
	}

	return goType + "*" + entity.GetType().ToGoType()
}

func isType(name string, field func(entity *tlparser.Type) string, schema *tlparser.Schema) bool {
	name = normalizeEntityName(name)
	for _, entity := range schema.Types {
		if name == field(entity) {
			return true
		}
	}

	return false
}

func getType(name string, field func(entity *tlparser.Type) string, schema *tlparser.Schema) *TdlibType {
	name = normalizeEntityName(name)
	for _, entity := range schema.Types {
		if name == field(entity) {
			return NewTdlibType(entity.Name, schema)
		}
	}

	return nil
}

func isClass(name string, field func(entity *tlparser.Class) string, schema *tlparser.Schema) bool {
	name = normalizeEntityName(name)
	for _, entity := range schema.Classes {
		if name == field(entity) {
			return true
		}
	}

	return false
}

func getClass(name string, field func(entity *tlparser.Class) string, schema *tlparser.Schema) *TdlibClass {
	name = normalizeEntityName(name)
	for _, entity := range schema.Classes {
		if name == field(entity) {
			return NewTdlibClass(entity.Name, schema)
		}
	}

	return nil
}

func normalizeEntityName(name string) string {
	if name == "Bool" {
		name = "boolFalse"
	}

	return tlparser.FixAAA(name)
}
