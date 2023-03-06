package service

import (
	"mediasoup-signal-controller/common"
	"mediasoup-signal-controller/rtp"

	"github.com/cloudwebrtc/go-protoo/peer"
)

type Device_t struct {
	Flag    string `json:"flag"`
	Name    string `json:"name"`
	Version string `json:"version"`
}
type PeerInfo struct {
	Id          string   `json:"id"`
	Joined      bool     `json:"joined"`
	DisplayName string   `json:"displayName"`
	Device      Device_t `json:"device"`
}

type PeerInfos struct {
	Peers []PeerInfo `json:"peers"`
}
type PeerData struct {
	PeerInfo
	SctpCapabilities common.SctpCapabilities_t `json:"sctpCapabilities"`
	RtpCapabilities  rtp.RtpCapabilities_In    `json:"rtpCapabilities"`
}

type PeerWrapper struct {
	peer          *peer.Peer
	data          PeerData
	transports    map[string]interface{}
	producers     map[string]*Producer
	consumers     map[string]*Consumer
	dataProducers map[string]*DataProducer
	dataConsumers map[string]*DataConsumer
}
