package service

import (
	"encoding/json"
	"mediasoup-signal-controller/common"
	"mediasoup-signal-controller/conf"
	"mediasoup-signal-controller/rtp"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/peer"
	"github.com/cloudwebrtc/go-protoo/room"
	"github.com/cloudwebrtc/go-protoo/transport"
)

type Room struct {
	cf             *conf.Config
	roomId         string
	protooRoom     *room.Room
	router         *Router
	peers          map[string]*PeerWrapper
	producerToPeer map[string]*PeerWrapper
}

func CreateNewRoom(cf *conf.Config, worker *Worker, roomId string) *Room {
	rom := &Room{
		cf:         cf,
		roomId:     roomId,
		protooRoom: nil,
		router:     nil,
	}

	rom.protooRoom = room.NewRoom(roomId)
	rom.peers = make(map[string]*PeerWrapper)
	rom.producerToPeer = make(map[string]*PeerWrapper)

	rom.router = worker.CreateRouter(rom)
	// Create a mediasoup AudioObserver

	return rom
}

func (rom *Room) CreatePeer(peerId string, transport *transport.WebSocketTransport) *peer.Peer {

	pr := rom.protooRoom.GetPeer(peerId)
	if pr != nil {
		rom.protooRoom.RemovePeer(peerId)
		pr.Close()
		logger.Warnf("there is already a protoo Peer with same peerId, closing it [peerId:%s]", peerId)
	}
	pr = peer.NewPeer(peerId, transport)

	return pr
}

func (rom *Room) Notify(from *peer.Peer, method string, data interface{}) {
	logger.Debugf("Notify data:%s,%+v=====================================", method, (data.(common.ScoreDataResp)))
	rom.protooRoom.Notify(from, method, data)
}

func (rom *Room) NotifyAll(method string, data interface{}) {
	for _, v := range rom.peers {
		rom.Notify(v.peer, method, data)
	}
}

func (rom *Room) Close() {
	rom.protooRoom.Close()
}

func (rom *Room) HandleProtooConnection(peerId string, transport *transport.WebSocketTransport) {

}

func (rom *Room) HandleProtooRequest(pr *peer.Peer, request peer.Request, accept peer.RespondFunc, reject peer.RejectFunc) {
	method := request.Method
	logger.Debugf("%s", string(request.Data))

	peerWapper := rom.peers[pr.ID()]
	if peerWapper == nil {
		peerWapper = &PeerWrapper{
			peer: pr,
		}
		rom.peers[pr.ID()] = peerWapper

		peerWapper.transports = make(map[string]interface{})
		peerWapper.producers = make(map[string]*Producer)
		peerWapper.consumers = make(map[string]*Consumer)
		peerWapper.dataProducers = make(map[string]*DataProducer)
		peerWapper.dataConsumers = make(map[string]*DataConsumer)
	}

	switch method {
	case "getRouterRtpCapabilities":
		rtpcapabilities := rom.router.rtpCapabilities

		data, _ := json.Marshal(rtpcapabilities)
		logger.Infof("%s", string(data))
		accept(rtpcapabilities)
		break
	case "join":
		peerWapper := rom.peers[pr.ID()]
		peerData := peerWapper.data
		if peerData.Joined {
			return
		}

		_ = json.Unmarshal(request.Data, &peerWapper.data)
		logger.Infof("%v", peerWapper.data)
		peerWapper.data.Id = pr.ID()
		rom.peers[pr.ID()] = peerWapper

		var peerInfos PeerInfos
		peerInfos.Peers = make([]PeerInfo, 0)
		joinedPeers := rom.getJoindPeers()

		for _, v := range joinedPeers {
			if v.data.Id != pr.ID() {
				peerInfos.Peers = append(peerInfos.Peers, PeerInfo{Id: v.data.Id, DisplayName: v.data.DisplayName, Device: v.data.Device})
			}
		}
		accept(peerInfos)
		peerWapper.data.Joined = true

		for _, joinedPeer := range joinedPeers {

			if joinedPeer.data.Id == pr.ID() {
				continue
			}
			for _, producer := range joinedPeer.producers {

				rom.CreateConsumer(peerWapper, joinedPeer, producer)
			}
		}

		for _, v := range joinedPeers {
			logger.Debugf("===============notify peer:%s to peer:%s================", peerWapper.peer.ID(), v.peer.ID())

			v.peer.Notify("newPeer",
				struct {
					Id          string   `json:"id"`
					DisplayName string   `json:"displayName"`
					Device      Device_t `json:"device"`
				}{
					Id:          peerWapper.peer.ID(),
					DisplayName: peerWapper.data.DisplayName,
					Device:      peerWapper.data.Device,
				})
		}
		break
	case "createWebRtcTransport":
		// data = {"forceTcp":false,"producing":true,"consuming":false,"sctpCapabilities":{"numStreams":{"OS":1024,"MIS":1024}}}
		transport := rom.createWebRtcTransport(pr, request.Data)
		if transport == nil {
			logger.Errorf("createWebRtcTransport failed for peer:%s", pr.ID())
			return
		}

		logger.Debugf("=========createWebRtcTransport:add transport for peer:%s", peerWapper.peer.ID())
		peerWapper.transports[transport.Id()] = transport
		wrta := common.WebRtcTransportAccept{
			Id:             transport.data.Id,
			IceParameters:  transport.data.IceParameters,
			IceCandidates:  transport.data.IceCandidates,
			DtlsParameters: transport.data.DtlsParameters,
			SctpParameter:  transport.data.SctpParameter,
		}
		accept(wrta)
		break
	case "connectWebRtcTransport":
		logger.Infof("receive connectWebRtcTransport============")
		var cwrtd common.ConnectWebRtcTransportData
		_ = json.Unmarshal(request.Data, &cwrtd)
		transport := peerWapper.transports[cwrtd.TransportId].(*WebRtcTransport)
		transport.connect(&cwrtd.DtlsParameters)
		rom.router.channel.AddTransport(cwrtd.TransportId, rom.router)
		accept(common.NilAccept{})
		break
	case "restartIce":
		break
	case "produce":
		logger.Infof("receive produce============")
		var pd common.ClientProduceData
		_ = json.Unmarshal(request.Data, &pd)

		transport := peerWapper.transports[pd.TransportId].(*WebRtcTransport)
		if transport == nil {
			reject(400, "No transport existed")
		}
		peerWapper := rom.peers[pr.ID()]
		producer := transport.produce(peerWapper, &pd, "", false, 5000)
		rom.router.channel.AddProducer(producer.id, rom.router)
		peerWapper.producers[producer.id] = producer
		rom.producerToPeer[producer.id] = peerWapper
		logger.Infof("produce id :%s========================", producer.id)
		accept(common.ProduceResp{
			Id: producer.Id(),
		})

		// create consumer

		logger.Debugf("==========room peer number:%d===========", len(rom.peers))
		for _, otherPeer := range rom.peers {
			logger.Debugf("==========otherPeer:%s,peer:%s===========", otherPeer.peer.ID(), peerWapper.peer.ID())
			if otherPeer.peer.ID() != peerWapper.peer.ID() {
				rom.CreateConsumer(otherPeer, peerWapper, producer)
			}
		}
		break
	case "closeProducer":
		break
	case "pauseProducer":
		break
	case "resumeProducer":
		break
	case "pauseConsumer":
		break
	case "resumeConsumer":
		break
	case "setConsumerPreferredLayers":
		break
	case "setConsumerPriority":
		break
	case "requestConsumerKeyFrame":
		break
	case "produceData":
		var pdd common.ProduceDataData

		if !peerWapper.data.Joined {
			logger.Errorf("Peer not yet joined")
		}
		_ = json.Unmarshal(request.Data, &pdd)

		transport := peerWapper.transports[pdd.TransportId].(*WebRtcTransport)
		if transport == nil {
			reject(400, "No transport existed")
		}

		dataProducer := transport.produceData(&pdd)
		peerWapper.dataProducers[dataProducer.Id()] = dataProducer

		accept(common.DataProduceResp{
			Id: dataProducer.Id(),
		})

		switch dataProducer.Label() {
		case "chat":
			for _, otherPeer := range rom.peers {
				if otherPeer.peer.ID() != peerWapper.peer.ID() {
					rom.CreateDataConsumer(otherPeer, peerWapper, dataProducer)
				}
			}
			break
		case "bot":
			break
		default:
			break
		}

		break
	case "changeDisplayName":
		break
	case "getTransportStats":

		type TransportStatReq struct {
			TransportId string `json:"transportId"`
		}

		var transportStatReq TransportStatReq
		_ = json.Unmarshal(request.Data, &transportStatReq)
		transport := peerWapper.transports[transportStatReq.TransportId]

		wrt, ok := transport.(*WebRtcTransport)
		if ok {
			stats := wrt.getStats()

			accept(stats)
		} else {
			accept(&common.NilAccept{})
		}

		//accept(&common.NilAccept{})

		break
	case "getProducerStats":

		type ProducerStatReq struct {
			ProducerId string `json:"producerId"`
		}

		var psr ProducerStatReq
		_ = json.Unmarshal(request.Data, &psr)

		producer := peerWapper.producers[psr.ProducerId]

		stats := producer.getStats()
		accept(stats)

		break
	case "getConsumerStats":
		type ConsumerStatReq struct {
			ConsumerId string `json:"consumerId"`
		}

		var csr ConsumerStatReq
		_ = json.Unmarshal(request.Data, &csr)

		consumer := peerWapper.producers[csr.ConsumerId]

		stats := consumer.getStats()
		accept(stats)

		break
	case "getDataProducerStats":
		// var dp common.DataProducer
		// _ = json.Unmarshal(request.Data, &dp)

		// dataProducer := peerWapper.dataProducers[dp.DataProducerId]

		// stats := dataProducer.getStats()
		// accept(stats)

		accept(&common.NilAccept{})
		break
	case "getDataConsumerStats":
		// var dc common.DataConsumer
		// _ = json.Unmarshal(request.Data, &dc)
		// logger.Debugf("=======??????????????receive getDataConsumerStats:%d=========", dc.DataConsumerId)
		// dataConsumer := peerWapper.dataConsumers[dc.DataConsumerId]

		// stats := dataConsumer.getStats()
		// accept(stats)
		accept(&common.NilAccept{})
		break
	case "applyNetworkThrottle":
		accept(&common.NilAccept{})
		break
	case "resetNetworkThrottle":
		accept(&common.NilAccept{})
		break
	default:
		break
	}
}

func (rom *Room) GetProducerbyId(producerId string) *Producer {
	peer := rom.producerToPeer[producerId]
	if peer == nil {
		return nil
	}

	return peer.producers[producerId]
}

func (rom *Room) NotifyByProducer(method string, producerId string, resp interface{}) {
	peer := rom.producerToPeer[producerId]

	if peer == nil {
		logger.Errorf("=================No peer for producer :%s", producerId)
		return
	}

	peer.peer.Notify(method, resp)
	//rom.Notify(peer.peer, method, resp)
}

func (rom *Room) HandleNotification(pr *peer.Peer, notification map[string]interface{}) {
	logger.Infof("handleNotification => %s", notification["method"])

	method := notification["method"].(string)
	data := notification["data"].(map[string]interface{})

	//Forward notification to testRoom.
	rom.Notify(pr, method, data)
}

func (rom *Room) CreateConsumer(consumerPeer *PeerWrapper, producerPeer *PeerWrapper, producer *Producer) {

	logger.Debugf("==========CreateConsumer peer:%s,transport len:%d=========", consumerPeer.peer.ID(), len(consumerPeer.transports))
	for _, v := range consumerPeer.transports {
		if transport, ok := v.(*WebRtcTransport); ok {

			codecs := make([]interface{}, 0)
			for _, c := range consumerPeer.data.RtpCapabilities.Codecs {
				codecs = append(codecs, c)
			}
			rtpCap := rtp.RtpCapabilities{
				Codecs:           codecs,
				HeaderExtensions: consumerPeer.data.RtpCapabilities.HeaderExtensions,
				FecMechanisms:    consumerPeer.data.RtpCapabilities.FecMechanisms,
			}
			consumer := transport.consume(consumerPeer, producer.id, rtpCap, false, false)
			if consumer == nil {
				break
			}

			consumerPeer.consumers[consumer.internal.ConsumerId] = consumer
			if producerPeer != nil {
				consumerPeer.peer.Request("newConsumer", common.NewConsumerData{
					PeerId:         producerPeer.peer.ID(),
					ProducerId:     producer.id,
					Id:             consumer.internal.ConsumerId,
					Kind:           consumer.data.kind,
					RtpParameters:  consumer.data.rtpParameters,
					ConsumerType:   consumer.data.consumerType,
					ProducerPaused: consumer.producerPaused,
					AppData: common.NewConsumerAppData{
						PeerId: producerPeer.peer.ID(),
					},
				},
					func(result json.RawMessage) {
						logger.Infof("newConsumer success: =>  %s", result)
					},
					func(code int, err string) {
						logger.Infof("newConsumer reject: %d => %s", code, err)
					})
			}
		}
	}
}

func (rom *Room) CreateDataConsumer(dataConsumerPeer *PeerWrapper, dataProducerPeer *PeerWrapper, dataProducer *DataProducer) {

	bFound := false
	for _, transport := range dataConsumerPeer.transports {

		wrt, ok := transport.(*WebRtcTransport)
		if ok && wrt.appData.Consuming {
			bFound = true

			dataConsumer := wrt.consumeData(dataProducer.Id())
			if dataConsumer == nil {

				logger.Errorf("CreateDataConsumer() | Create data cosumer fail")
				return
			}

			//dataConsumerPeer.dataConsumers[dataComsumer.id] = dataConsumer
			// Send a protoo request to the remote Peer with Consumer parameters.
			break
		}
	}

	if !bFound {
		logger.Errorf("CreateDataConsumer() | Transport for consuming not found")
	}

}

func (rom *Room) HandleClose(pr *peer.Peer, code int, err string) {

}

func (rom *Room) OnSctpStateChange(sctpState string) {

}

func (rom *Room) OnDtlsStateChange(dtlsState string) {

}

func (rom *Room) OnTrace(trace []byte) {

}

func (rom *Room) createWebRtcTransport(pr *peer.Peer, data json.RawMessage) *WebRtcTransport {
	type TempCmd struct {
		SctpCapabilities json.RawMessage `json:sctpCapabilities`
	}

	enableSctp := false
	var tc TempCmd
	json.Unmarshal(data, &tc)
	if len(tc.SctpCapabilities) > 0 {
		enableSctp = true
	}
	var cm common.CmdMessage
	json.Unmarshal(data, &cm)
	logger.Debugf("%+v", cm)

	wrto := &common.WebRtcTransportOptions{
		WebRtcTransportOptions: rom.cf.Mediasoup.WebRtcTransportOptions,
		NumSctpStreams:         cm.SctpCapabilities.NumStreams,
		EnableSctp:             enableSctp,
		AppData: common.TransportAppData{
			Producing: cm.Producing,
			Consuming: cm.Consuming,
		},
	}

	if cm.ForceTcp {
		wrto.EnableUdp = false
		wrto.EnableTcp = true
	} else {
		wrto.EnableUdp = true
		wrto.EnableTcp = false
	}

	return rom.router.CreateWebRtcTransport(wrto)
}

func (rom *Room) getJoindPeers() []*PeerWrapper {
	peers := make([]*PeerWrapper, 0)

	for _, v := range rom.peers {
		if v.data.Joined {
			peers = append(peers, v)
		}
	}
	return peers
}
