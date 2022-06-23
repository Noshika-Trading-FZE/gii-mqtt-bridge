/*
 * Copyright:  Pixel Networks <support@pixel-networks.com> 
 */

package pgcore

import (
    "encoding/json"
    "errors"
    
    "app/pgerrors"
    "app/pgtmpl"
)



type GetUserProfileIdRespone struct {
	Data struct {
        GetUserProfileId string `json:"getUserProfileId"`
	} `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) GetUserProfileId() (string, error) {
    var err error
    var result string
    gqReq := `{
        "query": "query GetUserProfileId {
                          getUserProfileId
        }"
    }`

    tmpl := pgtmpl.NewTemplate(gqReq)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp GetUserProfileIdRespone
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }
    if gqResp.Errors != nil {
        err = errors.New(gqResp.Errors.GetMessages())
        return result, err
    }

    result = gqResp.Data.GetUserProfileId
    return result, err
}
//
// CheckObjectExists
//
type GetUserIdByLoginRespone struct {
	Data struct {
        GetUserId string `json:"getUserId"`
		//Users []struct {
		//	Id    string `json:"id"`
		//	Login string `json:"login"`
		//} `json:"users"`
	} `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) GetUserIdByLogin(login string) (string, error) {
    var err error
    var result string
    //gqReq := `{
    //    "variables": {
    //        "login": "<<login>>"
    //    },
    //    "query": "query MyQuery($login: String) {
    //                      users(condition: {login: $login}) {
    //                        id
    //                        login
    //                      }
    //    }"
    //}`
    gqReq := `{
        "query": "query GetUserIdByLogin {
                          getUserId
        }"
    }`

    tmpl := pgtmpl.NewTemplate(gqReq)
    //tmpl.SetStrRepl("login", login)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp GetUserIdByLoginRespone
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }
    if gqResp.Errors != nil {
        err = errors.New("check schema exists: " + gqResp.Errors.GetMessages())
        return result, err
    }

    result = gqResp.Data.GetUserId
    return result, err
}
//EOF
