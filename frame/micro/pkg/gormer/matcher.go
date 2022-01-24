package main

var tableInfo []TableInfo

type TableInfo struct {
	Name            string `yaml:"name"`
	GoStruct        string `yaml:"goStruct"`
	CreateTime      string `yaml:"createTime"`
	UpdateTime      string `yaml:"updateTime"`
	SoftDeleteKey   string `yaml:"softDeleteKey"`
	SoftDeleteValue int    `yaml:"softDeleteValue"`
}

func getTableMatcher() map[string]TableInfo {
	var tMatcher = make(map[string]TableInfo)
	for _, matcher := range tableInfo {
		tMatcher[matcher.Name] = matcher
	}
	return tMatcher
}

//func getTableMatcher(tableMatcher string, tables []string) map[string]string {
//	var tMatcher = make(map[string]string)
//	matchers := strings.Split(tableMatcher, ",")
//	for _, matcher := range matchers {
//		m := strings.Split(matcher, ":")
//		if len(m) == 2 {
//			tMatcher[m[0]] = m[1]
//		}
//	}
//	for _, table := range tables {
//		if _, ok := tMatcher[table]; ok {
//			continue
//		}
//		tMatcher[table] = table
//	}
//	return tMatcher
//}
