package command

import (
	"reflect"
	"regexp"
	"strings"
)

type ProjectScopeTags struct {
	BaseRepo              string
	PrNumber              string
	Project               string
	ProjectPath           string
	TerraformDistribution string
	TerraformVersion      string
	Workspace             string
}

func (s ProjectScopeTags) Loadtags() map[string]string {
	tags := make(map[string]string)

	v := reflect.ValueOf(s)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		tags[toSnakeCase(t.Field(i).Name)] = v.Field(i).String()
	}

	return tags
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
