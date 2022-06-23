/*
 * Copyright:  Pixel Networks <support@pixel-networks.com>
 */

package main

import (
    "errors"
    "encoding/json"
    "context"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "time"
    "sync"
    "strconv"

    "app/pmlog"
    "app/pmconfig"
    "app/pgschema"
    "app/pgcore"
    "app/pmtools"
    "app/pmtopics"
    "app/mqtrans"

    mqtt "github.com/eclipse/paho.mqtt.golang"
)
//
// Main
//
func main() {
    app := NewApplication()
    err := app.StartApplication()
    if err != nil {
        pmlog.LogError("app error:", err)
        os.Exit(1)
    }
}
//
//
type Application struct {
    schema              *pgschema.Schema
    config              *pmconfig.Config
    pg                  *pgcore.Pixcore
    tr                  *mqtrans.Transport

    brokerUrl           string
    username            string
    password            string

    userId              pgschema.UUID
    objectId            pgschema.UUID

    appCtx              context.Context
    appCancel           context.CancelFunc

    controlWatcherCtx       context.Context
    controlWatcherCancel    context.CancelFunc
    controlWatcherWG        *sync.WaitGroup

    propertyWatcherCtx      context.Context
    propertyWatcherCancel   context.CancelFunc
    propertyWatcherWG       *sync.WaitGroup

    subscrControlCancel     context.CancelFunc
    subscrControlWG         *sync.WaitGroup

    subscrPropertyCancel context.CancelFunc
    subscrPropertyWG     *sync.WaitGroup

    topics              *pmtopics.Topics

    autoProvision       bool
    autoProvisionMutex  sync.Mutex
}
//
//
func (this *Application) SetAutoProvision(state bool) {
    this.autoProvisionMutex.Lock()
    defer this.autoProvisionMutex.Unlock()
    this.autoProvision = state
}
//
//
func (this *Application) GetAutoProvision() bool {
    this.autoProvisionMutex.Lock()
    defer this.autoProvisionMutex.Unlock()
    return this.autoProvision
}
//
//
func NewApplication() *Application {
    var app Application

    app.appCtx, app.appCancel = context.WithCancel(context.Background())

    app.controlWatcherCtx, app.controlWatcherCancel = context.WithCancel(app.appCtx)
    var controlWatcherWG sync.WaitGroup
    app.controlWatcherWG = &controlWatcherWG 

    app.propertyWatcherCtx, app.propertyWatcherCancel = context.WithCancel(app.appCtx)
    var propertyWatcherWG sync.WaitGroup
    app.propertyWatcherWG = &propertyWatcherWG 
    
    var subscrControlWG sync.WaitGroup
    app.subscrControlWG = &subscrControlWG

    var subscrPropertyWG sync.WaitGroup
    app.subscrPropertyWG = &subscrPropertyWG

    app.config  = pmconfig.New()
    app.pg      = pgcore.New(app.appCtx)
    app.schema  = pgschema.NewSchema()
    app.tr      = mqtrans.NewTransport()
    app.topics  = pmtopics.NewTopics()

    return &app
}
//
//
func (this *Application) GetConfig() error {
    var err error
    exeName := filepath.Base(os.Args[0])
    this.config.Read(exeName + ".yml")

    flag.Usage = func() {
        fmt.Println(exeName + " version " + this.config.Version)
        fmt.Println("")
        fmt.Printf("usage: %s command [option]\n", exeName)
        fmt.Println("")
        flag.PrintDefaults()
        fmt.Println("")
    }
    flag.Parse()
    if len(os.Getenv("CONFIG_API_USERNAME")) > 0 {
        this.config.Core.Username = os.Getenv("CONFIG_API_USERNAME")
    }
    if len(os.Getenv("CONFIG_API_PASSWORD")) > 0 {
        this.config.Core.Password = os.Getenv("CONFIG_API_PASSWORD")
    }
    if len(os.Getenv("CONFIG_API_URL")) > 0 {
        this.config.Core.URL = os.Getenv("CONFIG_API_URL")
    }
    return err
}
//
//
func (this *Application) StartApplication() error {
    var err error
    pmlog.LogInfo("trying to start application")
    err = this.GetConfig()
    if err != nil {
        return err
    }

    err = this.DefineAppSchema()
    if err != nil {
        return err
    }
    err = this.ConnectionSetup()
    if err != nil {
        return err
    }

    err = this.BindCore()
    if err != nil {
        return err
    }

    err = this.GetUserId()
    if err != nil {
        return err
    }

    err = this.SetupAppSchema()
    if err != nil {
        return err
    }

    err = this.BindCore()
    if err != nil {
        return err
    }

    err = this.GetAppObjectId()
    if err != nil {
        return err
    }
    err = this.GetAppProperties()
    if err != nil {
        return err
    }

    err = this.GetTransProperties()
    if err != nil {
        return err
    }
    err = this.BindTransport()
    if err != nil {
        pmlog.LogError("application transport error:", err)
    }
    err = this.SubscribeToTopics()
    if err != nil {
        pmlog.LogError("application transport error:", err)
    }

    err = this.StartControlSubsription()
    if err != nil {
        return err
    }
    go this.StartControlSubsrWatcher()

    //err = this.StartPropertySubsription()
    //if err != nil {
    //    return err
    //}
    //go this.StartPropertySubsrWatcher()

    err = this.StartLoop()
    if err != nil {
        return err
    }
    return err
}

func (this *Application) ReStartApplication() error {
    var err error
    pmlog.LogInfo("trying to restart application")

    this.UnbindTransport()

    //this.StopWoWPropertySubsrWatcher()
    //this.StopWWPropertySubsription()
    //this.WaitSPropertySubsrWatcher()

    this.StopWoWControlSubsrWatcher()
    this.StopWWControlSubsription()
    this.WaitSControlSubsrWatcher()

    err = this.BindCore()
    if err != nil {
        return err
    }

    err = this.GetUserId()
    if err != nil {
        return err
    }

    err = this.SetupAppSchema()
    if err != nil {
        return err
    }

    err = this.BindCore()
    if err != nil {
        return err
    }
    err = this.GetAppObjectId()
    if err != nil {
        return err
    }
    err = this.GetAppProperties()
    if err != nil {
        return err
    }
    err = this.GetTransProperties()
    if err != nil {
        return err
    }
    
    //err = this.BindTransport()
    //if err != nil {
        //pmlog.LogError("application transport error:", err)
    //}
    //err = this.SubscribeToTopics()
    //if err != nil {
        //pmlog.LogError("application transport error:", err)
    //}

    err = this.StartControlSubsription()
    if err != nil {
        return err
    }
    go this.StartControlSubsrWatcher()

    //err = this.StartPropertySubsription()
    //if err != nil {
    //    return err
    //}
    //go this.StartPropertySubsrWatcher()

    return err
}
//
//
func (this *Application) ConnectionSetup() error { 
    err := this.pg.Setup(this.config.Core.URL, this.config.Core.Username,
                    this.config.Core.Password, this.config.Core.JwtTTL,
                    this.schema.Metadata.MTags)
    return err
}
//
//
func (this *Application) BindCore() error { 
    var err error
    timer := time.NewTicker(bindReconnectInterval * time.Second)
    for _ = range timer.C {
        err := this.pg.Bind()
        if err != nil {
            pmlog.LogInfo("application wainting connection to core, error:", err)
            continue
        }
        pmlog.LogInfo("application connected to core")
        break
    }
    return err
}
//
//
func (this *Application) SetupAppSchema() error {
    var err error
    pmlog.LogInfo("trying to import application schema")
    schemaJson := this.schema.GetJSON()
    schemaId, err := this.pg.ImportSchema(schemaJson)
    if err != nil {
        return err
    }
    pmlog.LogInfo("done import application schema", schemaId)
    return err
}
//
//
func (this *Application) GetAppObjectId() error {
    var err error

    pmlog.LogInfo("trying to get application object id")
    this.objectId, err = this.pg.GetUserProfileId()
    if err != nil {
        return err
    }

    pmlog.LogInfo("application object id is ", this.objectId)
    return err
}
//
//
func (this *Application) GetAppProperties() error {
    var err error
    pmlog.LogInfo("application trying to get own property")

    autoProvisionStr, err := this.pg.GetObjectPropertyValue(this.objectId, propertyAutoProvisionName)
    if err != nil {
        return err
    }
    autoProvisionBool, err := strconv.ParseBool(autoProvisionStr)
    if err != nil {
        return err
    }
    this.autoProvision = autoProvisionBool

    pmlog.LogInfo("application got own property")
    return err
}



func (this *Application) GetTransProperties() error {
    var err error
    pmlog.LogInfo("application trying to get own property")

    this.brokerUrl, err = this.pg.GetObjectPropertyValue(this.objectId, mqttPropertyBrokerUrlName)
    if err != nil {
        return err
    }
    this.username, err = this.pg.GetObjectPropertyValue(this.objectId, mqttPropertyUsernameName)
    if err != nil {
        return err
    }
    this.password, err = this.pg.GetObjectPropertyValue(this.objectId, mqttPropertyPasswordName)
    if err != nil {
        return err
    }

    topicsString, err := this.pg.GetObjectPropertyValue(this.objectId, mqttPropertyTopicsName)
    if err != nil {
        return err
    }
    pmlog.LogInfo("application got topics", topicsString)
    this.topics = pmtopics.TopicsFromString(topicsString)

    pmlog.LogInfo("application got own property")
    return err
}
//
//
func (this *Application) GetUserId() error {
    var err error
    userId, err := this.pg.GetUserIdByLogin(this.config.Core.Username)
    if err != nil {
        return err
    }
    pmlog.LogInfo("application received user id:", userId)
    this.userId = userId
    return err
}
//
//
func (this *Application) BindTransport() error {
    var err error

    pmlog.LogInfo("application trying to connect to broker", this.brokerUrl )
    err = this.tr.Bind(this.brokerUrl, this.username, this.password)
    if err != nil {
        return err
    }
    pmlog.LogInfo("application successfully contacted to broker", this.brokerUrl)
    if this.tr.IsConnected() {
        this.pg.UpdateObjectPropertyByName(this.objectId, propertyConnectionStateName, "connected")
        if err != nil {
            return err
        }
    }
    return err
}
//
//
func (this *Application) UnbindTransport() error {
    var err error
    this.tr.Disconnect()
    pmlog.LogInfo("application disconnected from broker")

    this.pg.UpdateObjectPropertyByName(this.objectId, propertyConnectionStateName, "disconnected")
    if err != nil {
        return err
    }

    return err
}
//
//
func (this *Application) TransportReconnect() error {
    var err error
    pmlog.LogInfo("application will try to reconnect transport")

    this.UnbindTransport()

    //err = this.GetTransProperties()
    //if err != nil {
    //    return err
    //}

    err = this.BindTransport()
    if err != nil {
        return err
    }
    if this.tr.IsConnected() {
        pmlog.LogInfo("transport reconnected successfully and will subscribe to topics")
        err = this.SubscribeToTopics()
        if err != nil {
            return err
        }
    }
    return err
}
//
//
func (this *Application) TransportRestart() error {
    var err error
    pmlog.LogInfo("application will try to restart transport")

    this.UnbindTransport()

    err = this.GetTransProperties()
    if err != nil {
        return err
    }

    err = this.BindTransport()
    if err != nil {
        return err
    }
    if this.tr.IsConnected() {
        pmlog.LogInfo("transport restart successfully and will subscribe to topics")
        err = this.SubscribeToTopics()
        if err != nil {
            return err
        }
    }
    return err
}
//
//
func (this *Application) SubscribeToTopics() error {
    var err error

    topicHandler := this.CreateRealTopicHandler()

    for _, topic := range this.topics.GetArray() {
        err := this.tr.Subscribe(topic, topicHandler)
        if err != nil {
            pmlog.LogInfo("application", this.objectId, "with topic:", topic, "have subsription error:", err)
        }
        pmlog.LogInfo("application", this.objectId, "subscribe to topic", topic)
    }
    pmlog.LogInfo("application", this.objectId, "set up topic handling")
    return err
}
//
// BridgeApp: StartLoop()
//
const (
    aliveMessage            string          = "Alive"
    jwtExpireTh             int64           = 15
    bindReconnectInterval   time.Duration   = 1   // sec
    loopInterval            time.Duration   = 1   // sec
    aliveInterval           time.Duration   = 20  // sec
)

func (this *Application) StartLoop() error {
    var err error
    pmlog.LogInfo("start application loop")

    for {
        needRestart := false
        time.Sleep(loopInterval * time.Second)

        if !this.tr.IsConnected() {
            pmlog.LogInfo("mqtt transport will reconnect")
            this.TransportReconnect()
        }

        if (time.Now().Unix() % int64(aliveInterval)) == 0 {
            pmlog.LogInfo("application is still alive")
            _, err = this.pg.UpdateObjectPropertyByName(this.objectId, propertyMessageName, aliveMessage)
            if err != nil {
                pmlog.LogError("error update message property:", err)
                needRestart = true
            }
        }

        expirePeriod := this.pg.GetJWTExpire() - time.Now().Unix()
        if expirePeriod < jwtExpireTh {

            err = this.pg.UpdateJWToken()
            if err != nil {
                pmlog.LogError("error update jwt:", err)
                needRestart = true
            }
        }
        if needRestart {
                for {
                    pmlog.LogInfo("restart application loop")
                    err = this.ReStartApplication()
                    if err == nil {
                        break
                    }
                    time.Sleep(loopInterval * time.Second)
                }
        }
    }
    return err
}
//
//
const subcrRestartWT  time.Duration = 5 // sec
//
//
//*********************************************************************//
//
func (this *Application) StartControlSubsription() error {
    var err error
    pmlog.LogInfo("application trying to start control subscription")

    handler := func(controlMessage pgcore.ControlExecutionMessage) error {
        var err error
        err = this.RouteControlMessage(controlMessage)
        if err != nil {
            pmlog.LogError("control error:", err)
        }
        return err
    }
    loopFunc, cancel, err := this.pg.SubscrOnControl(this.appCtx,
                                        this.subscrControlWG, handler)
    if err != nil {
        return err
    }
    this.subscrControlCancel = cancel

    go loopFunc()
    pmlog.LogInfo("application started control subscription")
    return err
}
//
//
func (this *Application) StopWWControlSubsription() error {
    var err error
    pmlog.LogInfo("application trying to stop control subscription")
    this.subscrControlCancel()
    this.subscrControlWG.Wait()
    pmlog.LogInfo("application stoped control subscription")
    return err
}
//
//
//
func (this *Application) StartControlSubsrWatcher() error {
    var err error

    pmlog.LogInfo("start control subscription watcher")
    this.controlWatcherCtx, this.controlWatcherCancel = context.WithCancel(this.appCtx)
    this.controlWatcherWG.Add(1)
    
    for {
        this.subscrControlWG.Wait()
        select {
            case <- this.controlWatcherCtx.Done():
                pmlog.LogInfo("control subscription watcher canceled")
                this.controlWatcherWG.Done()
                return err
            default:
        }

        time.Sleep(subcrRestartWT * time.Second)
        
        pmlog.LogWarning("application trying to re-start control subscription")
        err := this.StartControlSubsription()
        if err != nil {
            pmlog.LogError("application control subscription error:", err)
            continue
        }
        pmlog.LogWarning("application re-started control subscription")
    }
    return err
}
//
//
func (this *Application) StopWoWControlSubsrWatcher() error {
    var err error
    pmlog.LogInfo("application trying to stop control subscription watcher")
    this.controlWatcherCancel()
    return err
}

func (this *Application) WaitSControlSubsrWatcher() error {
    var err error
    pmlog.LogInfo("application waiting stop control subscription watcher")
    this.controlWatcherWG.Wait()
    pmlog.LogInfo("application stoped control subscription watcher")
    return err
}
//
//*********************************************************************//
//
func (this *Application) RoutePropertyMessage(propertyMessage pgcore.ObjectPropertyMessage) error {
    var err error
    if propertyMessage.ObjectId == this.objectId {
        //pmlog.LogDebug("route property message to application", propertyMessage.GetJSON())
        switch propertyMessage.Property {
            case mqttPropertyBrokerUrlName:
                this.brokerUrl = propertyMessage.Value
                err := this.TransportReconnect()
                if err != nil {
                    pmlog.LogError("transport reconnect error", err)
                }
            case mqttPropertyUsernameName:
                this.username = propertyMessage.Value
                err := this.TransportReconnect()
                if err != nil {
                    pmlog.LogError("transport reconnect error", err)
                }
            case mqttPropertyPasswordName:
                this.password = propertyMessage.Value
                err := this.TransportReconnect()
                if err != nil {
                    pmlog.LogError("transport reconnect error", err)
                }
            case mqttPropertyTopicsName:
                topicsString := propertyMessage.Value
                this.topics = pmtopics.TopicsFromString(topicsString)
                err := this.TransportReconnect()
                if err != nil {
                    pmlog.LogError("transport reconnect error", err)
                }
        }
    }
    return err
}
//
//*********************************************************************//
//
func (this *Application) StartPropertySubsription() error {
    var err error
    pmlog.LogInfo("application trying to start property subscription")

    handler := func(propertyMessage pgcore.ObjectPropertyMessage) error {
        var err error
        err = this.RoutePropertyMessage(propertyMessage)
        if err != nil {
            pmlog.LogError("property rountig error:", err)
        }
        return err
    }
    loopFunc, cancel, err := this.pg.SubscrOnObjectProperty(this.appCtx,
                                        this.subscrPropertyWG, handler)
    if err != nil {
        return err
    }
    this.subscrPropertyCancel = cancel

    go loopFunc()
    pmlog.LogInfo("application started property subscription")
    return err
}
//
//
func (this *Application) StopWWPropertySubsription() error {
    var err error
    pmlog.LogInfo("application trying to stop property subscription")
    this.subscrPropertyCancel()
    this.subscrPropertyWG.Wait()
    pmlog.LogInfo("application stoped property subscription")
    return err
}
//
//
func (this *Application) StartPropertySubsrWatcher() error {
    var err error

    this.propertyWatcherCtx, this.propertyWatcherCancel = context.WithCancel(this.appCtx)
    this.propertyWatcherWG.Add(1)
    
    pmlog.LogInfo("start property subscription watcher")
    for {
        this.subscrPropertyWG.Wait()
        select {
            case <- this.propertyWatcherCtx.Done():
                this.propertyWatcherWG.Done()
                pmlog.LogInfo("property subscription watcher canceled")
                return err
        }

        time.Sleep(subcrRestartWT * time.Second)
        pmlog.LogWarning("application trying to re-start property subscription")

        err = this.StartPropertySubsription()
        if err != nil {
            pmlog.LogError("application property subscription error:", err)
            this.subscrPropertyWG.Done()
            continue
        }
        pmlog.LogWarning("application re-started property subscription")
    }
    return err
}
//
//
func (this *Application) StopWoWPropertySubsrWatcher() error {
    var err error
    pmlog.LogInfo("application trying to stop property subscription watcher")
    this.propertyWatcherCancel()
    return err
}

func (this *Application) WaitSPropertySubsrWatcher() error {
    var err error
    pmlog.LogInfo("application waiting stop property subscription watcher")
    this.propertyWatcherWG.Wait()
    pmlog.LogInfo("application stoped property subscription watcher")
    return err
}
//
//*********************************************************************//
//
const (
    appSchemaVersion                    string  = "1.41"

    // Property groups
    propertyGroupMeasurement            string = "Measurements"
    propertyGroupCredential             string = "Credentials"
    propertyGroupTopics                 string = "Topics"
    propertyGroupHealthCheck            string = "HealthCheck"

    // Common properties
    propertyStatusName                  string = "Status"
    propertyMessageName                 string = "Message"
    propertyTimeoutName                 string = "Timeout"
    propertyAutoProvisionName           string = "AUTO_PROVISION"

    propertyStatusDefaultValue          string = "true"
    propertyTimeoutDefaultValue         string = "120"
    propertyAutoProvisionDefaultValue   string = "true"

    // Common controls
    controlPublishName                  string  = "SendDownlink"
    controlReloadName                   string  = "Reload"
    controlTestModuleName               string  = "TestModule"
    controlSetAutoProvisionName         string  = "SetAutoProvision"

    controlSetTopicsName                string  = "SetTopics"
    controlSetBrokerURLName              string  = "SetBrokerURL"
    controlSetUsernameName              string  = "SetUsername"
    controlSetPasswordName              string  = "SetPassword"

    // MQTT specific property
    mqttPropertyBrokerUrlName           string  = "BrokerURL"
    mqttPropertyUsernameName            string  = "Username"
    mqttPropertyPasswordName            string  = "Password"
    mqttPropertyTopicsName              string  = "Topics"

    transportStateDisconnectedValue     string = "disconnected"
    transportStateConnectedValue        string = "connected"

    propertyConnectionStateName                 string = "ConnectionState"
    propertyConnectionStateNameDefaultValue     string = "undefined"

    //mqttBrokerUrlPropertyName   string  = "BrokerURL"
    //mqttUsernamePropertyName    string  = "Username"
    //mqttPasswordPropertyName    string  = "Password"
    //mqttPropertyTopicBaseName   string  = "TopicBase"

    //propertyMQTTUseOldVersionName           string = "UserOldMQTT"
    //propertyMQTTUseOldVersionDefaultValue   string  = "true"

    // MQTT specific defaults
    mqttPropertyBrokerUrlDefaultValue   string  = "tcp://v7.unix7.org:1883"
    mqttPropertyUsernameDefaultValue    string  = "device"
    mqttPropertyPasswordDefaultValue    string  = "qwerty"
    mqttPropertyTopicsDefaultValue      string  = "/gw/#,SENSO8/#"

    controlDecodePayloadName            string = "DecodePayload"

    mqttBridgeAppSchemaTag              string  = "mqtt bridge"
    mqttDriverAppSchemaTag              string  = "mqtt driver"
    mqttDriverSchemaTag                 string  = "mqtt device"

    applicationTag                      string  = "application"
    appProfileTag                       string  = "app profile"
)

func (this *Application) DefineAppSchema() error {
    var err error
    schema := pgschema.NewSchema()
    metadata := pgschema.NewMetadata()
    metadata.Id                   = this.config.AppSchemaId
    metadata.MExternalId          = this.config.AppSchemaId

    metadata.MTags                = append(metadata.MTags, applicationTag)
    metadata.MTags                = append(metadata.MTags, mqttBridgeAppSchemaTag)
    metadata.MTags                = append(metadata.MTags, appProfileTag)

    metadata.MVersion             = appSchemaVersion
    metadata.Name                 = "MQTT Bridge"
    metadata.Description          = "MQTT Bridge"
    metadata.Type                 = pgschema.MetadataTypeApp
    schema.Metadata = metadata

    schema.Controls     = append(schema.Controls, this.newPublishControl())
    schema.Controls     = append(schema.Controls, this.newPublishControlArgTopicName())
    schema.Controls     = append(schema.Controls, this.newPublishControlArgPayload())

    schema.Controls     = append(schema.Controls, this.newReloadControl())
    schema.Controls     = append(schema.Controls, this.newTestModuleControl())

    schema.Controls     = append(schema.Controls, this.newSetAutoProvisionControl())
    schema.Controls     = append(schema.Controls, this.newSetAutoProvisionControlArgEnable())

    schema.Properties   = append(schema.Properties, this.newStatusProperty())
    schema.Properties   = append(schema.Properties, this.newMessageProperty())
    schema.Properties   = append(schema.Properties, this.newTimeoutProperty())

    schema.Properties   = append(schema.Properties, this.newMqttBrokerUrlProperty())
    schema.Properties   = append(schema.Properties, this.newMqttUsernameProperty())
    schema.Properties   = append(schema.Properties, this.newMqttPasswordProperty())
    schema.Properties   = append(schema.Properties, this.newMqttTopicsProperty())

    schema.Properties   = append(schema.Properties, this.newAutoProvisionProperty())
    schema.Properties   = append(schema.Properties, this.newConnectionStateProperty())

    schema.Controls     = append(schema.Controls, this.newSetTopicsControl())
    schema.Controls     = append(schema.Controls, this.newSetTopicsControlArgTopicName())
    schema.Controls     = append(schema.Controls, this.newSetBrokerURLControl())
    schema.Controls     = append(schema.Controls, this.newSetBrokerURLControlArgTopicName())
    schema.Controls     = append(schema.Controls, this.newSetUsernameControl())
    schema.Controls     = append(schema.Controls, this.newSetUsernameControlArgTopicName())
    schema.Controls     = append(schema.Controls, this.newSetPasswordControl())
    schema.Controls     = append(schema.Controls, this.newSetPasswordControlArgTopicName())

    this.schema = schema
    return err
}
//
// BridgeApp: newMessageControl()
//
func (this *Application) newPublishControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Send raw message to topic"
    control.Hidden          = false
    control.RPC             = controlPublishName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
func (this *Application) newPublishControlArgTopicName() *pgschema.Control {
    control := this.newPublishControl()
    control.Description     = "Topic name"
    control.Type            = pgschema.StringType
    control.Argument        = "topicName"
    return control
}
func (this *Application) newPublishControlArgPayload() *pgschema.Control {
    control := this.newPublishControl()
    control.Description     = "Topic payload"
    control.Type            = pgschema.StringType
    control.Argument        = "payload"
    return control
}
//
//
func (this *Application) newReloadControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Reload bridges"
    control.Hidden          = false
    control.RPC             = controlReloadName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
//
//
func (this *Application) newTestModuleControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Test module"
    control.Hidden          = false
    control.RPC             = controlTestModuleName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
//
//
func (this *Application) newSetAutoProvisionControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Set auto provision"
    control.Hidden          = true
    control.RPC             = controlSetAutoProvisionName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
func (this *Application) newSetAutoProvisionControlArgEnable() *pgschema.Control {
    control := this.newSetAutoProvisionControl()
    control.Description     = "Enable"
    control.Type            = pgschema.BoolType
    control.Argument        = "enable"
    control.DefaultValue    = "true"
    return control
}

//
//*********************************************************************//
//
func (this *Application) newSetTopicsControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Set topic list"
    control.Hidden          = false
    control.RPC             = controlSetTopicsName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
func (this *Application) newSetTopicsControlArgTopicName() *pgschema.Control {
    control := this.newSetTopicsControl()
    control.Description     = "Topic list"
    control.Type            = pgschema.StringType
    control.Argument        = "topics"
    return control
}

//
//*********************************************************************//
//
func (this *Application) newSetBrokerURLControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Set broker URL"
    control.Hidden          = false
    control.RPC             = controlSetBrokerURLName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
func (this *Application) newSetBrokerURLControlArgTopicName() *pgschema.Control {
    control := this.newSetBrokerURLControl()
    control.Description     = "Broker URL"
    control.Type            = pgschema.StringType
    control.Argument        = "hostname"
    return control
}
//
//*********************************************************************//
//
func (this *Application) newSetUsernameControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Set username"
    control.Hidden          = false
    control.RPC             = controlSetUsernameName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
func (this *Application) newSetUsernameControlArgTopicName() *pgschema.Control {
    control := this.newSetUsernameControl()
    control.Description     = "Username"
    control.Type            = pgschema.StringType
    control.Argument        = "username"
    return control
}
//
//*********************************************************************//
//
func (this *Application) newSetPasswordControl() *pgschema.Control {
    control := pgschema.NewControl()
    control.Description     = "Set password"
    control.Hidden          = false
    control.RPC             = controlSetPasswordName
    control.Type            = pgschema.StringType
    control.Argument        = control.RPC
    return control
}
func (this *Application) newSetPasswordControlArgTopicName() *pgschema.Control {
    control := this.newSetPasswordControl()
    control.Description     = "Password"
    control.Type            = pgschema.StringType
    control.Argument        = "password"
    return control
}


//
//
func (this *Application) newStatusProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = propertyStatusName
    property.Type           = pgschema.BoolType
    property.Description    = "Application online"
    property.GroupName      = propertyGroupHealthCheck
    property.DefaultValue   = propertyStatusDefaultValue
    return property
}
//
//
func (this *Application) newMessageProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = propertyMessageName
    property.Type           = pgschema.StringType
    property.Description    = "Status message"
    property.GroupName      = propertyGroupHealthCheck
    property.DefaultValue   = pgschema.TimeUnixEpoch
    return property
}
//
//
func (this *Application) newTimeoutProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = propertyTimeoutName
    property.Type           = pgschema.IntType
    property.Description    = "Timeout for offline status"
    property.GroupName      = propertyGroupHealthCheck
    property.DefaultValue   = propertyTimeoutDefaultValue
    return property
}
//
//
func (this *Application) newAutoProvisionProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = propertyAutoProvisionName
    property.Type           = pgschema.BoolType
    property.Description    = "Auto Provision"
    property.GroupName      = propertyGroupCredential
    //property.DefaultValue   = propertyAutoProvisionDefaultValue
    return property
}
//
//
func (this *Application) newMqttBrokerUrlProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = mqttPropertyBrokerUrlName
    property.Type           = pgschema.StringType
    property.Description    = "MQTT broker url"
    property.GroupName      = propertyGroupCredential
    property.DefaultValue   = mqttPropertyBrokerUrlDefaultValue
    return property
}
//
//
func (this *Application) newMqttUsernameProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = mqttPropertyUsernameName
    property.Type           = pgschema.StringType
    property.Description    = "MQTT broker username"
    property.GroupName      = propertyGroupCredential
    property.DefaultValue   = mqttPropertyUsernameDefaultValue
    return property
}
//
//
func (this *Application) newMqttPasswordProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = mqttPropertyPasswordName
    property.Type           = pgschema.StringType
    property.Description    = "MQTT broker password"
    property.GroupName      = propertyGroupCredential
    property.DefaultValue   = mqttPropertyPasswordDefaultValue
    return property
}
//
//
func (this *Application) newMqttTopicsProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = mqttPropertyTopicsName
    property.Type           = pgschema.StringType
    property.Description    = "MQTT topics"
    property.GroupName      = propertyGroupCredential
    property.DefaultValue   = mqttPropertyTopicsDefaultValue
    return property
}
//
//
func (this *Application) newConnectionStateProperty() *pgschema.Property {
    property := pgschema.NewProperty()
    property.Property       = propertyConnectionStateName
    property.Type           = pgschema.StringType
    property.Description    = "Connection state"
    property.GroupName      = propertyGroupHealthCheck
    property.DefaultValue   = propertyConnectionStateNameDefaultValue
    return property
}
//
//
func (this *Application) CreateDummyTopicHandler() mqtrans.Handler {
    return func(client mqtt.Client, message mqtt.Message) {
        topic   := message.Topic()
        payload := message.Payload()
        pmlog.LogInfo("application", this.objectId, "got message from topic:", topic, "message:", string(payload))
    }
}
//
//
const (
    mqttPropertyTopicBaseName       string  = "TopicBase"
    mqttPropertyBridgeObjectIdName  string  = "BRIDGE"

    genericDriverSchemaId       pgschema.UUID   = "6a34e442-cc3c-4586-853e-9058e1fd7739"
    senso8BLEGWDriverSchemaId   pgschema.UUID   = "23035572-f87a-4d00-bfd3-ac68ddafb855"

    minewG1Pattern      string  = "^(/gw/)(.*)/status$"
    senso8DataPattern   string  = "^(SENSO8/nbiot/data/)(.*)$"
    senso8SysPattern    string  = "^(SENSO8/nbiot/sys/)(.*)$"
)
//
//
const maxPayloadSize    int = 64 * 1024
//
func (this *Application) CreateRealTopicHandler() mqtrans.Handler {

    return func(client mqtt.Client, message mqtt.Message) {
        var err error

        mqttTopic := message.Topic()

        //pmlog.LogDetail("receive mqtt message topic:", message.Topic(), "with payload", string(message.Payload()))

        argument := pgcore.NewTopicArguments()
        argument.TopicName  = mqttTopic
        argument.Payload    = message.Payload()

        if len(argument.Payload) > maxPayloadSize {
            pmlog.LogWarning("payload size is more 64k")
            return
        }

        controlExecution := func() {

            controlName := controlDecodePayloadName

            patterns := make([]string, 0)
            patterns = append(patterns, senso8SysPattern) 
            patterns = append(patterns, senso8DataPattern) 
            patterns = append(patterns, minewG1Pattern)
            //var patternDefined bool
            var foundTemplate string
            topicBase := mqttTopic

            for _, pattern := range patterns {
                patternDefined, _ := regexp.MatchString(pattern, mqttTopic)
                if patternDefined  {
                    foundTemplate = pattern
                    break
                }
            }

            re := regexp.MustCompile(foundTemplate)
            res := re.FindStringSubmatch(mqttTopic)

            var gwCode string
            switch foundTemplate {
                case senso8SysPattern, senso8DataPattern:
                    if len(res) == 3 {
                        gwCode = res[2]
                        topicBase = "SENSO8/nbiot/data/" + gwCode
                    }
                case minewG1Pattern:
                    if len(res) == 3 {
                        gwCode = res[2]
                        topicBase = "gw/" + gwCode
                    }
                default:
                    topicBase = mqttTopic
            }

            
            autoProvisionBool := this.GetAutoProvision()
            pmlog.LogDebug("autoProvision:", autoProvisionBool)

            if autoProvisionBool {
                patternDefined  := false
                var autoTopicBase string
                if !patternDefined  {
                    autoTopicBase, patternDefined, err = this.CheckOrCreateGenericDevice(argument.TopicName, argument.Payload)
                    if err != nil {
                        return
                    }
                }
                if patternDefined {
                    topicBase = autoTopicBase
                }
            }
    
            pmlog.LogInfo("make control message for mqtt topic:",  mqttTopic, "with topicBase:", topicBase, )

            err = this.pg.CreateControlExecutionStealthByPropertyValue(controlName, argument.Pack(), propertyGroupCredential, mqttPropertyTopicBaseName, topicBase)
            if err != nil {
                pmlog.LogError("real topic handler error: unable control call ", controlName, "with error:", err.Error())
            }
        }

        controlExecution()
        //pmlog.LogDebug("do real topic handler rpc call:", controlDecodePayloadName)
        return
    }
}
//
//
func (this *Application) CheckOrCreateGenericDevice(topicName string, payload []byte) (string, bool, error) {
    var err             error
    var topicBase       string
    var patternDefined  bool
    var foundTemplate  string 

    patterns := make([]string, 0)
    patterns = append(patterns, minewG1Pattern)
    patterns = append(patterns, senso8DataPattern) 
    
    for _, pattern := range patterns {
        patternDefined , _ = regexp.MatchString(pattern, topicName)
        if patternDefined  {
            foundTemplate = pattern
            break
        }
    }

    if !patternDefined  {
        return topicBase, patternDefined , err
    }

    re := regexp.MustCompile(foundTemplate)
    res := re.FindStringSubmatch(topicName)

    var gwCode string
    var nameHint  string 
    switch foundTemplate {
        case minewG1Pattern:
            if len(res) < 3 {
                return topicBase, patternDefined , err
            }
            topicBase = res[1] + res[2]
            gwCode = res[2]
            nameHint  = "(" + topicBase + ")"

        case senso8DataPattern:
            if len(res) < 3 {
                return topicBase, patternDefined , err
            }
            topicBase = res[1] + res[2]
            gwCode = res[2]
            nameHint  = "(" + topicBase + ")"
    }

    if len(gwCode) == 0 {
        return topicBase, patternDefined , errors.New("zero application code")
    }

    // Checking for Generic
    objects, err := this.pg.ListObjectsBySchemaId(genericDriverSchemaId)
    if err != nil {
        return topicBase, patternDefined , err
    }
    
    //for i := range objects {
    //    storedTopicBase, _ := this.pg.GetObjectPropertyValue(objects[i].Id, mqttPropertyTopicBaseName)
    //    if topicBase == storedTopicBase {
    //        return topicBase, patternDefined , err
    //    }
    //}

    // Checink for SENSO8
    sensoObjects, err := this.pg.ListObjectsBySchemaId(senso8BLEGWDriverSchemaId)
    if err != nil {
        return topicBase, patternDefined , err
    }
    objects = append(objects, sensoObjects...)

    
    for i := range objects {
        storedTopicBase, _ := this.pg.GetObjectPropertyValue(objects[i].Id, mqttPropertyTopicBaseName)
        if topicBase == storedTopicBase {
            return topicBase, patternDefined , err
        }
    }

    pmlog.LogInfo("application trying to create new generic device for ", gwCode)

    object := pgschema.NewObject()
    object.Id               = pmtools.GetNewUUID()
    object.SchemaId         = genericDriverSchemaId
    object.Name             = "Generic MQTT Device "+ nameHint  + " #" + gwCode
    object.Description      = "Generic MQTT Device "+ nameHint  + " #" + gwCode

    objectId, err := this.pg.CreateObject(object)
    if err != nil {
        pmlog.LogError("error create new generic device:", err)
        return topicBase, patternDefined , err
    }

    _, err = this.pg.UpdateObjectPropertyByName(objectId, mqttPropertyBridgeObjectIdName, this.objectId)
    if err != nil {
        pmlog.LogInfo("error update message property:", err)
    }

    _, err = this.pg.UpdateObjectPropertyByName(objectId, mqttPropertyTopicBaseName, topicBase)
    if err != nil {
        pmlog.LogInfo("error update message property:", err)
    }
    pmlog.LogInfo("application created new generic device", objectId)
    return topicBase, patternDefined , err
}
//
// BridgeApp: Piblish()
//
func (this *Application) Publish(topic string, payload string) error {
    var err error
    err = this.tr.Publish(topic, payload)
    if err != nil {
        return err
    }
    return err
}
//
//
const (
    testReportMessage string = "module test successful"
)
func (this *Application) RouteControlMessage(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    err = this.pg.UpdateControlExecutionAck(controlMessage.Id)
    if err != nil {
        return err
    }
    switch  controlMessage.Name {
        case controlPublishName:
            this.PublishController(controlMessage)

        case controlReloadName:
            this.ReloadController(controlMessage)

        case controlTestModuleName:
            this.LogController(controlMessage)

        case controlSetTopicsName:
            this.SetTopicsController(controlMessage)

        case controlSetBrokerURLName:
            this.SetBrokerURLController(controlMessage)

        case controlSetUsernameName:
            this.SetUsernameController(controlMessage)

        case controlSetPasswordName:
            this.SetPasswordController(controlMessage)

        case controlSetAutoProvisionName:
            this.SetAutoProvisionController(controlMessage)

        default:
            pmlog.LogError("*** unknown control message:", controlMessage.GetJSON())
            err = errors.New(fmt.Sprintf("unknown control message name %s", controlMessage.Name))
    }
    //if err != nil {
    //    pmlog.LogError("*** control error:", err)
    //}
    err = this.pg.CreateControlExecutionEmptyReport(controlMessage.Id, false, true)
    if err != nil {
        return err
    }
    
    return err
}
func (this *Application) SetTopicsController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := UnpackTopicsArguments(controlMessage.Params)
    pmlog.LogInfo("set topics:", arguments.Topics)
    _, err = this.pg.UpdateObjectPropertyByName(this.objectId, mqttPropertyTopicsName, arguments.Topics)
    if err != nil {
        return err
    }
    err = this.TransportRestart()
    if err != nil {
        return err
    }
    return err
}
func (this *Application) SetBrokerURLController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := UnpackBrokerURLArguments(controlMessage.Params)
    pmlog.LogInfo("set broker url:", arguments.BrokerURL)
    _, err = this.pg.UpdateObjectPropertyByName(this.objectId, mqttPropertyBrokerUrlName, arguments.BrokerURL)
    if err != nil {
        return err
    }
    err = this.TransportRestart()
    if err != nil {
        return err
    }
    return err
}
func (this *Application) SetUsernameController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := UnpackUsernameArguments(controlMessage.Params)
    pmlog.LogInfo("set username:", arguments.Username)
    _, err = this.pg.UpdateObjectPropertyByName(this.objectId, mqttPropertyUsernameName, arguments.Username)
    if err != nil {
        return err
    }
    err = this.TransportRestart()
    if err != nil {
        return err
    }
    return err
}
func (this *Application) SetPasswordController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := UnpackPasswordArguments(controlMessage.Params)
    pmlog.LogInfo("set password:", arguments.Password)
    _, err = this.pg.UpdateObjectPropertyByName(this.objectId, mqttPropertyPasswordName, arguments.Password)
    if err != nil {
        return err
    }
    err = this.TransportRestart()
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *Application) SetAutoProvisionController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    arguments, _ := pgcore.UnpackBoolArguments(controlMessage.Params)
    enableFlag, err := strconv.ParseBool(arguments.Enable)
    if err != nil {
        return err
    }

    this.SetAutoProvision(enableFlag)

    pmlog.LogInfo("set auto provision:", enableFlag)
    _, err = this.pg.UpdateObjectPropertyByName(this.objectId, propertyAutoProvisionName, strconv.FormatBool(enableFlag))
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *Application) LogController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    pmlog.LogInfo("*** log controller message:", controlMessage.GetJSON())
    return err
}
//
//
func (this *Application) ReloadController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error
    err = this.ReStartApplication()
    if err != nil {
        return err
    }
    return err
}
//
//
func (this *Application) PublishController(controlMessage pgcore.ControlExecutionMessage) error {
    var err error

    pmlog.LogDetail("*** publish controller full control message:", controlMessage.GetJSON())
    pmlog.LogDetail("*** publish controller message params:", controlMessage.Params)

    arguments, _ := pgcore.UnpackPublishArguments(controlMessage.Params)
    if err != nil {
        return err
    }
    pmlog.LogDetail("*** publish controller sent message with topic:", arguments.TopicName, "payload:", string(arguments.Payload))

    err = this.tr.Publish(arguments.TopicName, string(arguments.Payload))
    if err != nil {
        return err
    }
    return err
}

//
//*********************************************************************//
//
type TopicsArguments struct {
    Topics   string      `json:"topics"`
}

func NewTopicsArguments() *TopicsArguments {
    var arguments TopicsArguments
    return &arguments
}

func UnpackTopicsArguments(jsonString string) (*TopicsArguments, error) {
    var err error
    var arguments TopicsArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *TopicsArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return pgcore.Escape(string(jsonBytes))
}

func (this *TopicsArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}

//*********************************************************************//
//
//
type BrokerURLArguments struct {
    BrokerURL   string      `json:"hostname"`
}

func NewBrokerURLArguments() *BrokerURLArguments {
    var arguments BrokerURLArguments
    return &arguments
}

func UnpackBrokerURLArguments(jsonString string) (*BrokerURLArguments, error) {
    var err error
    var arguments BrokerURLArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *BrokerURLArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return pgcore.Escape(string(jsonBytes))
}

func (this *BrokerURLArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}


//*********************************************************************//
//
//
type UsernameArguments struct {
    Username   string      `json:"username"`
}

func NewUsernameArguments() *UsernameArguments {
    var arguments UsernameArguments
    return &arguments
}

func UnpackUsernameArguments(jsonString string) (*UsernameArguments, error) {
    var err error
    var arguments UsernameArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *UsernameArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return pgcore.Escape(string(jsonBytes))
}

func (this *UsernameArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}

//
//*********************************************************************//
//
type PasswordArguments struct {
    Password   string      `json:"password"`
}

func NewPasswordArguments() *PasswordArguments {
    var arguments PasswordArguments
    return &arguments
}

func UnpackPasswordArguments(jsonString string) (*PasswordArguments, error) {
    var err error
    var arguments PasswordArguments
    err = json.Unmarshal([]byte(jsonString), &arguments)
    return &arguments, err
}

func (this *PasswordArguments) Pack() string {
    jsonBytes, _ := json.Marshal(this)
    return pgcore.Escape(string(jsonBytes))
}

func (this *PasswordArguments) GetJSON() string {
    jsonBytes, _ := json.Marshal(this)
    return string(jsonBytes)
}
//EOF
