// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	 "github.com/mainflux/device/pkg/config"
	exp "github.com/mainflux/device/pkg/config"
	"github.com/mainflux/device/pkg/device"
	"github.com/mainflux/device/pkg/sensor"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/senml"

)

const (
	svcName           = "device_test"
	defMqttHost       = "tcp://localhost:1883"
	defConfigFile     = "/configs/export/config.toml"
	envMqttHost = "MF_EXPORT_MQTT_HOST"
	envConfigFile = "MF_EXPORT_CONFIG_FILE"
	heartbeatSubject = "heartbeat"
	service          = "service"
	senmlBaseName = "sensorTest"
)

func main() {
	var dev *device.Devicer
	cfg, err := loadConfigs()
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println(cfg)
	logger, err := logger.New(os.Stdout, cfg.Server.LogLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}
	mqttHost := mainflux.Env(envMqttHost, defMqttHost)
	dev, err = device.New(cfg, logger, mqttHost, svcName)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create service :%s", err))
		os.Exit(1)
	}
	go func() {
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to encode senml message :%s", err))
		}

		subTopic := senmlBaseName
		for {
			sens := sensor.New(senmlBaseName, senml.JSON)
			record := sens.NewRecord()
			buildVoltageRecord(record)
			sens.FillSenmelMessage(*record)
			payload, err := sens.EncodeSenmelMessage()
			err = dev.Publish(subTopic, payload)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to Publish payload :%s", err))
				os.Exit(1)
			}
			time.Sleep(2 * time.Second)
		}
	}()
	dev.Subscribe("")

	// Publish heartbeat
	ticker := time.NewTicker(10000 * time.Millisecond)
	go func() {
		for range ticker.C {
			if err := dev.HeratbeatPublish(); err != nil {
				logger.Error(fmt.Sprintf("Failed to publish heartbeat, %s", err))
			}
		}
	}()

	errs := make(chan error, 2)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)

	}()

	err = <-errs
	logger.Error(fmt.Sprintf("export writer service terminated: %s", err))
}
func buildVoltageRecord(record *senml.Record) {
	rand.Seed(time.Now().UnixNano())
	num4 := rand.Float64()
	record.Name = "voltage"
	record.Value = &num4
	record.Unit = "v"
}

func loadConfigs() (exp.Config, error) {
	configFile := mainflux.Env(envConfigFile, defConfigFile)
	cfg, err := config.ReadFile(configFile)
	if err != nil {
		return cfg, err
	}
	mqtt, err := loadCertificate(cfg.MQTT)
	if err != nil {
		return cfg, err
	}
	cfg.MQTT = mqtt
	if cfg.BoardCfg.ThingID == "" || cfg.BoardCfg.ThingKey == "" || cfg.BoardCfg.Token == "" ||
		cfg.BoardCfg.ControlChannelID == "" || cfg.BoardCfg.ExportChannelID == "" {
		return cfg, errors.New("config file is invalid.BoardCfg : " + fmt.Sprintln(cfg.BoardCfg))
	}
	log.Println(fmt.Sprintf("Configuration loaded from file %s", configFile))
	return cfg, nil
}

func loadCertificate(cfg exp.MQTT) (exp.MQTT, error) {
	var caByte []byte
	var cc []byte
	var pk []byte
	cert := tls.Certificate{}
	if cfg.MTLS == false {
		return cfg, nil
	}

	caFile, err := os.Open(cfg.CAPath)
	if err != nil {
		return cfg, errors.New(err.Error())
	}
	defer caFile.Close()
	caByte, _ = ioutil.ReadAll(caFile)

	if cfg.ClientCertPath != "" {
		clientCert, err := os.Open(cfg.ClientCertPath)
		if err != nil {
			return cfg, errors.New(err.Error())
		}
		defer clientCert.Close()
		cc, err = ioutil.ReadAll(clientCert)
		if err != nil {
			return cfg, err
		}
	}

	if len(cc) == 0 && cfg.ClientCert != "" {
		cc = []byte(cfg.ClientCert)
	}

	if cfg.ClientPrivKeyPath != "" {
		privKey, err := os.Open(cfg.ClientPrivKeyPath)
		defer privKey.Close()
		if err != nil {
			return cfg, errors.New(err.Error())
		}
		pk, err = ioutil.ReadAll((privKey))
		if err != nil {
			return cfg, err
		}
	}

	if len(pk) == 0 && cfg.ClientCertKey != "" {
		pk = []byte(cfg.ClientCertKey)
	}

	if len(pk) == 0 || len(cc) == 0 {
		return cfg, errors.New("failed loading client certificate")
	}

	cert, err = tls.X509KeyPair([]byte(cc), []byte(pk))
	if err != nil {
		return cfg, errors.New(err.Error())
	}

	cfg.TLSCert = cert
	cfg.CA = caByte

	return cfg, nil
}
