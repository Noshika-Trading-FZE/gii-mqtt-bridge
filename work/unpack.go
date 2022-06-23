package main

import (
    "encoding/json"
    "encoding/base64"
    "fmt"
)


type Bytes []byte

func (this *Bytes) UnmarshalJSON(data []byte) error {
    var err error 
 
    _, err = base64.StdEncoding.DecodeString(string(data))
    if err != nil {
        var result string
        err := json.Unmarshal(data, &result)
        *this = []byte(result)
        return err
    }

    var result []byte
    err = json.Unmarshal(data, &result)
    *this = []byte(result)
    return err
}

func (this *Bytes) MarshalJSON() ([]byte, error) {
    return json.Marshal(*this)
}


type Arguments struct {
    TopicName   string      `json:"topicName"`
    Payload     Bytes       `json:"payload"`
}



func main() {
    var err error

    data := []byte(`{"topicName":"EEEEEEEEEEEEE","payload":"{ \"xxxxx\": 12345 }"}`)

    var arguments Arguments

    err = json.Unmarshal(data, &arguments)
    if err != nil {
        fmt.Println(err)
        
    }

    fmt.Println(arguments.TopicName)
    fmt.Println(string(arguments.Payload))

    j, _ := json.Marshal(arguments)
    fmt.Println(string(j))

}
