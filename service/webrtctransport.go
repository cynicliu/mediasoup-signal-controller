package service

import (
	"encoding/json"
	"fmt"
	"mediasoup-signal-controller/common"
	"mediasoup-signal-controller/rtp"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/go-basic/uuid"
)

type WebRtcTransport struct {
	Transport

	data common.WebRtcTransportData
	role string
}

func createWebRtcTransport(internal *common.RTCTransportInternal, data json.RawMessage,
	channel *Channel, payloadChannel *PayloadChannel, rtpCapabilities interface{},
	producerId int, dataProducerId int, router *Router, appData common.TransportAppData) *WebRtcTransport {

	transport := &WebRtcTransport{
		Transport: Transport{
			router:              router,
			closed:              false,
			channel:             channel,
			payloadChannel:      payloadChannel,
			internal:            *internal,
			nextMidForConsumers: 0,
		},
	}

	transport.producers = make(map[string]*Producer)
	transport.consumers = make(map[string]*Consumer)
	transport.dataConsumers = make(map[string]*DataConsumer)
	transport.dataProducers = make(map[string]*DataProducer)

	_ = json.Unmarshal(data, &transport.data)

	// Emit observer event.
	// this._observer.safeEmit('newtransport', transport);
	return transport
}

func (wrt *WebRtcTransport) Close() {
	if wrt.closed {
		return
	}

	wrt.data.IceState = "closed"
	wrt.data.IceSelectedTuple = nil
	wrt.data.SctpState = "closed"

	wrt.channel.Request("transport.close", wrt.internal, nil, wrt.router, nil, nil)
}

func (wrt *WebRtcTransport) Id() string {
	return wrt.internal.TransportId
}

func (wrt *WebRtcTransport) connect(dtlsParameter *common.DtlsParameter_t) {
	dtlspd := &common.DtlsParametersData{
		DtlsParameters: *dtlsParameter,
	}
	role := wrt.router.Connect(dtlspd, &wrt.internal)
	wrt.role = role
}

func (wrt *WebRtcTransport) produce(prw *PeerWrapper, cpd *common.ClientProduceData, id string, paused bool, keyFrameRequestDelay int) *Producer {

	if len(id) > 0 {
		if wrt.producers[id] != nil {
			return nil
		}
	}

	if cpd.Kind != "audio" && cpd.Kind != "video" {
		return nil
	}

	routerRtpCapabilities := &wrt.router.rtpCapabilities

	logger.Debugf("router rtpCapabilities========%+v", wrt.router.rtpCapabilities)

	logger.Debugf("client request RtpParameters========%+v", cpd.RtpParameters)

	rtpMapping := rtp.GetProducerRtpParametersMapping(&cpd.RtpParameters, routerRtpCapabilities)

	logger.Debugf("rtpMapping========%+v", *rtpMapping)

	consumableRtpParameters := rtp.GetConsumableRtpParameters(cpd.Kind, &cpd.RtpParameters, routerRtpCapabilities, rtpMapping)

	logger.Debugf("consumableRtpParameters========%+v", *consumableRtpParameters)

	internal := &common.ProducerInternal{
		RTCTransportInternal: wrt.internal,
		ProducerId:           uuid.New(),
	}

	producerData := &common.ProducerData{
		Kind:                 cpd.Kind,
		RtpParameters:        &cpd.RtpParameters,
		RtpMapping:           rtpMapping,
		KeyFrameRequestDelay: keyFrameRequestDelay,
		Paused:               paused,
	}

	producerType := wrt.router.Produce(producerData, internal)

	producerProperty := &ProducerProperty{
		kind:                    cpd.Kind,
		rtpParameters:           cpd.RtpParameters,
		producerType:            producerType,
		consumableRtpParameters: *consumableRtpParameters,
	}

	producer := &Producer{
		id:             internal.ProducerId,
		data:           producerProperty,
		internal:       *internal,
		channel:        wrt.channel,
		payloadChannel: wrt.payloadChannel,
		paused:         paused,
		PeerInfo:       prw,
		router:         wrt.router,
	}

	logger.Debugf("=============WebRtcTransport add producer:%s, transport:%p", producer.id, wrt)
	wrt.producers[producer.id] = producer
	wrt.router.OnNewProducer(producer)

	wrt.channel.AddListener(producer.id, producer)

	return producer
}

func (wrt *WebRtcTransport) produceData(pdd *common.ProduceDataData) *DataProducer {

	internal := &common.DataProducerInternal{
		RTCTransportInternal: wrt.internal,
		DataProducerId:       uuid.New(),
	}

	dataProducerData := &common.DataProducerData{
		Type:                 "sctp",
		SctpStreamParameters: pdd.SctpStreamParameters,
		Label:                pdd.Label,
		Protocol:             pdd.Protocol,
	}

	data := wrt.router.ProduceData(dataProducerData, internal)

	dp := &DataProducer{
		internal: *internal,
		data: &DataProducerProperty{
			label:                data.Label,
			protocol:             data.Protocol,
			DataProduceType:      "sctp",
			sctpStreamParameters: data.SctpStreamParameters,
		},
	}

	wrt.channel.AddListener(internal.DataProducerId, dp)

	return dp
}

func (wrt *WebRtcTransport) consume(consumerPeer *PeerWrapper, producerId string, rtpCapabilities rtp.RtpCapabilities, paused bool, pipe bool) *Consumer {

	logger.Debugf("==============WebRtcTransport consume:producerID:%s, tranport:%p==============", producerId, wrt)
	producer := wrt.router.GetProducerbyId(producerId)
	if producer == nil {
		logger.Errorf("There is no producer for id:%s", producerId)
		return nil
	}

	rtpParameters := rtp.GetConsumerRtpParameters(&producer.data.consumableRtpParameters, rtpCapabilities, false)

	if rtpParameters == nil {
		logger.Debugf("WebRtcTransport consume rtpParameter is nil")
		return nil
	}

	logger.Debugf("WebRtcTransport consume rtpParameter:%+v", rtpParameters)
	if !pipe {
		rtpParameters.Mid = fmt.Sprintf("%d", wrt.nextMidForConsumers)
		wrt.nextMidForConsumers++

		if wrt.nextMidForConsumers == 100000000 {
			wrt.nextMidForConsumers = 0
		}
	}

	internal := &common.ConsumerInternal{
		RTCTransportInternal: wrt.internal,
		ConsumerId:           uuid.New(),
		ProducerId:           producerId,
	}

	consumerType := "pipe"
	if !pipe {
		consumerType = producer.data.producerType
	}

	consumerData := &common.ConsumeData{
		Kind:                   producer.data.kind,
		RtpParameters:          rtpParameters,
		ConsumableRtpEncodings: producer.data.consumableRtpParameters.Encodings,
		ConsumerType:           consumerType,
		Paused:                 paused,
	}

	status := wrt.router.Consume(consumerData, internal)

	consumerProperty := ConsumerProperty{
		kind:          producer.data.kind,
		rtpParameters: *rtpParameters,
		consumerType:  consumerType,
	}

	consumer := &Consumer{
		internal:       *internal,
		data:           consumerProperty,
		channel:        wrt.channel,
		payloadChannel: wrt.payloadChannel,
		paused:         status.Paused,
		producerPaused: status.ProducerPaused,
		score:          status.Score,
		peer:           consumerPeer,
		router:         wrt.router,
	}

	wrt.consumers[internal.ConsumerId] = consumer

	wrt.channel.AddListener(internal.ConsumerId, consumer)

	return consumer
}

func (wrt *WebRtcTransport) consumeData(dataProducerId string) *DataConsumer {
	producer := wrt.router.GetDataProducerbyId(dataProducerId)
	if producer == nil {
		logger.Errorf("There is no dataProducer for id:%s", dataProducerId)
		return nil
	}

	return nil
}

func (wrt *WebRtcTransport) getStats() json.RawMessage {

	return wrt.router.getTransportStats(&wrt.internal)
}
