/*
 * Copyright:  Pixel Networks <support@pixel-networks.com>
 */

package mqtrans

import (
    "errors"
    "time"

    "app/pgcore"
    "app/pmtools"
    "app/pmlog"

    mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
    keepaliveTimeout    time.Duration   = 3 // sec
    waitTimeout         time.Duration   = 3 // sec
    pingTimeout         time.Duration   = 3 // sec
    reconnectTimeout    time.Duration   = 3 // sec

    QosL1       byte = 1
    QosL2       byte = 2
    QosL3       byte = 4
)

type Transport struct {
    mc          mqtt.Client
    pg          *pgcore.Pixcore
    clientId    string
}

func NewTransport() *Transport {
    clientId := pmtools.GetNewUUID()
    return &Transport{
            clientId:   clientId,
    }
}

func (this *Transport) Bind(url string, username string, password string) error {
    var err error

    opts := mqtt.NewClientOptions()

    opts.AddBroker(url)

    opts.SetUsername(username)
    opts.SetPassword(password)
    opts.SetClientID(this.clientId)

    //opts.SetOrderMatters(true)
    opts.SetAutoReconnect(false)

    opts.SetKeepAlive(keepaliveTimeout)
    opts.SetPingTimeout(pingTimeout)

    onConnectHandler := func(client mqtt.Client) {
        pmlog.LogInfo("mqtt transport: connect to broker:", url)
    }
    opts.SetOnConnectHandler(onConnectHandler)

    onReconnectHandler := func(client mqtt.Client, opts *mqtt.ClientOptions) {
        pmlog.LogInfo("mqtt transport: reconnect to broker:", url)
        time.Sleep(reconnectTimeout * time.Second)
    }
    opts.SetReconnectingHandler(onReconnectHandler)

    this.mc = mqtt.NewClient(opts)

    token := this.mc.Connect()
    for !token.WaitTimeout(waitTimeout * time.Second) {}

    err = token.Error()
    if err != nil {
        return err
    }
    return err
}

type Handler = func(mqtt.Client, mqtt.Message)

func (this *Transport) Publish(topic string, message string) error {
    var err error
    if this.mc == nil {
        return errors.New("mqtt transport yet not exist")
    }
    token := this.mc.Publish(topic, QosL1, false, message)
    for !token.WaitTimeout(waitTimeout * time.Second) {}
    err = token.Error()
    if err != nil {
        return err
    }
    return err
}

func (this *Transport) Subscribe(topic string, handler Handler) error {
    var err error
    if this.mc == nil {
        return errors.New("mqtt transport yet not exist")
    }
    //if !this.mc.IsConnected() {}
    token := this.mc.Subscribe(topic, QosL1, handler)
    for !token.WaitTimeout(waitTimeout * time.Second) {}
    err = token.Error()
    if err != nil {
        return err
    }
    return err
}

func (this *Transport) IsConnected() bool {
    if this.mc == nil {
        return false
    }
    return this.mc.IsConnected()
}

const (
    disconnectTime uint = 1 // ms
)

func (this *Transport) Disconnect() error {
    var err error
    if this.mc != nil && this.mc.IsConnected() {
        this.mc.Disconnect(disconnectTime)
    }
    return err
}
//EOF
