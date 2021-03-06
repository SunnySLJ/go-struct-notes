package main

const (
	modelHeader = `// Code generated by gormer. DO NOT EDIT.
package %s

import "time"
`
)

var modelTmpl = `
// Table Level Info
const {{.StructName}}TableName = "{{.TableName}}"

// Field Level Info
type {{.StructName}}Field string
const (
{{range $item := .Columns}}
    {{$.StructName}}Field{{$item.FieldName}} {{$.StructName}}Field = "{{$item.GormName}}" {{end}}
)

var {{$.StructName}}FieldAll = []{{$.StructName}}Field{ {{range $k,$item := .Columns}}"{{$item.GormName}}", {{end}}}

// Kernel struct for table for one row
type {{.StructName}} struct { {{range $item := .Columns}}
	{{$item.FieldName}}	{{$item.FieldType}}	` + "`" + `gorm:"column:{{$item.GormName}}"` + "`" + ` // {{$item.Comment}} {{end}}
}

// Kernel struct for table operation
type {{.StructName}}Options struct {
    {{.StructName}} *{{.StructName}}
    Fields []string
}

// Match: case insensitive
var {{$.TableName}}FieldMap = map[string]string{
{{range $item := .Columns}}"{{$item.FieldName}}":"{{$item.GormName}}","{{$item.GormName}}":"{{$item.GormName}}",
{{end}}
}

func New{{.StructName}}Options(target *{{.StructName}}, fields ...{{$.StructName}}Field) *{{.StructName}}Options{
    options := &{{.StructName}}Options{
        {{.StructName}}: target,
        Fields: make([]string, len(fields)),
    }
    for index, field := range fields {
        options.Fields[index] = string(field)
    }
    return options
}

func New{{.StructName}}OptionsAll(target *{{.StructName}}) *{{.StructName}}Options{
    return New{{.StructName}}Options(target, {{$.StructName}}FieldAll...)
}

func New{{.StructName}}OptionsRawString(target *{{.StructName}}, fields ...string) *{{.StructName}}Options{
    options := &{{.StructName}}Options{
        {{.StructName}}: target,
    }
    for _, field := range fields {
        if f,ok := {{$.TableName}}FieldMap[field];ok {
             options.Fields = append(options.Fields, f)
        }
    }
    return options
}
`
