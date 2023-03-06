package service

import "mediasoup-signal-controller/common"

type TransportBase interface {
	Id() string
}

type Transport struct {
	router              *Router
	closed              bool
	producers           map[string]*Producer
	consumers           map[string]*Consumer
	dataProducers       map[string]*DataProducer
	dataConsumers       map[string]*DataConsumer
	nextMidForConsumers int
	nextSctpStreamId    int
	channel             *Channel
	payloadChannel      *PayloadChannel
	internal            common.RTCTransportInternal
	appData             common.TransportAppData
}
