/*
 * Copyright:  Pixel Networks <support@pixel-networks.com> 
 */

package pgcore

import (
    "context"
    "encoding/json"
    "encoding/base64"
    "sync"
    
    "app/pgquery"
    "app/pmlog"
)

type gwsDataHandlerFunc = func(string) error

type SubscrOnControlHandlerFunc = func(ControlExecutionMessage) error

type SubscrOnControlMessage struct {
	Data struct {
		Listen struct {
            ControlExecution    ControlExecutionMessage     `json:"controlExecution"`
            RelatedNodeId       string                      `json:"relatedNodeId"`
		} `json:"listen"`
	} `json:"data"`
}

type fastHelperParams struct {
    Enabled  bool   `json:"enabled"`
    SchemaId string `json:"schema_id"`
}
    
type HelperParams struct {
    Enabled  bool   `json:"enabled"`
    SchemaId string `json:"schema_id"`
}


type ControlExecutionMessage struct {
    Id          int64       `json:"id"`

    CallerId    UUID        `json:"callerId"`        // From:
    Controller  UUID        `json:"controller"`      // ActualReceiver:
    ObjectId    UUID        `json:"objectId"`        // To:
    Name        string      `json:"name"`
    Params      string      `json:"params"`
    Type        string      `json:"type"`

    Ack         bool        `json:"ack"`
    Done        bool        `json:"done"`

	//HelperParams HelperParams `json:"helperParams,omitempty"`
    HelperData  HelperParams      `json:"helperData,omitempty"`

    Error       string      `json:"error"`
}
func (this *ControlExecutionMessage) GetJSON() string {
    result, _ := json.Marshal(this)
    return string(result)
}

// for FastRPC mapping
/*
{
  "id": null,
  "object_id": "c0f8ceb4-6555-400f-87b2-4d3dfc7700b1",
  "controller": "4febcecb-5bf6-4a94-9bfb-ebd4b4598755",
  "created_at": "2021-05-15T06:37:25.759724+00:00",
  "type": "RPC_STEALTH",
  "name": "Hello",
  "params": {},
  "ack": null,
  "done": null,
  "error": null,
  "linked_control_id": null,
  "caller_id": "95c71b06-79cb-48c4-acc8-2310a50e8872",
  "helper_params": null
}
 */
 
type fastControlExecutionMessage struct {
    Id              int64       `json:"id"`
    CallerId        UUID        `json:"caller_id"`
    Controller      UUID        `json:"controller"`
    ObjectId        UUID        `json:"object_id"`
    Name            string      `json:"name"`
    Type            string      `json:"type"`
    Ack             bool        `json:"ack"`
    Done            bool        `json:"done"`
    Error           string      `json:"error"`
	HelperParams    fastHelperParams `json:"helper_params"`
    Params          string `json:"params"`
}

func (this *Pixcore) SubscrOnControl(parentCtx context.Context, wg *sync.WaitGroup, controlMessageHandler SubscrOnControlHandlerFunc) (SubscribeLoopFunc, context.CancelFunc, error) {
    var err error
    var subscrCancel context.CancelFunc
    var loopFunc SubscribeLoopFunc

    gQuery := pgquery.NewGQuery()
    gQuery.AddStrVar("topic", "controls:" + this.GetTokenId())

    gQuery.Query = `subscription SubscrOnControl($topic: String!) {
                        listen(topic: $topic){
                            controlExecution: relatedNode {
                                ... on ControlExecution {
                                    id
                                    objectId
                                    callerId
                                    createdAt
                                    controller
                                    name
                                    params
                                    helperData: helperParams
                                    type
                                    ack
                                    done
                                }
                            }
                            relatedNodeId
                        }
                }`

    dataHandler := func(data string) error {
        var err error
        var message SubscrOnControlMessage
        err = json.Unmarshal([]byte(data), &message)
        if err != nil {
            return err
        }

        // FastRPC apapter
        if len(message.Data.Listen.ControlExecution.CallerId) == 0 {
            jMessage, err := base64.StdEncoding.DecodeString(message.Data.Listen.RelatedNodeId)
            if err != nil {
                pmlog.LogError("error decoding relatedNodeId:", err)
                return err
            }
            messages := make([]fastControlExecutionMessage, 0)
            err = json.Unmarshal([]byte(jMessage), &messages)
            if err != nil {
                pmlog.LogError("error unmarshal fastControlExecutionMessage:", err)
                return err
            }
            
            var control ControlExecutionMessage
            control.Id          = -1
            control.CallerId    = messages[0].CallerId
            control.Controller  = messages[0].Controller
            control.ObjectId    = messages[0].ObjectId
            control.Name        = messages[0].Name
            control.Type        = messages[0].Type
            control.Done        = true
            control.HelperData.Enabled = messages[0].HelperParams.Enabled
            control.HelperData.SchemaId = messages[0].HelperParams.SchemaId

            control.Params      = messages[0].Params
            message.Data.Listen.ControlExecution = control
        }

        err = controlMessageHandler(message.Data.Listen.ControlExecution)
        if err != nil {
            return err
        }
        return err
    }

    loopFunc, subscrCancel, err = this.Subscribe(parentCtx, wg, gQuery, dataHandler)
    if err != nil {
        return loopFunc, subscrCancel, err
    }
    return loopFunc, subscrCancel, err
}
//
// Object
//
type SubscrOnObjectMessage struct {
	Data struct {
		Listen struct {
            Object ObjectMessage `json:"object"`
			RelatedNodeId string `json:"relatedNodeId"`
		} `json:"listen"`
	} `json:"data"`
}

type ObjectMessage struct {
    Enabled bool   `json:"enabled"`
    Id      string `json:"id"`
    Name    string `json:"name"`
    Schema  struct {
        ApplicationOwner string   `json:"applicationOwner"`
        Enabled          bool     `json:"enabled"`
        Id               string   `json:"id"`
        MTags            []string `json:"mTags"`
        MVersion         string   `json:"mVersion"`
        Name             string   `json:"name"`
    } `json:"schema"`
}

func (this *ObjectMessage) GetJSON() string {
    result, _ := json.Marshal(this)
    return string(result)
}

type SubscrOnObjectHandlerFunc = func(ObjectMessage) error

func (this *Pixcore) SubscrOnObject(parentCtx context.Context, wg *sync.WaitGroup, objectMessageHandler SubscrOnObjectHandlerFunc) (SubscribeLoopFunc, context.CancelFunc, error) {
    var err error
    var subscrCancel context.CancelFunc
    var loopFunc SubscribeLoopFunc


    gQuery := pgquery.NewGQuery()
    gQuery.AddStrVar("topic", "objects:" + this.GetTokenId()) 

    gQuery.Query = `subscription SubscrOnObject($topic: String!) {
                        listen(topic: $topic){
                            object: relatedNode {
                                ...on Object{
                                    id
                                    name
                                    enabled
                                    schema{
                                        id
                                        applicationOwner
                                        mTags
                                        mVersion
                                        enabled
                                        name
                                    }
                                }
                            }
                            relatedNodeId
                        }
            }`

    dataHandler := func(data string) error {
        var err error
        var message SubscrOnObjectMessage
        err = json.Unmarshal([]byte(data), &message)
        if err != nil {
            return err
        }
        err = objectMessageHandler(message.Data.Listen.Object)
        if err != nil {
            return err
        }
        return err
    }

    loopFunc, subscrCancel, err = this.Subscribe(parentCtx, wg, gQuery, dataHandler)
    if err != nil {
        return loopFunc, subscrCancel, err
    }
    return loopFunc, subscrCancel, err
}

//
// ObjectProperty
//
type SubscrOnObjectPropertyMessage struct {
	Data struct {
		Listen struct {
			ObjectProperty ObjectPropertyMessage `json:"objectProperty"`
			RelatedNodeId string `json:"relatedNodeId"`
		} `json:"listen"`
	} `json:"data"`
}

type ObjectPropertyMessage struct {
    Id          string    `json:"id"`
    ObjectId    string    `json:"objectId"`
    Property    string    `json:"property"`
    Stealth     bool      `json:"stealth"`
    Type        string    `json:"type"`
    Value       string    `json:"value"`

    Object struct {
        Id          string   `json:"id"`
        Enabled     bool     `json:"enabled"`
        SchemaId    string   `json:"schemaId"`
        SchemaTags  []string `json:"schemaTags"`
        Schema struct {
            Enabled     bool     `json:"enabled"`
            Id          string   `json:"id"`
            MTags       []string `json:"mTags"`
            Name        string   `json:"name"`
        } `json:"schema"`
    } `json:"object"`
    
}
func (this *ObjectPropertyMessage) GetJSON() string {
    result, _ := json.Marshal(this)
    return string(result)
}

type SubOnObjectPropertyHandlerFunc = func(ObjectPropertyMessage) error

func (this *Pixcore) SubscrOnObjectProperty(parentCtx context.Context, wg *sync.WaitGroup, objectPropertyMessageHandler SubOnObjectPropertyHandlerFunc) (SubscribeLoopFunc, context.CancelFunc, error) {
    var err error
    var subscrCancel context.CancelFunc
    var loopFunc SubscribeLoopFunc

    gQuery := pgquery.NewGQuery()
    gQuery.AddStrVar("topic", "objects:" + this.GetTokenId()) 
    gQuery.Query = `subscription SubscrOnObjectProperty($topic: String!) {
                        listen(topic: $topic){
                            objectProperty: relatedNode {
                                ... on ObjectProperty {
                                    id
                                    type
                                    value
                                    stealth
                                    property
                                    objectId
                                    object {
                                          id
                                          enabled
                                          schemaId
                                          schemaTags
                                          schema {
                                                enabled
                                                name
                                                id
                                                mTags
                                          }
                                    }
                                }
                            }
                            relatedNodeId
                        }
            }`

    dataHandler := func(data string) error {
        var err error
        var message SubscrOnObjectPropertyMessage

        err = json.Unmarshal([]byte(data), &message)
        if err != nil {
            return err
        }
        err = objectPropertyMessageHandler(message.Data.Listen.ObjectProperty)
        if err != nil {
            return err
        }
        return err
    }

    loopFunc, subscrCancel, err = this.Subscribe(parentCtx, wg, gQuery, dataHandler)
    if err != nil {
        return loopFunc, subscrCancel, err
    }
    return loopFunc, subscrCancel, err
}
//EOF
