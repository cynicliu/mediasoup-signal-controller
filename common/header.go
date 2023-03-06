package common

import (
	"encoding/json"
	"mediasoup-signal-controller/conf"
	"mediasoup-signal-controller/rtp"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = 5 * time.Second //(pongWait * 8) / 10

	// Maximum message size allowed from peer.
	MaxMessageSize = 131072 //128k
)

type SMHandler interface {
	OnTimeout()
}

type SendMessage struct {
	Id     int
	Method string
}

type ChannelMessage struct {
	Id        int             `json:"id"`
	ErrorInfo string          `json:"error"`
	Reason    string          `json:"reason"`
	TargetId  string          `json:"targetId"`
	Event     string          `json:"event"`
	Accepted  bool            `json:"accepted"`
	Data      json.RawMessage `json:"data"`
}

type ListenIp_t struct {
	Ip          string `json:"ip"`
	AnnouncedIp string `json:"announcedIp"`
}

type NumStreams_t struct {
	OS  int `json:"OS"`
	MIS int `json:"MIS"`
}

type SctpCapabilities_t struct {
	NumStreams NumStreams_t `json:numStreams`
}

type CmdMessage struct {
	ForceTcp         bool               `json:"forceTcp"`
	Producing        bool               `json:"producing"`
	Consuming        bool               `json:"consuming"`
	SctpCapabilities SctpCapabilities_t `json:"sctpCapabilities"`
}

type TransportAppData struct {
	Consuming bool
	Producing bool
}

type WebRtcTransportOptions struct {
	WebRtcTransportOptions conf.WebRtcTransportOptions_t
	EnableSctp             bool
	NumSctpStreams         NumStreams_t
	AppData                TransportAppData
	EnableUdp              bool
	EnableTcp              bool
	PreferUdp              bool
	PreferTcp              bool
	SctpSendBufferSize     int
}

type IP struct {
	ListenIp string `json:"ip"`
}

type WebRtcTransport_ReqData struct {
	ListenIp                        json.RawMessage `json:"listenIps"`
	EnableUdp                       bool            `json:"enableUdp"`
	EnableTcp                       bool            `json:"enableTcp"`
	PreferUdp                       bool            `json:"preferUdp"`
	PreferTcp                       bool            `json:"preferTcp"`
	InitialAvailableOutgoingBitrate int             `json:"initialAvailableOutgoingBitrate"`
	EnableSctp                      bool            `json:"enableSctp"`
	NumSctpStreams                  NumStreams_t    `json:"numSctpStreams"`
	MaxSctpMessageSize              int             `json:"maxSctpMessageSize"`
	SctpSendBufferSize              int             `json:"sctpSendBufferSize"`
	IsDataChannel                   bool            `json:"isDataChannel"`
}

type Internal_t struct {
	RouterId string `json:"routerId"`
}

type RTCTransportInternal struct {
	Internal_t
	TransportId string `json:"transportId"`
}

type ProducerInternal struct {
	RTCTransportInternal
	ProducerId string `json:"producerId"`
}

type DataProducerInternal struct {
	RTCTransportInternal
	DataProducerId string `json:"dataProducerId"`
}

type ConsumerInternal struct {
	RTCTransportInternal
	ConsumerId string `json:"consumerId"`
	ProducerId string `json:"producerId"`
}

type Request_t struct {
	Id       int         `json:"id"`
	Method   string      `json:"method"`
	Internal interface{} `json:"internal"`
	Data     interface{} `json:"data"`
}

////////////////////////////////////////////////////
type Fingerprint_t struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

type DtlsParameter_t struct {
	Role         string          `json:"role"`
	Fingerprints []Fingerprint_t `json:"fingerprints"`
}

type IceParameter_t struct {
	IceLite          bool   `json:"iceLite"`
	Password         string `json:"password"`
	UsernameFragment string `json:"usernameFragment"`
}

type IceCandidate_t struct {
	Foundation string `json:"foundation"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Priority   int    `json:"priority"`
	Protocol   string `json:"protocol"`
	Type       string `json:"type"`
}

type IceSelectedTuple_t struct {
	LocalIp    string `json:"localIp"`
	LocalPort  int    `json:"localPort"`
	RemoteIp   string `json:"remoteIp"`
	RemotePort int    `json:"remotePort"`
	Protocal   string `json:"protocal"`
}

type SctpParameter_t struct {
	MIS                int  `json:"MIS"`
	OS                 int  `json:"OS"`
	IsDataChannel      bool `json:"isDataChannel"`
	MaxMessageSize     int  `json:"maxMessageSize"`
	Port               int  `json:"port"`
	SctpBufferedAmount int  `json:"sctpBufferedAmount"`
	SendBufferSize     int  `json:"sendBufferSize"`
}

/*
	{
		"consumerIds": [],
		"dataConsumerIds": [],
		"dataProducerIds": [],
		"direct": false,
		"dtlsParameters": {
			"fingerprints": [{
				"algorithm": "sha-1",
				"value": "23:60:BB:B8:E1:AD:82:2E:61:61:AE:19:3C:41:6A:9D:FD:29:80:45"
			}, {
				"algorithm": "sha-224",
				"value": "7B:D9:23:56:BA:60:3F:D5:CF:40:67:67:0B:17:E3:47:25:F4:7F:97:DA:8A:5A:75:59:E2:21:39"
			}, {
				"algorithm": "sha-256",
				"value": "CE:19:28:A5:87:86:38:6F:DA:1A:FE:0D:D2:BE:7C:F8:42:D7:39:88:15:5B:28:46:E3:19:61:3C:B0:09:2A:01"
			}, {
				"algorithm": "sha-384",
				"value": "17:1F:F1:E1:B3:8F:43:3E:A6:24:60:EF:47:E3:B5:7E:A4:93:99:A3:0B:14:60:A6:6D:8D:56:91:CE:D6:E7:2C:92:28:93:38:F9:36:A7:F7:BA:8F:99:13:69:A0:7F:65"
			}, {
				"algorithm": "sha-512",
				"value": "1F:99:8E:13:48:5B:97:92:AE:0E:9B:A5:A4:F7:75:31:89:4C:B5:7B:F5:12:B8:07:0A:16:CD:38:23:A4:BC:07:61:78:B9:FC:22:78:89:7B:00:93:22:BE:72:51:97:61:EA:1F:64:71:FA:77:62:13:36:F2:FE:97:DA:B2:5A:AE"
			}],
			"role": "auto"
		},
		"dtlsState": "new",
		"iceCandidates": [{
			"foundation": "udpcandidate",
			"ip": "10.226.156.132",
			"port": 45336,
			"priority": 1076302079,
			"protocol": "udp",
			"type": "host"
		}],
		"iceParameters": {
			"iceLite": true,
			"password": "nx87pe8mo70dahl1sqt4mb1blk16j26u",
			"usernameFragment": "mcna81rxn2v0x48c"
		},
		"iceRole": "controlled",
		"iceState": "new",
		"id": "5a571985-4a1d-81bb-998a-f44e2d055f28",
		"mapRtxSsrcConsumerId": {},
		"mapSsrcConsumerId": {},
		"maxMessageSize": 262144,
		"producerIds": [],
		"recvRtpHeaderExtensions": {},
		"rtpListener": {
			"midTable": {},
			"ridTable": {},
			"ssrcTable": {}
		},
		"sctpListener": {
			"streamIdTable": {}
		},
		"sctpParameters": {
			"MIS": 1024,
			"OS": 1024,
			"isDataChannel": true,
			"maxMessageSize": 262144,
			"port": 5000,
			"sctpBufferedAmount": 0,
			"sendBufferSize": 262144
		},
		"sctpState": "new",
		"traceEventTypes": ""
	}
*/

// type WebRtcTransportData struct {
// 	IceRole          string             `json:"iceRole"`
// 	IceParameters    []IceParameter_t   `json:"iceParameters"`
// 	IceCandidates    []IceCandidate_t   `json:"iceCandidates"`
// 	IceState         string             `json:"iceState"`
// 	IceSelectedTuple IceSelectedTuple_t `json:"iceSelectedTuple"`
// 	DtlsParameters   []DtlsParameter_t  `json:"dtlsParameters"`
// 	DtlsState        string             `json:"dtlsState"`
// 	DtlsRemoteCert   string             `json:"dtlsRemoteCert"`
// 	SctpParameter    SctpParameter_t    `json:"sctpParameters"`
// 	SctpState        string             `json:"sctpState"`
// }

type WebRtcTransportData struct {
	Id               string          `json:"id"`
	IceRole          string          `json:"iceRole"`
	IceParameters    json.RawMessage `json:"iceParameters"`
	IceCandidates    json.RawMessage `json:"iceCandidates"`
	IceState         string          `json:"iceState"`
	IceSelectedTuple json.RawMessage `json:"iceSelectedTuple"`
	DtlsParameters   json.RawMessage `json:"dtlsParameters"`
	DtlsState        string          `json:"dtlsState"`
	DtlsRemoteCert   string          `json:"dtlsRemoteCert"`
	SctpParameter    json.RawMessage `json:"sctpParameters"`
	SctpState        string          `json:"sctpState"`
}

type WebRtcTransportAccept struct {
	Id             string          `json:"id"`
	IceParameters  json.RawMessage `json:"iceParameters"`
	IceCandidates  json.RawMessage `json:"iceCandidates"`
	DtlsParameters json.RawMessage `json:"dtlsParameters"`
	SctpParameter  json.RawMessage `json:"sctpParameters"`
}

////////////////////////////////////////////////////////////////////
type NilAccept struct {
}

// connectWebRtcTransport

type ConnectWebRtcTransportData struct {
	TransportId    string          `json:"transportId"`
	DtlsParameters DtlsParameter_t `json:"dtlsParameters"`
}

type DtlsParametersData struct {
	DtlsParameters DtlsParameter_t `json:"dtlsParameters"`
}

type DtlsRoleFB struct {
	DtlsLocalRole string `json:"dtlsLocalRole"`
}

// produce

type ClientProduceData struct {
	TransportId   string `json:"transportId"`
	Kind          string `json:"kind"`
	RtpParameters rtp.ClientRtpParameters
	AppData       json.RawMessage `json:"appData"`
}

type ProducerData struct {
	Kind                 string                   `json:"kind"`
	RtpParameters        *rtp.ClientRtpParameters `json:"rtpParameters"`
	RtpMapping           *rtp.RtpMapping          `json:"rtpMapping"`
	KeyFrameRequestDelay int                      `json:"keyFrameRequestDelay"`
	Paused               bool                     `json:"paused"`
}

type ProduceFB struct {
	Type string `json:"type"`
}

type ProduceResp struct {
	Id string `json:"id"`
}

// produce Data
type SctpStreamParameters_t struct {
	StreamId          int  `json:"streamId"`
	Ordered           bool `json:"ordered"`
	MaxPacketLifeTime int  `json:"maxPacketLifeTime"`
	MaxRetransmits    int  `json:"maxRetransmits"`
}

type ProduceDataData struct {
	TransportId          string                 `json:"transportId"`
	SctpStreamParameters SctpStreamParameters_t `json:"sctpStreamParameters"`
	Label                string                 `json:"label"`
	Protocol             string                 `json:"protocol"`
	AppData              json.RawMessage        `json:"appData"`
}

type DataProducerData struct {
	Type                 string                 `json:"type"`
	SctpStreamParameters SctpStreamParameters_t `json:"sctpStreamParameters"`
	Label                string                 `json:"label"`
	Protocol             string                 `json:"protocol"`
}

type DataProduceResp struct {
	Id string `json:"id"`
}

type DataProduceFB struct {
	Id                   string                 `json:"id"`
	DataType             string                 `json:"type"`
	Label                string                 `json:"label"`
	Protocol             string                 `json:"protocol"`
	SctpStreamParameters SctpStreamParameters_t `json:"sctpStreamParameters"`
}

////////////////////////////////////////
// consumer
type ConsumeData struct {
	Kind                   string                      `json:"kind"`
	RtpParameters          *rtp.ClientRtpParameters    `json:"rtpParameters"`
	ConsumerType           string                      `json:"type"`
	ConsumableRtpEncodings []rtp.RtpEncodingParameters `json:"consumableRtpEncodings"`
	Paused                 bool                        `json:"paused"`
}

type ConsumeFB struct {
	Paused          bool `json:"paused"`
	ProducerPaused  bool `json:"producerPaused"`
	Score           int  `json:"score"`
	PreferredLayers int  `json:"preferredLayers"`
}

type NewConsumerAppData struct {
	PeerId string `json:"peerId"`
}
type NewConsumerData struct {
	PeerId         string                  `json:"peerId"`
	ProducerId     string                  `json:"producerId"`
	Id             string                  `json:"id"`
	Kind           string                  `json:"kind"`
	RtpParameters  rtp.ClientRtpParameters `json:"rtpParameters"`
	ConsumerType   string                  `json:"type"`
	AppData        NewConsumerAppData      `json:"appData"`
	ProducerPaused bool                    `json:"producerPaused"`
}

////////////////////////////
//data consumer
type DataConsumer struct {
	DataConsumerId string `json:"dataConsumerId"`
}

type DataProducer struct {
	DataProducerId string `json:"dataProducerId"`
}

// score data
// {
// 	"encodingIdx": 1,
// 	"rid": "r1",
// 	"score": 10,
// 	"ssrc": 3580009065
// }

type ScoreData struct {
	EncodingIdx int    `json:"encodingIdx"`
	Rid         string `json:"rid"`
	Score       int    `json:"score"`
	Ssrc        int    `json:"ssrc"`
}

type ScoreDataResp struct {
	ProducerId string          `json:"producerId"`
	Score      json.RawMessage `json:"score"`
}
