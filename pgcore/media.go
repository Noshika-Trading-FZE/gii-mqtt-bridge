/*
 * Copyright:  Pixel Networks <support@pixel-networks.com>
 */

package pgcore

import (
    "encoding/json"
    "errors"

    "app/pgschema"
    "app/pgerrors"
    "app/pgtmpl"
)

const (
    defaultReadPerission    string  = "ffffffff-ffff-ffff-ffff-ffffffffffff"
    defaultEditPermission   string  = "ffffffff-ffff-ffff-ffff-ffffffffffff"
    defaultUsePermission    string  = "ffffffff-ffff-ffff-ffff-ffffffffffff"
)

type RegisterMediaFileResponse struct {
	Data struct {
		CreateObject struct {
			Object struct {
				Id string `json:"id"`
			} `json:"object"`
		} `json:"createObject"`
	} `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) RegisterMediaFile(id pgschema.UUID, name string, schemaId pgschema.UUID) (string, error) {
    var err error
    var result string

    //exists, err := CheckSchemaExists(schemaId)
    //if err != nil {
    //    return result, err
    //}
    //if !exists {
    //    return result, errors.New("unable register media data because media schema not exists")
    //}

    read    := defaultReadPerission
    use     := defaultUsePermission
    edit    := defaultEditPermission
    
    gqReq := `{
        "variables": {
                "id": "<<id>>",
                "name": "<<name>>",
                "schemaId": "<<schemaId>>",
                "edit": "<<edit>>",
                "read": "<<read>>",
                "use": "<<use>>"
        },
        "query": "mutation RegisterMediaFile($id: UUID, $name: String!, $schemaId: UUID!, $edit: UUID, $read: UUID, $use: UUID) {
                    createObject(input: { object: { id: $id, name: $name, schemaId: $schemaId, enabled: true, editorgroup: $edit, readergroup: $read, usergroup: $use } }) {
                        object {
                            id
                        }
                    }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("id", id)
    tmpl.SetStrRepl("schemaId", schemaId)
    tmpl.SetStrRepl("name", name)

    tmpl.SetStrRepl("read", read)
    tmpl.SetStrRepl("edit", edit)
    tmpl.SetStrRepl("use", use)
 
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp RegisterMediaFileResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }
    if gqResp.Errors != nil {
        err = errors.New("register media file: " + gqResp.Errors.GetMessages())
        return result, err
    }
    result = gqResp.Data.CreateObject.Object.Id

    return result, err
}
//EOF
