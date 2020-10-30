// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"fmt"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/device/pkg/config"
	logger "github.com/mainflux/mainflux/logger"
)
const (
	heartbeatSubject = "heartbeat"
	service          = "service"
)


type Devicer struct {
	id        string
	mqtt      mqtt.Client
	cfg       config.Config
	logger    logger.Logger
	connected chan bool
	status    uint32
	sync.RWMutex
}

// New create new instance of device service
func New(c config.Config, l logger.Logger, mqttHost string, serviceName string) (*Devicer,error) {
	// id := fmt.Sprintf("device-%s", c.MQTT.Username)

	e := Devicer{
		id:        serviceName,
		logger:    l,
		cfg:       c,
	}
	client, err := e.mqttConnect(c, l, mqttHost)
	if err != nil {
		return &e, err
	}
	e.mqtt = client
	
	return &e, nil
}
// pu
func (d *Devicer) Publish(sub_topic string, payload []byte) error {
	topic := "channels/" + d.cfg.BoardCfg.ExportChannelID + "/messages/services/" + d.id + "/" + sub_topic
	token := d.mqtt.Publish(topic, byte(d.cfg.MQTT.QoS), d.cfg.MQTT.Retain, payload)
	if token.Wait() && token.Error() != nil {
		d.logger.Error(fmt.Sprintf("Failed to publish to topic %s", topic))
		return token.Error()
	}
	return nil
}
// HeratbeatPublish
func (d *Devicer) HeratbeatPublish() error {
	topic := "channels/" + d.cfg.BoardCfg.ControlChannelID + "/messages/" + heartbeatSubject + "/" + d.id + "/" + service
	fmt.Printf("to %s\n", topic)
	token := d.mqtt.Publish(topic, byte(d.cfg.MQTT.QoS), d.cfg.MQTT.Retain, []byte{})
	if token.Wait() && token.Error() != nil {
		d.logger.Error(fmt.Sprintf("Failed to publish to topic %s", topic))
		return token.Error()
	}
	return nil
}
func (d *Devicer) handleMsg(mc mqtt.Client, msg mqtt.Message) {
	fmt.Printf("received mqtt msg\n")
	fmt.Println(msg)
}

func (d *Devicer) Subscribe(sub_topic string) error {
	topic := "channels/" + d.cfg.BoardCfg.ControlChannelID + "/messages/services/" + d.id + "/#"
	d.mqtt.Subscribe(topic, byte(d.cfg.MQTT.QoS), d.handleMsg)
	
	return nil
}

func (d *Devicer) mqttConnect(conf config.Config, logger logger.Logger, mqttHost string) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(mqttHost).
		SetClientID(d.id).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetOnConnectHandler(d.conn).
		SetConnectionLostHandler(d.lost)
	
	opts.SetUsername(conf.BoardCfg.ThingID)
	opts.SetPassword(conf.BoardCfg.ThingKey)
	// if conf.MQTT.MTLS {
	// 	cfg := &tls.Config{
	// 		InsecureSkipVerify: conf.MQTT.SkipTLSVer,
	// 	}

	// 	if conf.MQTT.CA != nil {
	// 		cfg.RootCAs = x509.NewCertPool()
	// 		cfg.RootCAs.AppendCertsFromPEM(conf.MQTT.CA)
	// 	}
	// 	if conf.MQTT.TLSCert.Certificate != nil {
	// 		cfg.Certificates = []tls.Certificate{conf.MQTT.TLSCert}
	// 	}

	// 	cfg.BuildNameToCertificate()
	// 	opts.SetTLSConfig(cfg)
	// 	opts.SetProtocolVersion(4)
	// }
	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()

	if token.Error() != nil {
		logger.Error(fmt.Sprintf("Client %s had error connecting to the broker: %v", d.id, token.Error()))
		return nil, token.Error()
	}
	return client, nil
}
func (d *Devicer) conn(client mqtt.Client) {
	d.logger.Debug(fmt.Sprintf("Client %s connected", d.id))
}

func (d *Devicer) lost(client mqtt.Client, err error) {
	d.logger.Debug(fmt.Sprintf("Client %s disconnected", d.id))
}