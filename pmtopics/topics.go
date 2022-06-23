/*
 * Copyright:  Pixel Networks <support@pixel-networks.com>
 */

package pmtopics

import (
    "errors"
    "encoding/json"
    "fmt"
    "strings"
)
//
// Topics
//

const (
    separator string = ","
)

type Topics struct {
    Payload     []string
}

func NewTopics() *Topics {
    var topics Topics
    topics.Payload = make([]string, 0)
    return &topics
}

func TopicsFromString(source string) *Topics {
    list := strings.Split(source, separator)

    var topics Topics
    topics.Payload = make([]string, 0)
    for i := range list {
        topics.Add(list[i])
    }
    return &topics
}

func (this *Topics) Clean() {
    this.Payload = make([]string, 0)
}

func (this *Topics) GetArray() []string {
    return this.Payload 
}

func (this *Topics) Add(topicName string) error {
    var err error

    for i := range this.Payload {
        if this.Payload[i] == topicName {
            message := fmt.Sprintf("topic %s already exists", topicName)
            return errors.New(message)
        }
    } 
    this.Payload = append(this.Payload, topicName)
    return err
}

func (this *Topics) GetJSON() string {
    jBytes, _ := json.Marshal(this.Payload)
    return string(jBytes)
}

func (this Topics) MarshalJSON() ([]byte, error) {
    jBytes, err := json.Marshal(this.Payload)
    return jBytes, err
}
//EOF



