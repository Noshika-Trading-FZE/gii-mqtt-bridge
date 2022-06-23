/*
 * Copyright:  Pixel Networks <support@pixel-networks.com>
 */

package pgcore

import (
    "encoding/json"
    "errors"

    "app/pmtools"
    "app/pgschema"
    "app/pgerrors"
    "app/pgtmpl"
)

//
// UpdateObjectPropertyByName
//
func (this *Pixcore) UpdateObjectPropertyByName(objectId pgschema.UUID, propertyName string, value string) (string, error) {
    var err error
    var result string

    propertyId, err := this.GetObjectPropertyId(objectId, propertyName)
    if err != nil {
        return result, err
    }

    if len(propertyId) < uuidStringLen {
        err = errors.New(`update object property by name: property "`+ propertyName +`" not found or have wrong len`)
        return result, err
    }

    mutId, err := this.UpdateObjectProperty(propertyId, value)
    if err != nil {
        return result, err
    }
    result = mutId
    return result, err
}

//
// GetObjectProperty
//
type GetObjectPropertyResult struct {
    Data struct {
        ObjectProperties []struct {
            Id        string `json:"id"`
            ObjectId  string `json:"objectId"`
            Value     string `json:"value"`
            GroupName string `json:"groupName"`
        } `json:"objectProperties"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) GetObjectPropertyId(objectId pgschema.UUID, propertyName string) (string, error) {
    var err error
    var result string

    gqReq := `{
        "variables": {
                "objectId": "<<objectId>>",
                "property": "<<property>>"
        },
        "query": "query GetObjectPropertyId($objectId: UUID, $property: String) {
                    objectProperties(condition: { objectId: $objectId, property: $property }) {
                        id
                        objectId
                        value
                        groupName
                    }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("objectId", objectId)
    tmpl.SetStrRepl("property", propertyName)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp GetObjectPropertyResult
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }

    if len(gqResp.Data.ObjectProperties) > 0 {
        result = gqResp.Data.ObjectProperties[0].Id
    }
    if gqResp.Errors != nil {
        err = errors.New("update object property: " + gqResp.Errors.GetMessages())
        return result, err
    }
    return result, err
}

//
// UpdateObjectProperty
//
type UpdateObjectPropertyResponse struct {
    Data struct {
        UpdateObjectProperty struct {
            ClientMutationId string `json:"clientMutationId"`
        } `json:"updateObjectProperty"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}

func (this *Pixcore) UpdateObjectProperty(propertyId pgschema.UUID, value string) (string, error) {
    var err error
    var result string

    clientMutationId := pmtools.GetNewUUID()

    gqReq := `{
        "variables": {
            "id":               "<<id>>",
            "value":            "<<value>>",
            "clientMutationId": "<<clientMutationId>>"
        },
        "query": "mutation UpdateObjectProperty($id: UUID!, $value: String, $clientMutationId: String) {
                    updateObjectProperty(input: { id: $id, clientMutationId: $clientMutationId, patch: { value: $value }} ) {
                        clientMutationId
                    }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("id",               propertyId)
    tmpl.SetStrRepl("value",            value)
    tmpl.SetStrRepl("clientMutationId", clientMutationId)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp UpdateObjectPropertyResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }

    result = gqResp.Data.UpdateObjectProperty.ClientMutationId
    if gqResp.Errors != nil {
        err = errors.New("update object property: " + gqResp.Errors.GetMessages())
        return result, err
    }
    if clientMutationId != result {
        err = errors.New("update object property: wrong returned client mutation id")
        return result, err
    }

    return result, err
}

// ListPropertiesByObjectGroup
//
type ListPropertiesByObjectGroupResponse struct {
    Data struct {
        ObjectProperties []ObjectProperty   `json:"objectProperties"`
        //ObjectProperties []struct {
        //  ObjectId  string `json:"objectId"`
        //  Value     string `json:"value"`
        //  Property  string `json:"property"`
        //  GroupName string `json:"groupName"`
        //  Id        string `json:"id"`
        //  Type      string `json:"type"`
        //} `json:"objectProperties"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}

type ObjectProperty struct {
    ObjectId  string `json:"objectId"`
    Value     string `json:"value"`
    Property  string `json:"property"`
    GroupName string `json:"groupName"`
    Id        string `json:"id"`
    Type      string `json:"type"`
}

func (this *ObjectProperty) GetJSON() string {
    json, _ := json.Marshal(this)
    return string(json)
}

func (this *Pixcore) ListPropertiesByObjectGroup(objectId pgschema.UUID, groupName string) ([]ObjectProperty, error) {
    var err error
    var result []ObjectProperty

    gqReq := `{
        "variables": {
                "objectId": "<<objectId>>",
                "groupName": "<<groupName>>"
        },
        "query": "query ListPropertiesByObjectGroup($objectId: UUID, $groupName: String) {
                      objectProperties(condition: {groupName: $groupName, objectId: $objectId}) {
                        objectId
                        value
                        property
                        groupName
                        id
                        type
                      }
        }"
    }`

    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("objectId",   objectId)
    tmpl.SetStrRepl("groupName",  groupName)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp ListPropertiesByObjectGroupResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }

    if len(gqResp.Data.ObjectProperties) > 0 {
        result = gqResp.Data.ObjectProperties
    }

    if gqResp.Errors != nil {
        err = errors.New("get object property value: " + gqResp.Errors.GetMessages())
        return result, err
    }
    return result, err
}

func (this *Pixcore) ListProperties(objectId pgschema.UUID, groupName string) ([]ObjectProperty, error) {
    var err error
    var result []ObjectProperty

    gqReq := `{
        "variables": {
                "objectId": "<<objectId>>",
        },
        "query": "query ListProperties($objectId: UUID, $groupName: String) {
                      objectProperties(condition: { objectId: $objectId }) {
                        objectId
                        value
                        property
                        groupName
                        id
                        type
                      }
        }"
    }`

    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("objectId",   objectId)
    tmpl.SetStrRepl("groupName",  groupName)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp ListPropertiesByObjectGroupResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }

    if len(gqResp.Data.ObjectProperties) > 0 {
        result = gqResp.Data.ObjectProperties
    }

    if gqResp.Errors != nil {
        err = errors.New("get object property value: " + gqResp.Errors.GetMessages())
        return result, err
    }
    return result, err
}
//EOF
