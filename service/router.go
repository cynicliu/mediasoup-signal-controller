package service

import (
	"encoding/json"
	"mediasoup-signal-controller/common"
	"mediasoup-signal-controller/conf"
	"mediasoup-signal-controller/rtp"
	"sync"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/go-basic/uuid"
)

type Router struct {
	Waiting_Response
	rom             *Room
	channel         *Channel
	payloadChannel  *PayloadChannel
	internal        *common.Internal_t
	data            interface{}
	appData         interface{}
	rtpCapabilities rtp.RtpCapabilities

	producers     map[string]*Producer
	dataProducers map[string]*DataProducer
	transports    map[string]interface{}

	channelRecvChan chan common.ChannelMessage
}

func CreateNewRouter(rom *Room, mediaCodecs conf.RouterOptions_t, internal *common.Internal_t, data interface{}, channel *Channel, payloadChannel *PayloadChannel, appData interface{}) *Router {

	rtpCapabilities := rtp.GenerateRouterRtpCapabilities(mediaCodecs)
	logger.Debugf("rtpCapabilities========%+v", *rtpCapabilities)

	router := &Router{
		channel:         channel,
		rom:             rom,
		payloadChannel:  payloadChannel,
		internal:        internal,
		data:            data,
		appData:         appData,
		rtpCapabilities: *rtpCapabilities,
		Waiting_Response: Waiting_Response{
			waitId: -1,
		},
	}
	router.producers = make(map[string]*Producer)
	router.dataProducers = make(map[string]*DataProducer)
	router.transports = make(map[string]interface{})
	router.channelRecvChan = make(chan common.ChannelMessage)

	return router
}

func (router *Router) CreateWebRtcTransport(wrto *common.WebRtcTransportOptions) *WebRtcTransport {
	var wrtr common.WebRtcTransport_ReqData

	if wrto.WebRtcTransportOptions.InitialAvailableOutgoingBitrate == 0 {
		wrto.WebRtcTransportOptions.InitialAvailableOutgoingBitrate = 600000
	}

	if wrto.WebRtcTransportOptions.MaxSctpMessageSize == 0 {
		wrto.WebRtcTransportOptions.MaxSctpMessageSize = 262144
	}

	if wrto.SctpSendBufferSize == 0 {
		wrto.SctpSendBufferSize = 262144
	}

	ips := make([]interface{}, 0)
	for _, value := range wrto.WebRtcTransportOptions.ListenIps {
		if len(value.Ip) > 0 && len(value.AnnouncedIp) > 0 {
			var ip common.ListenIp_t
			ip.Ip = wrto.WebRtcTransportOptions.ListenIps[0].Ip
			ip.AnnouncedIp = value.AnnouncedIp
			ips = append(ips, ip)
		} else if len(value.Ip) > 0 {
			var ip common.IP
			ip.ListenIp = value.Ip
			ips = append(ips, ip)
		}
	}
	wrtr.ListenIp, _ = json.Marshal(&ips)

	wrtr.EnableUdp = wrto.EnableUdp
	wrtr.EnableTcp = wrto.EnableTcp
	wrtr.PreferTcp = wrto.PreferTcp
	wrtr.PreferUdp = wrto.PreferUdp
	wrtr.InitialAvailableOutgoingBitrate = wrto.WebRtcTransportOptions.InitialAvailableOutgoingBitrate
	wrtr.EnableSctp = wrto.EnableSctp
	wrtr.NumSctpStreams = wrto.NumSctpStreams
	wrtr.MaxSctpMessageSize = wrto.WebRtcTransportOptions.MaxSctpMessageSize
	wrtr.SctpSendBufferSize = wrto.SctpSendBufferSize
	wrtr.IsDataChannel = true

	var transport *WebRtcTransport
	transport = nil

	var wg sync.WaitGroup
	wg.Add(1)
	go func(wrtr *common.WebRtcTransport_ReqData) {

		internal := &common.RTCTransportInternal{
			Internal_t:  *router.internal,
			TransportId: uuid.New(),
		}
		_, err := router.channel.Request("router.createWebRtcTransport", internal, wrtr, router,

			func(result common.ChannelMessage) {
				logger.Infof("router.createWebRtcTransport success: =>  %d", result.Id)

				transport = createWebRtcTransport(internal, result.Data, router.channel, router.payloadChannel, nil, 0, 0, router, wrto.AppData)

				router.transports[internal.TransportId] = transport

				wg.Done()
			},
			func(code int, err string) {
				logger.Infof("router.createWebRtcTransport reject: %d => %s", code, err)

				wg.Done()
			})

		if err != nil {
			logger.Errorf("router.createWebRtcTransport send failed:%s", err.Error())
			wg.Done()
		}
	}(&wrtr)

	wg.Wait()

	transport.channel.AddListener(transport.internal.TransportId, transport)
	return transport
}

func (router *Router) Connect(dtlspd *common.DtlsParametersData, internal *common.RTCTransportInternal) string {

	var drf common.DtlsRoleFB
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		_, err := router.channel.Request("transport.connect", internal, dtlspd, router,

			func(result common.ChannelMessage) {
				logger.Infof("transport.connect success: =>  %d", result.Id)

				_ = json.Unmarshal(result.Data, &drf)

				wg.Done()
			},
			func(code int, err string) {
				logger.Infof("transport.connect reject: %d => %s", code, err)

				wg.Done()
			})
		if err != nil {
			logger.Errorf("transport.connect send failed:%s", err.Error())
			wg.Done()
		}
	}()

	wg.Wait()
	return drf.DtlsLocalRole
}

func (router *Router) Produce(pd *common.ProducerData, internal *common.ProducerInternal) string {

	var pf common.ProduceFB
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		_, err := router.channel.Request("transport.produce", internal, pd, router,
			func(result common.ChannelMessage) {
				logger.Infof("transport.produce success: =>  %d", result.Id)

				_ = json.Unmarshal(result.Data, &pf)

				wg.Done()
			},
			func(code int, err string) {
				logger.Infof("transport.produce reject: %d => %s", code, err)

				wg.Done()
			})

		if err != nil {
			logger.Errorf("transport.produce send failed:%s", err.Error())
			wg.Done()
		}
	}()

	wg.Wait()
	return pf.Type
}

func (router *Router) ProduceData(pd *common.DataProducerData, internal *common.DataProducerInternal) *common.DataProduceFB {

	var dpf common.DataProduceFB
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		_, err := router.channel.Request("transport.produceData", internal, pd, router,

			func(result common.ChannelMessage) {
				logger.Infof("=================transport.produceData success: =>  %d======================", result.Id)

				_ = json.Unmarshal(result.Data, &dpf)

				wg.Done()
			},
			func(code int, err string) {
				logger.Infof("==============transport.produceData reject: %d => %s===============", code, err)

				wg.Done()
			})

		if err != nil {
			logger.Errorf("====================transport.produceData send failed:%s================", err.Error())
			wg.Done()
		}
	}()

	wg.Wait()
	return &dpf
}

func (router *Router) Consume(consumerData *common.ConsumeData, internal *common.ConsumerInternal) common.ConsumeFB {

	var cf common.ConsumeFB
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		_, err := router.channel.Request("transport.consume", internal, consumerData, router,

			func(result common.ChannelMessage) {
				logger.Infof("transport.consume success: =>  %d", result.Id)

				_ = json.Unmarshal(result.Data, &cf)

				wg.Done()
			},
			func(code int, err string) {
				logger.Infof("transport.consume reject: %d => %s", code, err)

				wg.Done()
			})
		if err != nil {
			logger.Errorf("transport.consume send failed:%s", err.Error())
			wg.Done()
		}
	}()

	wg.Wait()
	return cf
}

func (router *Router) getTransportStats(internal *common.RTCTransportInternal) json.RawMessage {

	var fb json.RawMessage
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		_, err := router.channel.Request("transport.getStats", internal, nil, router,

			func(result common.ChannelMessage) {
				logger.Infof("transport.getStats success: =>  %d", result.Id)

				fb = result.Data

				wg.Done()
			},
			func(code int, err string) {
				logger.Infof("transport.getStats reject: %d => %s", code, err)

				wg.Done()
			})
		if err != nil {
			logger.Errorf("transport.getStats send failed:%s", err.Error())
			wg.Done()
		}
	}()

	wg.Wait()
	return fb
}

func (router *Router) getProducerStats(internal *common.ProducerInternal) json.RawMessage {

	var fb json.RawMessage
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		_, err := router.channel.Request("producer.getStats", internal, nil, router,

			func(result common.ChannelMessage) {
				logger.Infof("producer.getStats success: =>  %d", result.Id)

				fb = result.Data

				wg.Done()
			},
			func(code int, err string) {
				logger.Infof("producer.getStats reject: %d => %s", code, err)

				wg.Done()
			})
		if err != nil {
			logger.Errorf("producer.getStats send failed:%s", err.Error())
			wg.Done()
		}
	}()

	wg.Wait()
	return fb
}

func (router *Router) getConsumerStats(internal *common.ConsumerInternal) json.RawMessage {

	var fb json.RawMessage
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		_, err := router.channel.Request("consumer.getStats", internal, nil, router,

			func(result common.ChannelMessage) {
				logger.Infof("consumer.getStats success: =>  %d", result.Id)

				fb = result.Data

				wg.Done()
			},
			func(code int, err string) {
				logger.Infof("consumer.getStats reject: %d => %s", code, err)

				wg.Done()
			})
		if err != nil {
			logger.Errorf("consumer.getStats send failed:%s", err.Error())
			wg.Done()
		}
	}()

	wg.Wait()
	return fb
}

func (router *Router) Notify(rom *Room, producerID string, data json.RawMessage) {

	var resp common.ScoreDataResp
	resp.ProducerId = producerID
	resp.Score = data

	rom.NotifyByProducer("producerScore", producerID, resp)
}

func (router *Router) OnChannelMessage(msg common.ChannelMessage) {
	logger.Debugf("------------------router receive accepted message:id=%d, method:%t", msg.Id, msg.Accepted)
	router.channelRecvChan <- msg
}

func (router *Router) GetProducerbyId(id string) *Producer {
	return router.producers[id]
}

func (router *Router) GetDataProducerbyId(id string) *DataProducer {
	return router.dataProducers[id]
}

func (router *Router) OnTransportClose() {}

func (router *Router) OnNewProducer(producer *Producer) {
	router.producers[producer.id] = producer
}

func (router *Router) OnProducerClose(producer *Producer) {
	router.producers[producer.id] = nil
}

func (router *Router) OnNewDataProducer(producer *DataProducer) {
	router.dataProducers[producer.internal.DataProducerId] = producer
}

func (router *Router) OnDataProducerClose(producer *DataProducer) {
	router.producers[producer.internal.DataProducerId] = nil
}

func (router *Router) HandleNotification(id string, msg common.ChannelMessage) {
	switch msg.Event {
	case "layerschange":

		break
	default:
		break
	}
}
