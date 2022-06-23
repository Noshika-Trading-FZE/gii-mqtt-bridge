/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pgschema

import (
    "encoding/json"
)

const (
    BroadcastUUID       UUID    = "ffffffff-ffff-ffff-ffff-ffffffffffff"
    TimeUnixEpoch       string  = "1970-01-01T00:00:00Z"
    MetadataTypeDevice  string  = "device"
    MetadataTypeApp     string  = "application"
)

type UUID = string
//
// Schema
//
type Schema struct {
    Metadata    *Metadata       `json:"schema"`
    Properties  []*Property     `json:"properties,omitempty"`
    Controls    []*Control      `json:"controls,omitempty"`
}

func (this *Schema) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}

func NewSchema() *Schema{
    var schema Schema
    schema.Metadata = NewMetadata()
    return &schema
}
//
// Metadata
//
type Metadata struct {
    Id               UUID       `json:"id"`
    Name             string     `json:"name"`
    Type             string     `json:"type"`
    Enabled          bool       `json:"enabled"`

    ApplicationOwner UUID       `json:"application_owner,omitempty"`
    DefaultTTL       string     `json:"default_ttl,omitempty"`
    Description      string     `json:"description"`

    MAuthor          string     `json:"m_author"`
    MEmail           string     `json:"m_email"`
    MExternalId      UUID       `json:"m_external_id"`
    MIcon            string     `json:"m_icon"`
    MLongname        string     `json:"m_longname"`
    MManufacturer    string     `json:"m_manufacturer"`
    MPicture         string     `json:"m_picture"`
    MTags            []string   `json:"m_tags"`
    MVersion         string     `json:"m_version"`

    Editorgroup     UUID        `json:"editorgroup,omitempty"`
    Usergroup       UUID        `json:"usergroup,omitempty"`
    Readergroup     UUID        `json:"readergroup,omitempty"`

}

func NewMetadata() *Metadata{
    var metadata Metadata
    metadata.MTags = make([]string, 0)
    //metadata.DefaultTTL           = "10 years"
    metadata.Enabled              = true
    metadata.MAuthor              = "Pixel"
    metadata.MManufacturer        = "Pixel"
    metadata.MEmail               = "support@pixel-networks.com"
    //metadata.Editorgroup          = BroadcastUUID
    //metadata.Usergroup            = BroadcastUUID
    //metadata.Readergroup          = BroadcastUUID
    return &metadata
}

func (this *Metadata) GetJSON() string {
    jsonBytes, _ := json.MarshalIndent(this, "", "    ")
    return string(jsonBytes)
}
//
// Property
//
type PropertyType = string

const (
    HexType     PropertyType = "hex"
    BoolType    PropertyType = "bool"
    StringType  PropertyType = "string"
    FloatType   PropertyType = "float"
    IntType     PropertyType = "int"
    DoubleType  PropertyType = "double"
)

type Property struct {
    Property         string     `json:"property,omitempty"`
    Description      string     `json:"description,omitempty"`
    Type             string     `json:"type,omitempty"`

    DefaultValue     string     `json:"default_value,omitempty"`
    GroupDescription string     `json:"group_description,omitempty"`
    GroupName        string     `json:"group_name,omitempty"`

    Hidden           bool       `json:"hidden"`
    Index            int        `json:"index"`
    MTags            []string   `json:"m_tags,omitempty"`
    Mask             string     `json:"mask,omitempty"`
    Regex            string     `json:"regex,omitempty"`
    Units            string     `json:"units,omitempty"`
    ValueRange       string     `json:"value_range,omitempty"`
    ValueSet         string     `json:"value_set,omitempty"`
}

func (this *Property) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}

func NewProperty() *Property {
    var property Property
    property.MTags = make([]string, 0)
    return &property
}
//
// Control
//
type Control struct {
		Argument     string     `json:"argument"`
		DefaultValue string     `json:"default_value,omitempty"`
		Description  string     `json:"description,omitempty"`
		Hidden       bool       `json:"hidden"`
		Mask         string     `json:"mask,omitempty"`
		RPC          string     `json:"rpc,omitempty"`
		Regex        string     `json:"regex,omitempty"`
		Type         string     `json:"type"`
		Units        string     `json:"units,omitempty"`
		ValueRange   string     `json:"value_range,omitempty"`
		ValueSet     string     `json:"value_set,omitempty"`
}
func (this *Control) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}

func NewControl() *Control {
    var control Control
    return &control
}
//EOF
