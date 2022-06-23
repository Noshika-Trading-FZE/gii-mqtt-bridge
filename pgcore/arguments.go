/*
 * Copyright:  Pixel Networks <support@pixel-networks.com> 
 */


package pgcore

import (
    "encoding/json"
    "regexp"
)

const (
    ArgumentsTypeBasic    string = "basic"
    ArgumentsTypePublish  string = "publish"
    ArgumentsTypeTopic    string = "topic"
    ArgumentsTypeHello    string = "hello"
    ArgumentsTypeBool     string = "bool"
)

//*********************************************************************//

type Bytes []byte

func (this *Bytes) UnmarshalJSON(data []byte) error {
    var err error
    matched, _ := regexp.MatchString(`:`, string(data))
    if !matched {
        var result []byte
        err = json.Unmarshal(data, &result)
        if err != nil {
            return err
            
        }
        *this = result
        return err
    }

    var result string
    err = json.Unmarshal(data, &result)
    *this = []byte(result)
    return err
}

func (this *Bytes) MarshalJSON() ([]byte, error) {
    return json.Marshal(*this)
}

//*********************************************************************//
func Escape(data string) string {
    j, _ := json.Marshal(data)
    return string(j)[1:len(string(j))-1]
}

//func UnEscape(data string) string {
//    var result string
//    _  = json.Unmarshal([]byte(data), &result)
//    return `"` + result + `"`
//}



//*********************************************************************//
//
// BasicArguments
//
type BasicArguments struct {
}
func NewBasicArguments() *BasicArguments {
    return &BasicArguments{}
    var arguments BasicArguments 
    return &arguments
}
func UnpackBasicArguments(jsonString string) (*BasicArguments, error) {
    var err error
    var arguments BasicArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *BasicArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return Escape(string(jsonBytes))
}
func (this *BasicArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}
//*********************************************************************//
//
// ForwardArguments
//
type ForwardArguments = BasicArguments
func UnpackForwardArguments(jsonString string) (*BasicArguments, error) {
    var err error
    var arguments BasicArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}
//*********************************************************************//
//
// PublishArguments
//
type PublishArguments struct {
    BasicArguments
    //TopicArguments
    TopicName   string      `json:"topicName"`
    Payload     Bytes      `json:"payload"`
}

func NewPublishArguments() *PublishArguments {
    var arguments PublishArguments
    return &arguments
}
func UnpackPublishArguments(jsonString string) (*PublishArguments, error) {
    var err error
    var arguments PublishArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *PublishArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return Escape(string(jsonBytes))
}
func (this *PublishArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}

//*********************************************************************//
//
// TopicArguments
//
type TopicArguments struct {
    BasicArguments
    TopicName   string      `json:"topicName"`
    Payload     Bytes      `json:"payload"`
}

func NewTopicArguments() *TopicArguments {
    var arguments TopicArguments
    return &arguments
}

func UnpackTopicArguments(jsonString string) (*TopicArguments, error) {
    var err error
    var arguments TopicArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *TopicArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return Escape(string(jsonBytes))
}

func (this *TopicArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}

//*********************************************************************//
//
// BoolArguments
//
type BoolArguments struct {
    BasicArguments
    Enable   string      `json:"enable"`
}

func NewBoolArguments() *BoolArguments {
    var arguments BoolArguments
    return &arguments
}

func UnpackBoolArguments(jsonString string) (*BoolArguments, error) {
    var err error
    var arguments BoolArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *BoolArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return Escape(string(jsonBytes))
}

func (this *BoolArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}

//*********************************************************************//
//
//
//
type BLEMACArguments struct {
    BasicArguments
    //DeviceId    UUID     `json:"bridgeId"`   // From:
    //SchemaId    UUID     `json:"schemaId"`   // SenderSchema:
    BLEMAC      string         `json:"bleMac"`
    DistTh      string         `json:"distTh,omitempty"`
    LoTempTh    string         `json:"loTempTh,omitempty"`
    HiTempTh    string         `json:"hiTempTh,omitempty"`
    LoHumiTh    string         `json:"loHumiTh,omitempty"`
    HiHumiTh    string         `json:"hiHumiTh,omitempty"`
}

func NewBLEMACArguments() *BLEMACArguments {
    var arguments BLEMACArguments
    //arguments.Type = ArgumentsTypeTopic
    return &arguments
}

func UnpackBLEMACArguments(jsonString string) (*BLEMACArguments, error) {
    var err error
    var arguments BLEMACArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *BLEMACArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return Escape(string(jsonBytes))
}

func (this *BLEMACArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}

//EOF
