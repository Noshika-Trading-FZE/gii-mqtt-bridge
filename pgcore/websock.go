/*
 * Copyright:  Pixel Networks <support@pixel-networks.com> 
 */

package pgcore

import (
    "context"
    "encoding/json"
    "time"

    "net/http"
    "net/url"
    "sync"

    "app/pgquery"
    "app/pmlog"
    
    "github.com/gorilla/websocket"
)

//https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md

const (
    gPayloadMessageKey      string = "message"

    gwsConnectionInit       string  = "connection_init"
    gwsConnectionError      string  = "conn_err"
    gwsStart                string  = "start"
    gwsStop                 string  = "stop"
    gwsError                string  = "error"
    gwsData                 string  = "data"
    gwsComplete             string  = "complete"
    gwsConnectionKeepAlive  string  = "ka"
    gwsConnectionAck        string  = "connection_ack"
    gwsConnectionTerminate  string  = "connection_terminate"
    gwsUnknown              string  = "unknown"
    gwsInternal             string  = "internal"
)

type GMessage struct {
    Id      string         `json:"id,omitempty"`
    Type    string         `json:"type"`
    Payload interface{}    `json:"payload,omitempty"`
}

func NewGMessage(mtype string, payload interface{}) *GMessage {
    if payload == nil {
        payload = pgquery.NewGQuery()
    }
    return &GMessage{
        Type:        mtype,
        Payload:     payload,
    }
}
func (this *GMessage) GetJSON() string {
    result, _ := json.Marshal(this)
    return string(result)
}

func (this *GMessage) GetMessage() string {
    var message string
    if this.Payload != nil {
        payload := this.Payload.(map[string]interface{})   // !!! dangeros! 
        j, _ := json.Marshal(payload[gPayloadMessageKey])
        message = string(j)
    }
    return message
}

type GwsDataHandlerFunc = func(string) error

//
// GetWsURL()
//
func (this *Pixcore) GetWsURL() (*url.URL, error) {
    var err error
    var url *url.URL


    schema      := "ws://"              // !!! todo: case wss if https
    hostname    := this.gqURL.Hostname()
    port        := this.gqURL.Port()
    path        := this.gqURL.EscapedPath()
    wsRef       := schema + hostname + ":" + port + path
    url, err = url.Parse(wsRef)
    if err != nil {
        return url, err
    }
    return url, err
}

const wsReadTimeout time.Duration = 20
type SubscribeLoopFunc = func()
//
// Subscribe
//
func (this *Pixcore) Subscribe(externalParentCtx context.Context, wg *sync.WaitGroup, gQuery *pgquery.GQuery, dataHandler GwsDataHandlerFunc) (SubscribeLoopFunc, context.CancelFunc, error) {
    var err error
    var loopFunc SubscribeLoopFunc

    wsCtx, wsCancel := context.WithCancel(externalParentCtx)

    wsRef, err := this.GetWsURL()
    if err != nil {
        return loopFunc, wsCancel, err
    }

    headers := make(http.Header) 
    headers.Add("Sec-Websocket-Protocol", "graphql-ws")
    headers.Add("Authorization", "Bearer " + this.GetJWTToken())

    // Start Level1
    wsConn, _, err := websocket.DefaultDialer.DialContext(wsCtx, wsRef.String(), headers)
    if err != nil {
        return loopFunc, wsCancel, err
    }

    closeHandler := func(code int, text string) error {
        var err error
        pmlog.LogInfo("web socket closed on level1 handler, code:", code, "text:", text)
        // Send terminate message on Level2, wo control response
        gMessage := NewGMessage(gwsConnectionTerminate, nil)
        err = wsConn.WriteJSON(gMessage)
        wsCancel()
        return err
    }
    wsConn.SetCloseHandler(closeHandler)

    //wsConn.SetPongHandler(func(str string) error {
    //    var err error
    //    pmlog.LogDebug("#### pong received:", str)
    //    return err
    //})

    //wsConn.SetPingHandler(func(str string) error {
    //    var err error
    //    pmlog.LogDebug("#### ping received:", str)
    //    return err
    //})
    
    // Start Level2
    gMessage := NewGMessage(gwsConnectionInit, nil)
    err = wsConn.WriteJSON(gMessage)
    if err != nil {
        return loopFunc, wsCancel, err
    }

    wg.Add(1)
    loopFunc = func() {
        defer wg.Done()
        defer pmlog.LogWarning("subsription done")
        
        for {
            // Check context
            // !!! todo: may be in background?
            select {
                case <- wsCtx.Done():
                    pmlog.LogInfo("subsription canceled")
                    // Send terminate message on Level2, wo control response
                    gMessage := NewGMessage(gwsConnectionTerminate, nil)
                    err = wsConn.WriteJSON(gMessage)
                    if err != nil {
                        pmlog.LogInfo("gws write connection terminate error:", err)
                    }
                    wsCancel()
                    return
                default:
            }

            err = wsConn.SetReadDeadline(time.Now().Add(wsReadTimeout * time.Second))
            if err != nil {
                err = wsConn.Close()
                pmlog.LogInfo("gws close websocket")
                if err != nil {
                    pmlog.LogError("gws closing websocket error:", err)
                }
                wsCancel()
                pmlog.LogInfo("gws canceled websocket and exit from loop")
                continue 
            }
            
            var gMessage GMessage
            err = wsConn.ReadJSON(&gMessage)

            //pmlog.LogDebug("raw gws message:", gMessage.GetJSON())

            if err != nil {
                pmlog.LogInfo("gws inloop reading error:", err)
                //subCancel()

                    gMessage := NewGMessage(gwsConnectionTerminate, nil)
                    err = wsConn.WriteJSON(gMessage)
                    if err != nil {
                        pmlog.LogInfo("gws write connection terminate error:", err)
                    }
                    wsCancel()


                return
            }
            
            switch gMessage.Type {
                case gwsConnectionAck:
                    gMessage := NewGMessage(gwsStart, gQuery)
                    err = wsConn.WriteJSON(gMessage)
                    if err != nil {
                        pmlog.LogError("gws start writing error:", err)
                    }

                case gwsConnectionKeepAlive:
                    //pmlog.LogDebug("#### gws got keepalive message")
                    continue

                case gwsData:
                    data, err := json.MarshalIndent(gMessage.Payload, " ", "    ")
                    if err != nil {
                        pmlog.LogError("gws data marshall error:", err)
                        continue
                    }
                    go dataHandler(string(data))        // Async handling

                case gwsConnectionError:
                    pmlog.LogError("gws got connection error message: "  + gMessage.GetMessage())
                    gMessage := NewGMessage(gwsConnectionTerminate, nil)
                    err := wsConn.WriteJSON(gMessage)
                    if err != nil {
                        pmlog.LogError("gws connection terminate writing error:", err)
                    }
                    err = wsConn.Close()
                    pmlog.LogInfo("gws close websocket")
                    if err != nil {
                        pmlog.LogError("gws closing websocket error:", err)
                    }
                    wsCancel()
                    pmlog.LogInfo("gws canceled websocket and exit from loop")
                    return
                    //return errors.New("connection error with message:" + gMessage.GetMessage())

                case gwsComplete:
                    pmlog.LogWarning("gws got request compete message: "  + gMessage.GetMessage())
                    gMessage := NewGMessage(gwsConnectionTerminate, nil)
                    err := wsConn.WriteJSON(gMessage)
                    if err != nil {
                        pmlog.LogError("gws connection terminate writing error:", err)
                    }
                    err = wsConn.Close()
                    pmlog.LogInfo("gws close websocket")
                    if err != nil {
                        pmlog.LogError("gws closing websocket error:", err)
                    }
                    wsCancel()
                    pmlog.LogInfo("gws canceled websocket and exit from loop")
                    return

                case gwsError:
                    pmlog.LogError("gws got generic error message: " + gMessage.GetMessage())
                    gMessage := NewGMessage(gwsConnectionTerminate, nil)
                    err := wsConn.WriteJSON(gMessage)
                    if err != nil {
                        pmlog.LogError("gws unable write message:", err)
                    }
                    err = wsConn.Close()
                    pmlog.LogInfo("gws close websocket")
                    if err != nil {
                        pmlog.LogError("gws closing websocket error:", err)
                    }
                    wsCancel()
                    pmlog.LogInfo("gws canceled websocket and exit from loop")
                    return

                default:
                    pmlog.LogInfo("gws got unknown type of message:", gMessage.GetJSON())
                    continue
            } 
        }
    }
    return loopFunc, wsCancel, err
}
//EOF
