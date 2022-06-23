/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pgtmpl

import (
    "strings"
    "regexp"
    "strconv"
)

const (
    startDelimeter  string  = "<<"
    endDelimeter    string  = ">>"

    startArrayDelimeter  string  = "##"
    endArrayDelimeter    string  = "##"

)

type Template struct {
    templ   string
    vars    map[string]string
    varArrays  map[string][]string
}

func NewTemplate(templ string) *Template {
    vars := make(map[string]string)
    varArrays := make(map[string][]string)
    return &Template{
        templ:      templ,
        vars:       vars,
        varArrays:    varArrays,
    }
}

func (this *Template) SetStrRepl(key string, value string) {
    this.vars[key] = value
}

func (this *Template) SetIntRepl(key string, value int) {
    this.vars[key] = strconv.Itoa(value)
}

func (this *Template) SetInt64Repl(key string, value int64) {
    this.vars[key] = strconv.FormatInt(value, 10)
}

func (this *Template) SetBoolRepl(key string, value bool) {
    this.vars[key] = strconv.FormatBool(value)
}

func (this *Template) SetStrArrayRepl(key string, value []string) {
    this.varArrays[key] = value
}

func (this *Template) Pack() string {
    result := []byte(this.templ)
    result = []byte(strings.Replace(string(result), "\n", " ", -1))
    result = []byte(strings.Replace(string(result), "\t", " ", -1))

    reg := regexp.MustCompile(`\s+`)
    result = reg.ReplaceAll([]byte(result), []byte(" "))

    for key, value := range this.vars {
        reg := regexp.MustCompile(startDelimeter + key + endDelimeter)
        result = reg.ReplaceAll([]byte(result), []byte(value))
    }

    for key, _ := range this.varArrays {
        var value string = "[ "
        for i := range this.varArrays[key] {
            value += `"` + this.varArrays[key][i] + `", ` 
        }
        value += "]"
        value = strings.Replace(value, `, ]`, ` ]`, -1)
        reg := regexp.MustCompile(startArrayDelimeter + key + endArrayDelimeter)
        result = reg.ReplaceAll([]byte(result), []byte(value))
    }

    return string(result)
}
//EOF

