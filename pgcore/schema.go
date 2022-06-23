/*
 * Copyright:  Pixel Networks <support@pixel-networks.com> 
 */

package pgcore

import (
    "encoding/json"
    "errors"
    "strings"
   
    "app/pmtools"
    "app/pgschema"
    "app/pgerrors"
    "app/pgtmpl"
    "app/pmlog"
)

//
// CheckSchemaExists
//
type checkSchemaExistsRespone struct {
    Data struct {
        Schema struct {
            Id pgschema.UUID `json:"id"`
        } `json:"schema"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) CheckSchemaExists(id string) (bool, error) {
    var err error
    var result bool

    gqReq := `{
        "variables": {
            "id": "<<id>>"
        },
        "query": "query CheckSchemaExists($id: UUID!) {
                    schema(id: $id) {
                        id
                    }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("id", id)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp checkSchemaExistsRespone
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }
    if gqResp.Errors != nil {
        err = errors.New("check schema exists: " + gqResp.Errors.GetMessages())
        return result, err
    }

    returnedId := gqResp.Data.Schema.Id
    if id == returnedId {
        result = true
        return result, err
    }
    return result, err
}

//
// DeleteSchema()
//
type DeleteSchemaResponse struct {
    Data struct {
        DeleteSchema struct {
            ClientMutationId pgschema.UUID `json:"clientMutationId"`
        } `json:"deleteSchema"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) DeleteSchema(uuid pgschema.UUID) (string, error) {
    var err error
    var result string

    clientMutationId := pmtools.GetNewUUID()
    
    gqReq := `{
        "variables": {
            "uuid":         "<<uuid>>",
            "mutationId":   "<<mutationId>>"
        },
        "query": "mutation DeleteSchema($uuid: UUID!, $mutationId: String) {
                      deleteSchema(input: {id: $uuid, clientMutationId: $mutationId}) {
                            clientMutationId
                      }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("uuid", uuid)
    tmpl.SetStrRepl("mutationId", clientMutationId)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp DeleteSchemaResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }

    if gqResp.Errors != nil {
        err = errors.New("delete schema: " + gqResp.Errors.GetMessages())
        return result, err
    }
    result = gqResp.Data.DeleteSchema.ClientMutationId
    if clientMutationId != result {
        err = errors.New("delete schema: wrong returned client mutation id")
        return result, err
    }
    return result, err
}


//
// ImportShhema
//
type ImportSchemaResponse struct {
    Data struct {
        ImportSchema struct {
            UUID pgschema.UUID `json:"uuid"`
        } `json:"importSchema"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors,omitempty"`
}

func (this *Pixcore) ImportSchema(schema string) (pgschema.UUID, error) {
    var err     error
    var result  string
    var tmpl    *pgtmpl.Template

    tmpl = pgtmpl.NewTemplate(schema)
    schema = tmpl.Pack()
    // escape schema for json injection
    //schema = strings.Replace(schema, `"`, `\"`, -1)

    gqReq := `{
        "variables": {
            "jsonSchema": <<schema>>
        },
        "query": "mutation ImportSchema($jsonSchema: JSON) {
                    importSchema(input: {jsonSchema: $jsonSchema}) {
                        uuid
                    }
        }"
    }`
    tmpl = pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("schema", schema)
    gqReq = tmpl.Pack()

    pmlog.LogDebug(gqReq)

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    //pmlog.LogDebug(string(httpRespBody))

    var gqResp ImportSchemaResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }

    if gqResp.Errors != nil {
        err = errors.New("import schema: " + gqResp.Errors.GetMessages())
        return result, err
    }

    result = gqResp.Data.ImportSchema.UUID
    return result, err
}

//
// ImportShhema
//
type ExportSchemaResponse struct {
    Data struct {
		ExportSchema string `json:"exportSchema"`
	} `json:"data"`
    Errors  pgerrors.Errors  `json:"errors,omitempty"`
}

func (this *Pixcore) ExportSchema(schemaId pgschema.UUID) (string, error) {
    var err     error
    var result  string
    var tmpl    *pgtmpl.Template

    gqReq := `{
        "variables": {
            "schemaId": "<<schemaId>>"
        },
        "query": "query ExportSchema($schemaId: UUID) {
                    exportSchema(schemaId: $schemaId)
        }"
    }`
    tmpl = pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("schemaId", schemaId)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    pmlog.LogDebug(string(httpRespBody))

    var gqResp ExportSchemaResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }

    if gqResp.Errors != nil {
        err = errors.New("export schema: " + gqResp.Errors.GetMessages())
        return result, err
    }

    result = gqResp.Data.ExportSchema
    result = strings.Replace(result, `"\`, `"`, -1)
   
    return result, err
}
//
// ListSchemas
//
type Schema struct {
    Id          string      `json:"id"`
    Name        string      `json:"name"`
    Type        string      `json:"type"`
    Enabled     bool        `json:"enabled"`
    MTags       []string    `json:"mTags"`
    Owner       string      `json:"applicationOwner"`
    MVersion    string      `json:"mVersion"`
}
func (this *Schema) ToJson() string {
    json, _ := json.Marshal(this)
    return string(json)
}

type ListSchemasResponse struct {
    Data struct {
        Schemas   []Schema  `json:"schemata"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) ListSchemas() ([]Schema, error) {
    var err error
    var result []Schema = make([]Schema, 0)

    gqReq := `{"query": "query {
                     schemata {
                        id
                        name
                        type
                        enabled
                        mTags
                        mVersion
                        applicationOwner
                    }
            }"
    }`

    tmpl := pgtmpl.NewTemplate(gqReq)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp ListSchemasResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }
    if gqResp.Errors != nil {
        err = errors.New("list schemas: " + gqResp.Errors.GetMessages())
        return result, err
    }
    result = gqResp.Data.Schemas
    return result, err
}

//
// ListSchemaProperties
//
type ListSchemaPropertiesResponse struct {
    Data struct {
        SchemaProperties []SchemaProperty   `json:"schemaProperties"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}

type SchemaProperty struct {
    SchemaId     string      `json:"schemaId"`
    DefaultValue string      `json:"defaultValue"`
    Description  string      `json:"description"`
    GroupName    string      `json:"groupName"`
    Hidden       bool        `json:"hidden"`
    Stealth      bool        `json:"stealth"`
    Type         string      `json:"type"`
    //Units        interface{} `json:"units"`
    //ValueRange   interface{} `json:"valueRange"`
    //ValueSet     interface{} `json:"valueSet"`
    //Regex        interface{} `json:"regex"`
    Id           string      `json:"id"`
    Property     string      `json:"property"`
    //Mask         interface{} `json:"mask"`
    Index        string      `json:"index"`
}

func (this *SchemaProperty) GetJSON() string {
    json, _ := json.Marshal(this)
    return string(json)
}
func (this *Pixcore) ListSchemaProperties(schemaId pgschema.UUID, ) ([]SchemaProperty, error) {
    var err error
    var result []SchemaProperty

    gqReq := `{
        "variables": {
                "schemaId": "<<schemaId>>"
        },
        "query": "query ListSchemaProperties($schemaId: UUID) {
                        schemaProperties(condition: {schemaId: $schemaId}) {
                            schemaId
                            defaultValue
                            description
                            groupName
                            hidden
                            stealth
                            type
                            units
                            valueRange
                            valueSet
                            regex
                            id
                            property
                            mask
                            index
                        }
        }"
    }`

    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("schemaId", schemaId)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp ListSchemaPropertiesResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }

    if len(gqResp.Data.SchemaProperties) > 0 {
        result = gqResp.Data.SchemaProperties
    }

    if gqResp.Errors != nil {
        err = errors.New("get schema property value: " + gqResp.Errors.GetMessages())
        return result, err
    }
    return result, err
}
//EOF
