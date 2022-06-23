/*
 * Copyright:  Pixel Networks <support@pixel-networks.com>
 */

package pmtopics

import (
    "fmt"
    "testing"
)


func TestTopics(t *testing.T) {

    topics := TopicsFromString("a1,b1,c2")
    topics.Add("d2")
    fmt.Println(topics.GetJSON())

    topics.Clean()
    fmt.Println(topics.GetJSON())

    topics.Add("e3")
    fmt.Println(topics.GetJSON())

    topics = TopicsFromString("a1,b1,c2")
    for i := range topics.Payload {
        fmt.Println(topics.Payload[i])
    }
}
