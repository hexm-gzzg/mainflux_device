// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sensor

import (
	"github.com/mainflux/senml"
)

type Sensor struct {
	SensorName string
	pack senml.Pack
	format senml.Format
}

// New ...
func New(sensorName string, format senml.Format) Sensor {
	return Sensor {
		SensorName : sensorName,
		pack : senml.Pack{},
		format: format,
	}
}
// NewRecord  ...
func (sens *Sensor)NewRecord() *senml.Record {
	return &senml.Record {
		BaseName: sens.SensorName,
	}
}
// FillSenmelMessage ...
func (sens *Sensor) FillSenmelMessage(r senml.Record) error{
	sens.pack.Records = append(sens.pack.Records,r)
	return nil
}
// EncodeSenmelMessage ...
func (sens *Sensor) EncodeSenmelMessage() ([]byte,error){
	bs,err := senml.Encode(sens.pack,sens.format)
	return bs,err
}