package main

import (
	"bytes"
	"html/template"
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
	{{$item.FieldName}}	{{$item.FieldType}}	` + "`" + `gorm:"column:{{$item.GormName}}"` + "`" + ` // {{$item.Comment}}{{end}}
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

var (
	daoTmplRepo = `
type {{.StructName}}Repo struct {
	db *gorm.DB
}

func New{{.StructName}}Repo(db *gorm.DB) *{{.StructName}}Repo {
	return &{{.StructName}}Repo{db: db}
}

`
	daoTmplAdd = `func (repo *{{.StructName}}Repo) Add{{.StructName}}({{.StructSmallCamelName}} *model.{{.StructName}}) (err error) {
    {{if ne .FieldCreateTime "" }}
    if {{.StructSmallCamelName}}.{{.FieldCreateTime}}.IsZero() {
		{{.StructSmallCamelName}}.{{.FieldCreateTime}} = time.Now()
	}{{end}}
	{{if ne .FieldUpdateTime "" }}
    if {{.StructSmallCamelName}}.{{.FieldUpdateTime}}.IsZero() {
		{{.StructSmallCamelName}}.{{.FieldUpdateTime}} = time.Now()
	}{{end}}
	err = repo.db.
		Table(model.{{.StructName}}TableName).
		Create({{.StructSmallCamelName}}).
		Error
	return
}

`
	daoTmplQuery = `func (repo *{{.StructName}}Repo) Querys{{.StructName}}(pageNumber, pageSize int, condition *model.{{.StructName}}Options) ({{.StructName}}s []model.{{.StructName}}, err error) {
	db := repo.db
	if condition != nil {
		db = db.Where(condition.{{.StructName}}, condition.Fields)
	}
{{if ne .FieldSoftDeleteKey "" }}
	db = db.Where("{{.TableSoftDeleteKey}} != ?", {{.TableSoftDeleteValue}})
{{ end }}
	err = db.
		Table(model.{{.StructName}}TableName).
		Limit(pageSize).
		Offset((pageNumber - 1) * pageSize).
		Find(&{{.StructName}}s).Error
	return
}

`
	daoTmplCount = `func (repo *{{.StructName}}Repo) Count{{.StructName}}s(condition *model.{{.StructName}}Options) (count int64, err error) {
	db := repo.db
	if condition != nil {
		db = db.Where(condition.{{.StructName}}, condition.Fields)
	}
{{if ne .FieldSoftDeleteKey "" }}
	db = db.Where("{{.TableSoftDeleteKey}} != ?", {{.TableSoftDeleteValue}})
{{ end }}
	err = db.
		Table(model.{{.StructName}}TableName).
		Count(&count).Error
	return
}

`
	daoTmplUpdate = `func (repo *{{.StructName}}Repo) Update{{.StructName}}(updated, condition *model.{{.StructName}}Options) (err error) {
	if updated == nil || len(updated.Fields) == 0 {
		return errors.New("update must choose certain fields")
	} else if condition == nil {
		return errors.New("update must include where condition")
	}
{{if ne .FieldUpdateTime "" }}
    if updated.{{.StructName}}.{{.FieldUpdateTime}}.IsZero() {
		updated.{{.StructName}}.{{.FieldUpdateTime}} = time.Now()
		updated.Fields = append(updated.Fields, "{{.TableUpdateTime}}")
	}{{end}}
	err = repo.db.
		Table(model.{{.StructName}}TableName).
		Where(condition.{{.StructName}}, condition.Fields).
		Select(updated.Fields).
		Updates(updated.{{.StructName}}).
		Error
	return
}

`
	daoTmplDelete = `func (repo *{{.StructName}}Repo) Delete{{.StructName}}(condition *model.{{.StructName}}Options) (err error) {
	if condition == nil {
		return errors.New("delete must include where condition")
	}

	err = repo.db.
        Table(model.{{.StructName}}TableName).
		Where(condition.{{.StructName}}, condition.Fields).
{{if eq .FieldSoftDeleteKey "" }} Delete(&model.{{.StructName}}{}).
{{ else }}  {{if eq .FieldUpdateTime "" }}
		Select("{{.TableSoftDeleteKey}}").
		Updates(&model.{{.StructName}}{
		{{.FieldSoftDeleteKey}}:{{.TableSoftDeleteValue}},
		}).
		{{ else }}
        Select("{{.TableSoftDeleteKey}}","{{.TableUpdateTime}}").
		Updates(&model.{{.StructName}}{
		{{.FieldSoftDeleteKey}}:{{.TableSoftDeleteValue}},
		{{.FieldUpdateTime}} : time.Now(),
		}).Error{{ end }}{{ end }}
	return
}
`
	daoTmpl = daoTmplRepo + daoTmplAdd + daoTmplQuery + daoTmplCount + daoTmplUpdate + daoTmplDelete
)

func parseToGormerTmpl(structData StructLevel) (string, error) {
	tmpl, err := template.New("t").Parse(modelTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, structData)
	return buf.String(), nil
}

func parseToDaoTmpl(structData StructLevel) (string, error) {
	tmpl, err := template.New("t").Parse(daoTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, structData)
	return buf.String(), nil
}
