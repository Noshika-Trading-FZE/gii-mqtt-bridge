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
// CreateObject
//
type CreateObjectResponse struct {
    Data struct {
        CreateObject struct {
            Object struct {
                Id string `json:"id"`
            } `json:"object"`
        } `json:"createObject"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}

func (this *Pixcore) CreateObject(object *pgschema.Object) (pgschema.UUID, error) {
    var err error
    var result string

    gqReq := `{
        "variables": {
                "id":           "<<id>>",
                "name":         "<<name>>",
                "enabled":       <<enabled>>,
                "description":  "<<description>>",
                "schemaId":     "<<schemaId>>"
        },
        "query": "mutation CreateObject(  $id:            UUID,
                                        $name:          String!,
                                        $description:   String!,
                                        $schemaId:      UUID!,
                                        $enabled:       Boolean
                                        ) {
                    createObject(input: {object: {id:               $id,
                                                    name:           $name,
                                                    enabled:        $enabled,
                                                    description:    $description,
                                                    schemaId:       $schemaId }}) {
                        object {
                            id
                        }
                    }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("id",           string(object.Id))
    tmpl.SetStrRepl("name",         object.Name)
    tmpl.SetBoolRepl("enabled",     object.Enabled)
    tmpl.SetStrRepl("description",  object.Description)
    tmpl.SetStrRepl("schemaId",     string(object.SchemaId))
    gqReq = tmpl.Pack()


    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp CreateObjectResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }
    if gqResp.Errors != nil {
        err = errors.New("create object: " + gqResp.Errors.GetMessages())
        return result, err
    }
    result = gqResp.Data.CreateObject.Object.Id
    return result, err
}
//
// DeleteObject
//
type DeleteObjectResponse struct {
    Data struct {
        DeleteObject struct {
            ClientMutationId pgschema.UUID `json:"clientMutationId"`
        } `json:"deleteObject"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) DeleteObject(uuid pgschema.UUID) error {
    var err error
    //var result string

    clientMutationId := pmtools.GetNewUUID()

    gqReq := `{
        "variables": {
            "uuid": "<<uuid>>",
            "mutationId": "<<clientMutationId>>"
        },
        "query": "mutation DeleteObject($uuid: UUID!, $mutationId: String) {
                      deleteObject(input: {id: $uuid, clientMutationId: $mutationId}) {
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
        return err
    }

    var gqResp DeleteObjectResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return err
    }

    if gqResp.Errors != nil {
        err = errors.New("delete object: " + gqResp.Errors.GetMessages())
        return err
    }
    //result = gqResp.Data.DeleteObject.ClientMutationId
    //if clientMutationId != result {
    //    err = errors.New("delete object: wrong returned client mutation id")
    //    return err
    //}
    return err
}


func (this *Pixcore) DisableObject(id pgschema.UUID) error {
    var err error
    //var result string

    clientMutationId := pmtools.GetNewUUID()

    gqReq := `{
        "variables": {
            "id": "<<id>>",
            "clientMutationId": "<<clientMutationId>>"
        },
        "query": "mutation DisableObject($id: UUID!, $clientMutationId: String) {
                    updateObject( input: {patch: { enabled: false }, id: $id, clientMutationId: $clientMutationId} ) {
                            clientMutationId
                      }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("id", id)
    tmpl.SetStrRepl("clientMutationId", clientMutationId)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return err
    }

    var gqResp DeleteObjectResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return err
    }

    if gqResp.Errors != nil {
        err = errors.New("disable object: " + gqResp.Errors.GetMessages())
        return err
    }
    //result = gqResp.Data.DeleteObject.ClientMutationId
    //if clientMutationId != result {
    //    err = errors.New("disable object: wrong returned client mutation id")
    //    return err
    //}
    return err
}

func (this *Pixcore) EnableObject(id pgschema.UUID) error {
    var err error
    //var result string

    clientMutationId := pmtools.GetNewUUID()

    gqReq := `{
        "variables": {
            "id": "<<id>>",
            "clientMutationId": "<<clientMutationId>>"
        },
        "query": "mutation EnableObject($id: UUID!, $clientMutationId: String) {
                    updateObject( input: {patch: { enabled: true }, id: $id, clientMutationId: $clientMutationId} ) {
                            clientMutationId
                      }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("id", id)
    tmpl.SetStrRepl("clientMutationId", clientMutationId)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return err
    }

    var gqResp DeleteObjectResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return err
    }

    if gqResp.Errors != nil {
        err = errors.New("disable object: " + gqResp.Errors.GetMessages())
        return err
    }
    //result = gqResp.Data.DeleteObject.ClientMutationId
    //if clientMutationId != result {
    //    err = errors.New("disable object: wrong returned client mutation id")
    //    return err
    //}
    return err
}


//
// CheckObjectExists
//
type checkObjectExistsRespone struct {
    Data struct {
        Object struct {
            Id string `json:"id"`
        } `json:"object"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) CheckObjectExists(id pgschema.UUID) (bool, error) {
    var err error
    var result bool

    gqReq := `{
        "variables": {
            "id": "<<id>>"
        },
        "query": "query CheckObjectExists($id: UUID!) {
                object(id: $id) {
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
    var gqResp checkObjectExistsRespone
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }
    if gqResp.Errors != nil {
        err = errors.New("check schema exists: " + gqResp.Errors.GetMessages())
        return result, err
    }

    returnedId := gqResp.Data.Object.Id
    if id == returnedId {
        result = true
        return result, err
    }
    return result, err
}


type ListObjectsResponse struct {
    Data struct {
        Objects []pgschema.Object    `json:"objects"`
        //Objects []struct {
        //  Id       string `json:"id"`
        //  Name     string `json:"name"`
        //  SchemaId string `json:"schemaId"`
        //} `json:"objects"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}

//
// ListObjects()
//
func (this *Pixcore) ListObjects() ([]pgschema.Object, error) {
    var err error
    var result []pgschema.Object = make([]pgschema.Object, 0)

    gqReq := `{
        "query": "query {
             objects {
                id
                name
                schemaId
            }
        }"
    }`

    tmpl := pgtmpl.NewTemplate(gqReq)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp ListObjectsResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }

    if gqResp.Errors != nil {
        err = errors.New("list objects by schema id: " + gqResp.Errors.GetMessages())
        return result, err
    }

    result = gqResp.Data.Objects
    return result, err
}

//
// ListObjectsBySchemaId()
//
func (this *Pixcore) ListObjectsBySchemaId(schemaId pgschema.UUID) ([]pgschema.Object, error) {
    var err error
    var result []pgschema.Object = make([]pgschema.Object, 0)

    gqReq := `{
        "variables": {
                "schemaId": "<<schemaId>>"
        },
        "query": "query ListObjectsBySchemaId($schemaId: UUID) {
                     objects (condition: { schemaId: $schemaId }) {
                        id
                        name
                        schemaId
                        editorgroup
                        usergroup
                        readergroup
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

    var gqResp ListObjectsResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }
    if gqResp.Errors != nil {
        err = errors.New("list objects by schema id: " + gqResp.Errors.GetMessages())
        return result, err
    }
    result = gqResp.Data.Objects
    return result, err
}


//type GetObjectPropertyResult struct {
//    Data struct {
//        ObjectProperties []struct {
//            Id        string `json:"id"`
//            ObjectId  string `json:"objectId"`
//            Value     string `json:"value"`
//            GroupName string `json:"groupName"`
//        } `json:"objectProperties"`
//    } `json:"data"`
//    Errors  pgerrors.Errors  `json:"errors"`
//}

//
// GetObjectPropertyValue
//
func (this *Pixcore) GetObjectPropertyValue(objectId pgschema.UUID, propertyName string) (string, error) {
    var err error
    var result string

    gqReq := `{
        "variables": {
                "objectId": "<<objectId>>",
                "property": "<<propertyName>>"
        },
        "query": "query GetObjectPropertyValue($objectId: UUID, $property: String) {
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
    tmpl.SetStrRepl("propertyName", propertyName)
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
        result = gqResp.Data.ObjectProperties[0].Value
    }

    if gqResp.Errors != nil {
        err = errors.New("get object property value: " + gqResp.Errors.GetMessages())
        return result, err
    }
    return result, err
}

//
// UpdateObjectProperty
//
//type UpdateObjectPropertyResponse struct {
//    Data struct {
//        UpdateObjectProperty struct {
//            ClientMutationId string `json:"clientMutationId"`
//        } `json:"updateObjectProperty"`
//    } `json:"data"`
//    Errors  pgerrors.Errors  `json:"errors"`
//}

//func (this *Pixcore) UpdateObjectProperty(propertyId pgschema.UUID, value string) (string, error) {
//    var err error
//    var result string

//    clientMutationId := pmtools.GetNewUUID()

//    gqReq := `{
//        "variables": {
//            "id":               "<<propertyId>>",
//            "value":            "<<value>>",
//            "clientMutationId": "<<clientMutationId>>"
//        },
//        "query": "mutation myMutation($id: UUID!, $value: String, $clientMutationId: String) {
//                    updateObjectProperty(input: { id: $id, clientMutationId: $clientMutationId, patch: { value: $value }} ) {
//                        clientMutationId
//                    }
//        }"
//    }`

//    tmpl := pgtmpl.NewTemplate(gqReq)
//    tmpl.SetStrRepl("propertyId", propertyId)
//    tmpl.SetStrRepl("value", value)
//    tmpl.SetStrRepl("clientMutationId", clientMutationId)
//    gqReq = tmpl.Pack()
//
//    httpRespBody, err := this.httpRequest(gqReq)
//    if err != nil {
//        return result, err
//    }
//
//    var gqResp UpdateObjectPropertyResponse
//    err = json.Unmarshal(httpRespBody, &gqResp)
//    if err != nil {
//        return result, err
//    }
//
//    result = gqResp.Data.UpdateObjectProperty.ClientMutationId
//    if gqResp.Errors != nil {
//        err = errors.New("update object property: " + gqResp.Errors.GetMessages())
//        return result, err
//    }
//    if clientMutationId != result {
//        err = errors.New("update object property: wrong returned client mutation id")
//        return result, err
//    }
//
//    return result, err
//}

//
// UpdateObjectPropertyByName
//

// For future, one mutation:
//mutation updateObjectProperty {
//  updateObjectProperty(
//    input: {patch: {objectId: "", property: "", value: ""}, id: "", clientMutationId: ""}
//  )
//}

//func (this *Pixcore) UpdateObjectPropertyByName(objectId pgschema.UUID, propertyName string, value string) (string, error) {
//    var err error
//    var result string

//    propertyId, err := this.GetObjectPropertyUUID(objectId, propertyName)
//    if err != nil {
//        return result, err
//    }

//    if len(propertyId) < uuidStringLen {
//        err = errors.New(`update object property by name: property "`+ propertyName +`" not found or have wrong len`)
//        return result, err
//    }

//    mutId, err := this.UpdateObjectProperty(propertyId, value)
//    if err != nil {
//        return result, err
//    }
//    result = mutId
//    return result, err
//}
//EOF
