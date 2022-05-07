package tplfunctions

import (
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// https://github.com/Masterminds/sprig/blob/master/strings.go
func strval(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// https://github.com/Masterminds/sprig/blob/master/strings.go
func strslice(v interface{}) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []interface{}:
		b := make([]string, 0, len(v))
		for _, s := range v {
			if s != nil {
				b = append(b, strval(s))
			}
		}
		return b
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			l := val.Len()
			b := make([]string, 0, l)
			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				if value != nil {
					b = append(b, strval(value))
				}
			}
			return b
		default:
			if v == nil {
				return []string{}
			}

			return []string{strval(v)}
		}
	}
}

func joinSingle(sep string, v interface{}) (result string) {
	strs := strslice(v)

	if len(strs) < 1 {
		return
	}
	result = strs[0]
	for i := 1; i < len(strs); i++ {
		result = strings.TrimRight(result, "/") + "/" + strings.TrimLeft(strs[i], "/")
	}
	return
}

func iterate(count int) []int {
	var i int
	var Items []int
	for i = 0; i < count; i++ {
		Items = append(Items, i)
	}
	return Items
}

var mediaserverRegexp = regexp.MustCompile("^mediaserver:([^/]+)/([^/]+)/(.+)$")

func MediaUrl(mediaUri, ext string) string {
	matches := mediaserverRegexp.FindStringSubmatch(mediaUri)
	if matches == nil {
		return ""
	}
	collection := matches[1]
	signature := matches[2]
	function := matches[3]

	functions := strings.Split(strings.ToLower(function), "/")
	cmd := functions[0]
	functions = functions[1:]
	sort.Strings(functions)
	functions = append([]string{cmd}, functions...)
	function = strings.Join(functions, "/")
	filename := strings.ToLower(fmt.Sprintf("%s_%s_%s.%s",
		collection,
		strings.ReplaceAll(signature, "$", "-"),
		strings.ReplaceAll(function, "/", "_"),
		strings.TrimPrefix(ext, ".")))
	if len(filename) > 203 {
		filename = fmt.Sprintf("%s-_-%s", filename[:100], filename[len(filename)-100:])
	}

	return filename
}

var funcMap = map[string]any{
	"joinSingle": joinSingle,
	"iterate":    iterate,
	"mediaUrl":   MediaUrl,
	"correctWeb": func(u string) string {
		if strings.HasPrefix(strings.ToLower(u), "http") {
			return u
		}
		return "https://" + u
	},
}

func GetFuncMap() template.FuncMap {
	return template.FuncMap(funcMap)
}
