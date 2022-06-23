/*
 * Copyright:  Pixel Networks <support@pixel-networks.com> 
 */

package pgcore

import (
    "encoding/json"
    "errors"
    "time"

    "app/pgschema"
    "app/pgerrors"
    "app/pgtmpl"
    "app/pmtools"
)

const (
    wsTimeout       time.Duration   = 10  // sec
)

type JSON = string

//
// CreateControlExecution
//
type CreateControlExecutionResponse struct {
    Data struct {
        CreateControlExecution struct {
            ControlExecution struct {
                Id     int    `json:"id"`
                Name   string `json:"name"`
                Params string `json:"params"`
            } `json:"controlExecution"`
        } `json:"createControlExecution"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}

func (this *Pixcore) CreateControlExecution(objectId pgschema.UUID, controlName string, params JSON) (int, error) {
    var err error
    var result int

    gqReq := `{
        "variables": {
                "objectId": "<<objectId>>",
                "controlName": "<<controlName>>",
                "params": "<<params>>"
        },
        "query": "mutation CreateControlExecution($objectId: UUID!, $controlName: String!, $params: JSON!) {
                      createControlExecution(input: {controlExecution: { objectId: $objectId, name: $controlName, params: $params }}) {
                            controlExecution {
                                id
                                name
                                params
                            }
                      }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("objectId",     objectId)
    tmpl.SetStrRepl("controlName",  controlName)
    tmpl.SetStrRepl("params",       params)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return result, err
    }

    var gqResp CreateControlExecutionResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return result, err
    }
    result = gqResp.Data.CreateControlExecution.ControlExecution.Id
    if gqResp.Errors != nil {
        err = errors.New("list objects by property name: " + gqResp.Errors.GetMessages())
        return result, err
    }
    return result, err
}

//
// CreateControlExecutionStealth()
//
type CreateControlExecutionStealthResponse struct {
    Data struct {
		CreateControlExecutionStealth struct {
			Boolean          bool   `json:"boolean"`
		} `json:"createControlExecutionStealth"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}

func (this *Pixcore) CreateControlExecutionStealth(objectId pgschema.UUID, controlName string, params JSON) error {
    var err error

    gqReq := `{
        "variables": {
                "objectId": "<<objectId>>",
                "controlName": "<<controlName>>",
                "params": "<<params>>"
        },
        "query": "mutation CreateControlExecutionStealth($objectId: UUID!, $controlName: String!, $params: JSON!) {
                        createControlExecutionStealth(input: { objectId: $objectId, name: $controlName, params: $params }) {
                            boolean
                        }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("objectId",     objectId)
    tmpl.SetStrRepl("controlName",  controlName)
    tmpl.SetStrRepl("params",       params)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return err
    }

    var gqResp CreateControlExecutionStealthResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return err
    }

    if gqResp.Errors != nil {
        err = errors.New("create control execution stealth: " + gqResp.Errors.GetMessages())
        return err
    }
    return err
}

//
// UpdateControlExecutionAck()
//
type UpdateControlExecutionAckResponse struct {
	Data struct {
		UpdateControlExecutionAck struct {
			ClientMutationIв string `json:"clientMutationId"`
		} `json:"updateControlExecutionAck"`
	} `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) UpdateControlExecutionAck(controlsExecutionId int64) error {
    var err error

    if controlsExecutionId < 0 {
        return err
    }

    clientMutationId := pmtools.GetNewUUID()
    
    gqReq := `{
        "variables": {
                "controlsExecutionId": "<<controlsExecutionId>>",
                "clientMutationId": "<<clientMutationId>>"
        },
        "query": "mutation UpdateControlExecutionAck($controlsExecutionId: BigInt!, $clientMutationId: String!) {
                        updateControlExecutionAck(input: {controlsExecutionId: $controlsExecutionId, clientMutationId: $clientMutationId}) {
                            clientMutationId
                        }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetInt64Repl("controlsExecutionId",  controlsExecutionId)
    tmpl.SetStrRepl("clientMutationId",     clientMutationId)
    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return err
    }

    var gqResp UpdateControlExecutionAckResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return err
    }

    if gqResp.Errors != nil {
        err = errors.New("list objects by property name: " + gqResp.Errors.GetMessages())
        return err
    }
    result := gqResp.Data.UpdateControlExecutionAck.ClientMutationIв
    if clientMutationId != result {
        err = errors.New("delete schema: wrong returned client mutation id")
        return err
    }
    return err
}

//
// CreateControlExecutionReport()
//
type CreateControlExecutionReportResponse struct {
	Data struct {
		CreateControlExecutionReport struct {
			ClientMutationIв string `json:"clientMutationId"`
		} `json:"createControlExecutionReport"`
	} `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
func (this *Pixcore) CreateControlExecutionReport(controlId int64, wError bool, done bool, report string) error {
    var err error

    if controlId < 0 {
        return err
    }

    //var report string = ""
    var reportDetails string = `"{ }"`

    clientMutationId := pmtools.GetNewUUID()
    
    gqReq := `{
        "variables": {
                "error": <<error>>,
                "done": <<done>>,
                "clientMutationId": "<<clientMutationId>>",
                "linkedControlId": <<linkedControlId>>,
                "report": "<<report>>",
                "reportDetails": <<reportDetails>>
        },
        "query": "mutation CreateControlExecutionReport($error: Boolean!,
                                    $done: Boolean!,
                                    $clientMutationId: String,
                                    $linkedControlId: Int!,
                                    $report: String!,
                                    $reportDetails: JSON!) {
                        createControlExecutionReport(
                            input: {
                                    linkedControlId: $linkedControlId,
                                    report: $report,
                                    reportDetails: $reportDetails,
                                    done: $done,
                                    error: $error,
                                    clientMutationId: $clientMutationId}
                            ) {
                                clientMutationId
                            }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)

    tmpl.SetBoolRepl("error", wError)
    tmpl.SetBoolRepl("done", done)
    tmpl.SetStrRepl("clientMutationId", clientMutationId)
    tmpl.SetInt64Repl("linkedControlId", controlId)
    tmpl.SetStrRepl("report", report)
    tmpl.SetStrRepl("reportDetails", reportDetails)

    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return err
    }

    var gqResp CreateControlExecutionReportResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return err
    }

    if gqResp.Errors != nil {
        err = errors.New("list objects by property name: " + gqResp.Errors.GetMessages())
        return err
    }
    result := gqResp.Data.CreateControlExecutionReport.ClientMutationIв
    if clientMutationId != result {
        err = errors.New("delete schema: wrong returned client mutation id")
        return err
    }
    return err
}

func (this *Pixcore) CreateControlExecutionEmptyReport(controlId int64, wError bool, done bool) error {
    return this.CreateControlExecutionReport(controlId, wError, done, "")
}

//
// CreateControlExecutionStealthByPropertyPattern()
//

type EnrichWith struct {
    GroupName   string `json:"groupName"`
    Property    string `json:"property"` 
}

func (this *Pixcore) CreateControlExecutionStealthByPropertyPattern(controlName string, params JSON, groupName string, property string, pattern string) error {
    var err error

    gqReq := `{
        "variables": {
                "controlName": "<<controlName>>",
                "params": "<<params>>",
                "groupName": "<<groupName>>",
                "property": "<<property>>",
                "pattern": "<<pattern>>"
        },
        "query": "mutation CreateControlExecutionStealthByPropertyPattern( $controlName: String!, $params: JSON!, $groupName: String!, $property: String!, $pattern: String!) {
                        createControlExecutionStealthByPropertyPattern(input: { name: $controlName, params: $params, groupName: $groupName, property: $property, value: $pattern, enrichWith: [] }) {
                            boolean
                        }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)

    tmpl.SetStrRepl("controlName",  controlName)
    tmpl.SetStrRepl("params",       params)

    tmpl.SetStrRepl("groupName",    groupName)
    tmpl.SetStrRepl("property",     property)
    tmpl.SetStrRepl("pattern",      pattern)

    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return err
    }

    var gqResp CreateControlExecutionStealthResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return err
    }

    if gqResp.Errors != nil {
        err = errors.New("create control execution stealth: " + gqResp.Errors.GetMessages())
        return err
    }
    return err
}
//
// CreateControlExecutionStealthByPropertyValue()
//
func (this *Pixcore) CreateControlExecutionStealthByPropertyValue(controlName string, params JSON, groupName string, property string, value string) error {
    var err error

    gqReq := `{
        "variables": {
                "controlName": "<<controlName>>",
                "params": "<<params>>",
                "groupName": "<<groupName>>",
                "property": "<<property>>",
                "value": "<<value>>"
        },
        "query": "mutation CreateControlExecutionStealthByPropertyValue( $controlName: String!, $params: JSON!, $groupName: String!, $property: String!, $value: String!) {
                        createControlExecutionStealthByPropertyValue(input: { name: $controlName, params: $params, groupName: $groupName, property: $property, value: $value, enrichWith: [] }) {
                            boolean
                        }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)

    tmpl.SetStrRepl("controlName",  controlName)
    tmpl.SetStrRepl("params",       params)

    tmpl.SetStrRepl("groupName",    groupName)
    tmpl.SetStrRepl("property",     property)
    tmpl.SetStrRepl("value",        value)

    gqReq = tmpl.Pack()

    httpRespBody, err := this.httpRequest(gqReq)
    if err != nil {
        return err
    }

    var gqResp CreateControlExecutionStealthResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return err
    }

    if gqResp.Errors != nil {
        err = errors.New("create control execution stealth: " + gqResp.Errors.GetMessages())
        return err
    }
    return err
}
//EOF
