package service

import (
	"encoding/json"
	"mediasoup-signal-controller/common"
)

type DataProducerProperty struct {
	DataProduceType      string
	sctpStreamParameters common.SctpStreamParameters_t
	label                string
	protocol             string
}

type DataProducer struct {
	internal       common.DataProducerInternal
	data           *DataProducerProperty
	channel        *Channel
	PayloadChannel *PayloadChannel
	AppData        json.RawMessage
}

func (dataProducer *DataProducer) Id() string {
	return dataProducer.internal.DataProducerId
}

func (dataProducer *DataProducer) Label() string {
	return dataProducer.data.label
}
