/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pgquery


import (
    "encoding/json"
    "regexp"
    "strings"
)

type GQuery struct {
    Variables   map[string]interface{}  `json:"variables,omitempty"`
    Query       string                  `json:"query,omitempty"`
}

func NewGQuery() *GQuery {
    variables := make(map[string]interface{})
    return &GQuery{
        Variables:  variables,
    } 
}

func (this *GQuery) AddStrVar(key string, value string)  {
    this.Variables[key] = value
}

func (this *GQuery) AddIntVar(key string, value int)  {
    this.Variables[key] = value
}

func (this *GQuery) SetQuery(value string)  {
    this.Query = value
}

func (this *GQuery) ToJson() string {
    result, _ := json.Marshal(this)
    return string(result)
}

func (this *GQuery) MarshalJSON() ([]byte, error) {
    var err error
    tmp := *this
    tmp.Query = (strings.Replace(tmp.Query, "\n", " ", -1))
    tmp.Query = (strings.Replace(tmp.Query, "\t", " ", -1))
    reg := regexp.MustCompile(`\s+`)
    tmp.Query = string(reg.ReplaceAll([]byte(tmp.Query), []byte(" ")))
    tmp.Query = strings.TrimSpace(tmp.Query)
    
    result, err := json.Marshal(tmp)
    return result, err
}
//EOF

