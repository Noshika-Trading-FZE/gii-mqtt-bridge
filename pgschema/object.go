
/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */


package pgschema

import (
    "encoding/json"
)
//
// Object
// 
type Object struct {
    Id          UUID            `json:"id"`
    Name        string          `json:"name"`
    SchemaId    UUID            `json:"schemaId"`
    Enabled     bool            `json:"enabled"`
    Editorgroup UUID            `json:"editorgroup,omitempty"`
    Usergroup   UUID            `json:"usergroup,omitempty"`
    Readergroup UUID            `json:"readergroup,omitempty"`
    Description string          `json:"description"`
}

func NewObject() *Object {
    var object Object
    object.Enabled      = true
    //object.Editorgroup  = BroadcastUUID
    //object.Usergroup    = BroadcastUUID
    //object.Readergroup  = BroadcastUUID
    return &object
}

func (this *Object) ToJson() string {
    json, _ := json.MarshalIndent(this, "", "    ")
    return string(json)
}
//EOF
