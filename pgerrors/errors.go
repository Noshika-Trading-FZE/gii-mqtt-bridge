
package pgerrors

import (
    "encoding/json"
    "strings"
    //"bytes"
    //"errors"
    //"fmt"
    //"io/ioutil"
    //"log"
    //"net/http"
    //"net/url"
    //"strconv"
    //"time"

    //"app/tools"
    //"app/schemas"
    //"app/jwt"
)


type Errors []Error

func (this *Errors) ToJson() string {
    jsonBytes, _ := json.Marshal(this)
    jsonString := strings.Replace(string(jsonBytes), "\\n", "\n", -1)
    return jsonString
}
func (this *Errors) ToJsonIndent() string {
    jsonBytes, _ := json.Marshal(this)
    jsonString := strings.Replace(string(jsonBytes), "\\n", "\n", -1)
    return jsonString
}

func (this *Errors) GetMessages() string {
    var messages string
    for _, gqError := range *this {
        messages = ";" + gqError.Message
    }
    messages = strings.Replace(messages, ";", "", 1)
    messages = strings.Replace(messages, `\\"`, `"`, -1)
    return messages
}


func (this *Errors) GetStack() string {
    var messages string
    for _, gqError := range *this {
        messages = "\n" + gqError.Stack
    }
    messages = strings.Replace(messages, "\n", "", 1)
    //messages = strings.Replace(messages, `\\"`, `"`, -1)
    return messages
}


type Error struct {
    Code        string      `json:"code,omitempty"`
    File        string      `json:"file,omitempty"`
    Line        string      `json:"line,omitempty"`
    Message     string      `json:"message,omitempty"`
    Path        []string    `json:"path,omitempty"`
    Routine     string      `json:"routine,omitempty"`
    Severity    string      `json:"severity,omitempty"`
    Stack       string      `json:"stack,omitempty"`

    Extensions struct {
        Exception struct {
            Severity string `json:"severity,omitempty"`
            Code     string `json:"code,omitempty"`
            File     string `json:"file,omitempty"`
            Line     string `json:"line,omitempty"`
            Routine  string `json:"routine,omitempty"`
        } `json:"exception,omitempty"`
    } `json:"extensions,omitempty"`

    Locations []struct {
        Line   int `json:"line,omitempty"`
        Column int `json:"column,omitempty"`
    } `json:"locations,omitempty"`
}

func (this *Error) ToJson() string {
    json, _ := json.Marshal(this)
    return string(json)
}

func (this *Error) ToJsonIndent() string {
    json, _ := json.MarshalIndent(this, "", "    ")
    return string(json)
}

func (this *Error) GetMessages() string {
    return strings.Replace(this.Message, `\\"`, `"`, -1)
}

