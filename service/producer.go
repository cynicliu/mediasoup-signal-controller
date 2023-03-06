package service

import (
	"encoding/json"
	"mediasoup-signal-controller/common"
	"mediasoup-signal-controller/rtp"
)

type ProducerProperty struct {
	kind                    string
	rtpParameters           rtp.ClientRtpParameters
	producerType            string
	consumableRtpParameters rtp.ClientRtpParameters
}
type Producer struct {
	id             string
	internal       common.ProducerInternal
	channel        *Channel
	payloadChannel *PayloadChannel
	appData        interface{}
	paused         bool
	data           *ProducerProperty
	router         *Router
	PeerInfo       *PeerWrapper
}

func CreateNewProducer() *Producer {
	producer := &Producer{}

	return producer
}

func (producer *Producer) Id() string {
	return producer.internal.ProducerId
}

func (producer *Producer) OnNotify(method string, data interface{}) {

	producer.PeerInfo.peer.Notify(method, data)
}

func (producer *Producer) HandleNotification(id string, msg common.ChannelMessage) {

	switch msg.Event {
	case "score":
		producer.PeerInfo.peer.Notify("producerScore",
			struct {
				Id   string          `json:"consumerId"`
				Data json.RawMessage `json:"score"`
			}{
				Id:   id,
				Data: msg.Data,
			})
		break
	default:
		break
	}
}

func (producer *Producer) getStats() json.RawMessage {

	return producer.router.getProducerStats(&producer.internal)
}
