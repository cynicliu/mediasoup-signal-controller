package service

import (
	"encoding/json"
	"mediasoup-signal-controller/common"
	"mediasoup-signal-controller/rtp"

	"github.com/cloudwebrtc/go-protoo/logger"
)

type ConsumerProperty struct {
	kind          string
	rtpParameters rtp.ClientRtpParameters
	consumerType  string
}

type Consumer struct {
	internal       common.ConsumerInternal
	data           ConsumerProperty
	channel        *Channel
	payloadChannel *PayloadChannel
	peer           *PeerWrapper
	appData        interface{}
	paused         bool
	producerPaused bool
	score          int
	router         *Router
}

func (consumer *Consumer) HandleNotification(id string, msg common.ChannelMessage) {
	switch msg.Event {
	case "transportclose":
		break
	case "producerclose":
		break
	case "producerpause":
		break
	case "producerresume":
		break
	case "score":
		consumer.peer.peer.Notify("consumerScore",
			struct {
				Id    string          `json:"consumerId"`
				Score json.RawMessage `json:"score"`
			}{
				Id:    id,
				Score: msg.Data,
			})
		break
	case "trace":
		break
	case "layerschange":

		//{"spatialLayer":2,"temporalLayer":0}
		type LLChanged struct {
			SpatialLayer  int `json:"spatialLayer"`
			TemporalLayer int `json:"temporalLayer"`
		}

		var llc LLChanged
		_ = json.Unmarshal(msg.Data, &llc)

		logger.Debugf("=============Consumer notify:consumerLayersChanged:id=%s,spatial=%d,temporal=%d", id, llc.SpatialLayer, llc.TemporalLayer)
		consumer.peer.peer.Notify("consumerLayersChanged",
			struct {
				Id            string `json:"consumerId"`
				SpatialLayer  int    `json:"spatialLayer"`
				TemporalLayer int    `json:"temporalLayer"`
			}{
				Id:            id,
				SpatialLayer:  llc.SpatialLayer,
				TemporalLayer: llc.TemporalLayer,
			})
		break
	default:
		logger.Errorf("ignoring unknown event: %s", msg.Event)
		break
	}
}

func (consumer *Consumer) getStats() json.RawMessage {
	return consumer.router.getConsumerStats(&consumer.internal)
}
