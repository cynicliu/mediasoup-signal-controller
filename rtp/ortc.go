package rtp

import (
	"fmt"
	"mediasoup-signal-controller/conf"
	libs "mediasoup-signal-controller/lib"
	"mediasoup-signal-controller/utils"
	"strings"

	"github.com/cloudwebrtc/go-protoo/logger"
)

type RtpCodecParameters struct {
	/**
	 * The codec MIME media type/subtype (e.g. 'audio/opus', 'video/VP8').
	 */
	MimeType string `json:"mimeType"`

	/**
	 * The value that goes in the RTP Payload Type Field. Must be unique.
	 */
	PayloadType int `json:"payloadType"`

	/**
	 * Codec clock rate expressed in Hertz.
	 */
	ClockRate int `json:"clockRate"`

	/**
	 * The number of channels supported (e.g. two for stereo). Just for audio.
	 * Default 1.
	 */
	Channels int `json:"channels"`

	/**
	 * Codec-specific parameters available for signaling. Some parameters (such
	 * as 'packetization-mode' and 'profile-level-id' in H264 or 'profile-id' in
	 * VP9) are critical for codec matching.
	 */
	Parameters RTPHeaderParameter `json:"parameters"`

	/**
	 * Transport layer and codec-specific feedback messages for this codec.
	 */
	RtcpFeedback []RtcpFeedback `json:"rtcpFeedback"`
}

/**
 * Provides information on RTCP settings within the RTP parameters.
 *
 * If no cname is given in a producer's RTP parameters, the mediasoup transport
 * will choose a random one that will be used into RTCP SDES messages sent to
 * all its associated consumers.
 *
 * mediasoup assumes reducedSize to always be true.
 */
type RtcpParameters struct {
	/**
	 * The Canonical Name (CNAME) used by RTCP (e.g. in SDES messages).
	 */
	Cname string `json:"cname"`

	/**
	 * Whether reduced size RTCP RFC 5506 is configured (if true) or compound RTCP
	 * as specified in RFC 3550 (if false). Default true.
	 */
	ReducedSize bool `json:"reducedSize"`

	/**
	 * Whether RTCP-mux is used. Default true.
	 */
	Mux bool `json:"mux"`
}

type ClientRtpParameters struct {
	/**
	 * The MID RTP extension value as defined in the BUNDLE specification.
	 */
	Mid string `json:"mid"`

	/**
	 * Supported media and RTX codecs.
	 */
	Codecs []RtpCodecParameters `json:"codecs"`
	/**
	 * Supported RTP header extensions.
	 */
	HeaderExtensions []RtpHeaderExtensionParameters `json:"headerExtensions"`

	/**
	 * Transmitted RTP streams and their settings.
	 */
	Encodings []RtpEncodingParameters `json:"encodings"`

	/**
	 * Parameters used for RTCP.
	 */
	Rtcp RtcpParameters `json:"rtcp"`
}

type RtpMappingCodec struct {
	PayloadType       int `json:"payloadType"`
	MappedPayloadType int `json:"mappedPayloadType"`
}

type RtpMappingEncoding struct {
	Ssrc            int    `json:"ssrc"`
	Rid             string `json:"rid"`
	ScalabilityMode string `json:"scalabilityMode"`
	MappedSsrc      int    `json:"mappedSsrc"`
}

type RtpMapping struct {
	Codecs    []RtpMappingCodec    `json:"codecs"`
	Encodings []RtpMappingEncoding `json:"encodings"`
}

func validateRtpCapabilities(mediaCodecs conf.RouterOptions_t) {
	for _, value := range mediaCodecs.MediaCodecs {
		validateRtpCodecCapability(value)
	}
}

func validateRtpCodecCapability(codec conf.MediaCodec_t) bool {

	return true
}

func GenerateRouterRtpCapabilities(mediaCodecs conf.RouterOptions_t) *RtpCapabilities {

	logger.Debugf("=========GenerateRouterRtpCapabilities: config codec:%+v", mediaCodecs)

	rtpcap := &RtpCapabilities{}
	//rtpcap.Codecs = make([]RtpCodecCapability, 0)
	rtpcap.Codecs = make([]interface{}, 0)

	DynamicPayloadTypes := []int{
		100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
		111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121,
		122, 123, 124, 125, 126, 127, 96, 97, 98, 99,
	}

	clonedSupportedRtpCapabilities := SupportedRtpCapabilities()

	rtpcap.HeaderExtensions = clonedSupportedRtpCapabilities.HeaderExtensions

	for _, mediaCodec := range mediaCodecs.MediaCodecs {

		for _, clonedCodec := range clonedSupportedRtpCapabilities.Codecs {

			supportedCodec, _ := clonedCodec.(RtpCodecCapability)
			if matchCodecs(mediaCodec, supportedCodec, false, false) {

				// Clone the supported codec.
				codec := supportedCodec

				// If the given media codec has preferredPayloadType, keep it.
				if mediaCodec.PreferredPayloadType > 0 {
					codec.PreferredPayloadType = mediaCodec.PreferredPayloadType
				} else if codec.PreferredPayloadType > 0 {
					// No need to remove it from the list since it's not a dynamic value.
				} else {
					codec.PreferredPayloadType = DynamicPayloadTypes[0]
					DynamicPayloadTypes = utils.DeleteItemByIndexInArray(DynamicPayloadTypes, 0)
				}
				logger.Debugf("GenerateRouterRtpCapabilities configCodec:%+v, routerCodec:%+v", mediaCodec, codec)

				duplicated := false
				for _, v := range rtpcap.Codecs {

					rtpCodec, ok1 := v.(RtpCodecCapability)
					rtxCodec, ok2 := v.(RtxCodecCapability)

					if ok1 {
						if rtpCodec.PreferredPayloadType == codec.PreferredPayloadType {
							duplicated = true
						}
					}

					if ok2 {
						if rtxCodec.PreferredPayloadType == codec.PreferredPayloadType {
							duplicated = true
						}
					}
				}

				if duplicated {
					logger.Errorf("======================GenerateRouterRtpCapabilities duplicated PreferredPayloadType:%s===============", codec.PreferredPayloadType)
					continue
				}

				// Merge the media codec parameters.
				//...

				logger.Debugf("===============GenerateRouterRtpCapabilities add rtc:%+v==================", codec)
				rtpcap.Codecs = append(rtpcap.Codecs, codec)

				// Add a RTX video codec if video.
				if codec.Kind == "video" {

					pt := DynamicPayloadTypes[0]
					DynamicPayloadTypes = utils.DeleteItemByIndexInArray(DynamicPayloadTypes, 0)
					rtxCodec := RtxCodecCapability{
						Kind:                 codec.Kind,
						MimeType:             fmt.Sprintf("%s/rtx", codec.Kind),
						PreferredPayloadType: pt,
						ClockRate:            codec.ClockRate,
						Parameters: RTPHeaderParameter{
							Apt: codec.PreferredPayloadType,
						},
					}
					logger.Debugf("===============GenerateRouterRtpCapabilities add rtx:%+v==================", rtxCodec)
					rtpcap.Codecs = append(rtpcap.Codecs, rtxCodec)
				}
			}
		}
	}

	return rtpcap
}

func GetProducerRtpParametersMapping(clientParams *ClientRtpParameters, routerCaps *RtpCapabilities) *RtpMapping {

	type CodecPair struct {
		//codec    *RtpCodecParameters
		//capCodec *RtpCodecCapability
		codeci    int
		capCodeci int
	}
	rtpMapping := &RtpMapping{}
	rtpMapping.Codecs = make([]RtpMappingCodec, 0)
	rtpMapping.Encodings = make([]RtpMappingEncoding, 0)

	codecToCapCodec := make([]CodecPair, 0)

	for i, codec := range clientParams.Codecs {
		if !isRtxCodec(codec.MimeType) {
			match := false
			for j, capCodec := range routerCaps.Codecs {
				if matchCodecs(codec, capCodec, true, true) {
					codecToCapCodec = append(codecToCapCodec, CodecPair{
						codeci:    i,
						capCodeci: j,
					})
					logger.Debugf("========GetProducerRtpParametersMapping match client codec:%+v, router codec:%+v", codec, capCodec)

					match = true
				}
			}

			if !match {
				logger.Errorf("unsupported codec [mimeType:%s, payloadType:%d]",
					codec.MimeType, codec.PayloadType)
			}
		}
	}

	for _, codec := range clientParams.Codecs {

		if isRtxCodec(codec.MimeType) {

			match := false
			for associatedMediaCodeci, mediaCodec := range clientParams.Codecs {
				if mediaCodec.PayloadType == codec.Parameters.Apt {
					match = true

					routerCodeci := -1
					for _, pair := range codecToCapCodec {
						if pair.codeci == associatedMediaCodeci {
							routerCodeci = pair.capCodeci
						}
					}

					if routerCodeci >= 0 {
						for j, capCodec := range routerCaps.Codecs {

							capRtpCodec, ok1 := capCodec.(RtpCodecCapability)
							routerCapRtpCodec, ok2 := routerCaps.Codecs[routerCodeci].(RtpCodecCapability)
							routerCapRtxCodec, ok3 := routerCaps.Codecs[routerCodeci].(RtxCodecCapability)

							if ok1 && ok2 && isRtxCodec(capRtpCodec.MimeType) && capRtpCodec.Parameters.Apt == routerCapRtpCodec.PreferredPayloadType {

								codecToCapCodec = append(codecToCapCodec, CodecPair{
									codeci:    associatedMediaCodeci,
									capCodeci: j,
								})
							} else if ok1 && ok3 && isRtxCodec(capRtpCodec.MimeType) && capRtpCodec.Parameters.Apt == routerCapRtxCodec.PreferredPayloadType {

								codecToCapCodec = append(codecToCapCodec, CodecPair{
									codeci:    associatedMediaCodeci,
									capCodeci: j,
								})
							}
						}
					}
				}
			}

			if !match {
				logger.Errorf("missing media codec found for RTX PT [payloadType:%d]", codec.PayloadType)
			}
		}
	}

	for _, codecPair := range codecToCapCodec {

		routerCapRtpCodec, ok1 := routerCaps.Codecs[codecPair.capCodeci].(RtpCodecCapability)
		routerCapRtxCodec, ok2 := routerCaps.Codecs[codecPair.capCodeci].(RtxCodecCapability)

		if ok1 {
			rtpMapping.Codecs = append(rtpMapping.Codecs, RtpMappingCodec{
				PayloadType:       clientParams.Codecs[codecPair.codeci].PayloadType,
				MappedPayloadType: routerCapRtpCodec.PreferredPayloadType,
			})
		}

		if ok2 {
			rtpMapping.Codecs = append(rtpMapping.Codecs, RtpMappingCodec{
				PayloadType:       clientParams.Codecs[codecPair.codeci].PayloadType,
				MappedPayloadType: routerCapRtxCodec.PreferredPayloadType,
			})
		}

	}

	mappedSsrc := utils.RandomNumberGenerator(100000000, 999999999)
	for _, encoding := range clientParams.Encodings {
		rtpMappingEncoding := RtpMappingEncoding{
			MappedSsrc:      mappedSsrc,
			Rid:             encoding.Rid,
			Ssrc:            encoding.Ssrc,
			ScalabilityMode: encoding.ScalabilityMode,
		}

		rtpMapping.Encodings = append(rtpMapping.Encodings, rtpMappingEncoding)
		mappedSsrc++
	}
	return rtpMapping
}

func isRtxCodec(mimeType string) bool {
	if strings.Contains(mimeType, "rtx") {
		return true
	}

	return false
}

func matchCodecs(aCodec interface{}, bCodec interface{}, strict bool, modify bool) bool {

	// var p *RtpCodecCapability
	// //var pp *RtpCodecParameters
	// var q *RtpCodecCapability
	// //var qp *RtpCodecParameters

	// p, ok := aCodec.(*RtpCodecCapability)
	// if !ok {
	// 	//pp, _ = aCodec.(*RtpCodecParameters)
	// }

	// q, ok = bCodec.(*RtpCodecCapability)
	// if !ok {
	// 	//qp, _ = bCodec.(*RtpCodecParameters)
	// }

	var p RtpCodecParameters
	var q RtpCodecCapability
	p, ok1 := aCodec.(RtpCodecParameters)
	q, ok2 := bCodec.(RtpCodecCapability)

	if ok1 && ok2 {
		aMimeType := strings.ToLower(p.MimeType)
		bMimeType := strings.ToLower(q.MimeType)

		if aMimeType != bMimeType {
			return false
		}

		if p.ClockRate != q.ClockRate {
			return false
		}

		if p.Channels != q.Channels {
			return false
		}

		// Per codec special checks.
		switch aMimeType {
		case "video/h264":
			{
				aPacketizationMode := p.Parameters.Packetization_mode
				bPacketizationMode := q.Parameters.Packetization_mode

				if aPacketizationMode != bPacketizationMode {
					return false
				}

				// If strict matching check profile-level-id.
				if strict {
					logger.Debugf("=============IsSameProfile:a:%+v=========b:%+v==========", p.Parameters, q.Parameters)
					if len(p.Parameters.Profile_level_id) > 0 && len(q.Parameters.Profile_level_id) > 0 && !libs.IsSameProfile(p.Parameters.Profile_level_id, q.Parameters.Profile_level_id) {
						return false
					}

					var selectedProfileLevelId string

					selectedProfileLevelId, ok := libs.GenerateProfileLevelIdForAnswer(p.Parameters.Profile_level_id, q.Parameters.Profile_level_id)
					if !ok {
						logger.Errorf("No profile for profile")
						return false
					}

					if modify {

						if len(selectedProfileLevelId) > 0 {
							//p.Parameters.Profile_level_id = selectedProfileLevelId
						}

						// else
						// 	delete aCodec.parameters['profile-level-id'];
					}
				}

				break
			}

		case "video/vp9":
			{
				// If strict matching check profile-id.
				if strict {
					aProfileId := p.Parameters.Profile_id
					bProfileId := q.Parameters.Profile_id

					if aProfileId != bProfileId {
						return false
					}
				}

				break
			}
		}

		return true
	} else {
		var p conf.MediaCodec_t
		var q RtpCodecCapability
		p, ok1 := aCodec.(conf.MediaCodec_t)
		q, ok2 := bCodec.(RtpCodecCapability)

		if ok1 && ok2 {
			aMimeType := strings.ToLower(p.MimeType)
			bMimeType := strings.ToLower(q.MimeType)

			if aMimeType != bMimeType {
				return false
			}

			if p.ClockRate != q.ClockRate {
				return false
			}

			if p.Channels != q.Channels {
				return false
			}

			// Per codec special checks.
			switch aMimeType {
			case "video/h264":
				{
					aPacketizationMode := p.Parameters.Packetization_mode
					bPacketizationMode := q.Parameters.Packetization_mode

					if aPacketizationMode != bPacketizationMode {
						return false
					}

					// If strict matching check profile-level-id.
					if strict {
						if !libs.IsSameProfile(p.Parameters.Profile_level_id, q.Parameters.Profile_level_id) {
							return false
						}

						var selectedProfileLevelId string

						selectedProfileLevelId, ok := libs.GenerateProfileLevelIdForAnswer(p.Parameters.Profile_level_id, q.Parameters.Profile_level_id)
						if !ok {
							logger.Errorf("No profile for profile")
							return false
						}

						if modify {

							if len(selectedProfileLevelId) > 0 {
								//p.Parameters.Profile_level_id = selectedProfileLevelId
							}

							// else
							// 	delete aCodec.parameters['profile-level-id'];
						}
					}

					break
				}

			case "video/vp9":
				{
					// If strict matching check profile-id.
					if strict {
						aProfileId := p.Parameters.Profile_id
						bProfileId := q.Parameters.Profile_id

						if aProfileId != bProfileId {
							return false
						}
					}

					break
				}
			}

			return true
		}
	}
	return false
}

/**
 * Generate RTP parameters to be internally used by Consumers given the RTP
 * parameters of a Producer and the RTP capabilities of the Router.
 */
func GetConsumableRtpParameters(kind string, params *ClientRtpParameters, caps *RtpCapabilities, rtpMapping *RtpMapping) *ClientRtpParameters {

	consumableParams := &ClientRtpParameters{
		Codecs:           make([]RtpCodecParameters, 0),
		HeaderExtensions: make([]RtpHeaderExtensionParameters, 0),
		Encodings:        make([]RtpEncodingParameters, 0),
	}

	//for (const codec of params.codecs)
	for _, codec := range params.Codecs {

		if !isRtxCodec(codec.MimeType) {
			var consumableCodecPt int
			for _, entry := range rtpMapping.Codecs {
				if entry.PayloadType == codec.PayloadType {
					consumableCodecPt = entry.MappedPayloadType
				}
			}

			var matchedCapCodec RtpCodecCapability

			for _, capCodec := range caps.Codecs {

				rtpCapCodec, ok1 := capCodec.(RtpCodecCapability)

				if ok1 {
					if rtpCapCodec.PreferredPayloadType == consumableCodecPt {
						matchedCapCodec = rtpCapCodec
					}
				}

			}

			consumableCodec := RtpCodecParameters{
				MimeType:     matchedCapCodec.MimeType,
				PayloadType:  matchedCapCodec.PreferredPayloadType,
				ClockRate:    matchedCapCodec.ClockRate,
				Channels:     matchedCapCodec.Channels,
				Parameters:   codec.Parameters, // Keep the Producer codec parameters.
				RtcpFeedback: matchedCapCodec.RtcpFeedback,
			}

			consumableParams.Codecs = append(consumableParams.Codecs, consumableCodec)

			var consumableCapRtxCodec *RtxCodecCapability
			consumableCapRtxCodec = nil

			for _, capRtxCodec := range caps.Codecs {

				rtxCapCodec, ok1 := capRtxCodec.(RtxCodecCapability)

				if ok1 && isRtxCodec(rtxCapCodec.MimeType) && rtxCapCodec.Parameters.Apt == consumableCodec.PayloadType {
					consumableCapRtxCodec = &rtxCapCodec
				}
			}

			if consumableCapRtxCodec != nil {
				consumableRtxCodec := RtpCodecParameters{
					MimeType:     consumableCapRtxCodec.MimeType,
					PayloadType:  consumableCapRtxCodec.PreferredPayloadType,
					ClockRate:    consumableCapRtxCodec.ClockRate,
					Parameters:   consumableCapRtxCodec.Parameters,
					RtcpFeedback: consumableCapRtxCodec.RtcpFeedback,
				}

				consumableParams.Codecs = append(consumableParams.Codecs, consumableRtxCodec)
			}
		}
	}

	//for (const capExt of caps.headerExtensions!)
	for _, capExt := range caps.HeaderExtensions {

		// Just take RTP header extension that can be used in Consumers.
		if !(capExt.Kind != kind || (capExt.Direction != "sendrecv" && capExt.Direction != "sendonly")) {
			consumableExt := RtpHeaderExtensionParameters{
				Uri:     capExt.Uri,
				Id:      capExt.PreferredId,
				Encrypt: capExt.PreferredEncrypt,
			}

			consumableParams.HeaderExtensions = append(consumableParams.HeaderExtensions, consumableExt)
		}
	}

	for i, encoding := range params.Encodings {
		//RtpEncodingParameters
		consumableEncoding := encoding

		consumableEncoding.Ssrc = rtpMapping.Encodings[i].MappedSsrc

		consumableParams.Encodings = append(consumableParams.Encodings, consumableEncoding)
	}

	// Clone Producer encodings since we'll mangle them.

	// const consumableEncodings = utils.clone(params.encodings) as RtpEncodingParameters[];
	// for (let i = 0; i < consumableEncodings.length; ++i)
	// {
	// 	const consumableEncoding = consumableEncodings[i];
	// 	const { mappedSsrc } = rtpMapping.encodings[i];

	// 	// Remove useless fields.
	// 	delete consumableEncoding.rid;
	// 	delete consumableEncoding.rtx;
	// 	delete consumableEncoding.codecPayloadType;

	// 	// Set the mapped ssrc.
	// 	consumableEncoding.ssrc = mappedSsrc;

	// 	consumableParams.encodings!.push(consumableEncoding);
	// }

	consumableParams.Rtcp = RtcpParameters{
		Cname:       params.Rtcp.Cname,
		ReducedSize: true,
		Mux:         true,
	}

	return consumableParams
}

func GetConsumerRtpParameters(consumableRtpParameters *ClientRtpParameters, rtpCapabilities RtpCapabilities, pipe bool) *ClientRtpParameters {
	consumerParams := &ClientRtpParameters{
		Codecs:           make([]RtpCodecParameters, 0),
		HeaderExtensions: make([]RtpHeaderExtensionParameters, 0),
		Encodings:        make([]RtpEncodingParameters, 0),
	}

	rtxSupported := false

	for _, v := range consumableRtpParameters.Codecs {

		codec := v
		for _, k := range rtpCapabilities.Codecs {

			rtpCapCodec, ok1 := k.(RtpCodecCapability)
			rtxCapCodec, ok2 := k.(RtxCodecCapability)

			if ok1 && matchCodecs(codec, rtpCapCodec, true, false) {
				codec.RtcpFeedback = rtpCapCodec.RtcpFeedback

				logger.Debugf("=============GetConsumerRtpParameters:consumerableCodec:%+v, rtpCapCodec:%+v================", codec, rtpCapCodec)
				consumerParams.Codecs = append(consumerParams.Codecs, v)
				break
			}

			if ok2 && matchCodecs(codec, rtxCapCodec, true, false) {
				codec.RtcpFeedback = rtxCapCodec.RtcpFeedback

				logger.Debugf("=============GetConsumerRtpParameters:consumerableCodec:%+v, rtxCapCodec:%+v================", codec, rtxCapCodec)
				consumerParams.Codecs = append(consumerParams.Codecs, v)
				break
			}

		}
	}

	// Must sanitize the list of matched codecs by removing useless RTX codecs.

	//is := make([]int, 0)
	for i := len(consumerParams.Codecs) - 1; i >= 0; i-- {

		codec := consumerParams.Codecs[i]
		if isRtxCodec(codec.MimeType) {
			bFound := false
			for _, mediaCodec := range consumerParams.Codecs {

				if mediaCodec.PayloadType == codec.Parameters.Apt {
					logger.Debugf("=============equal codec mimetype:%s, %s, mediaCodec:%d, codec:%d,len:%d, i:%d,", codec.MimeType, mediaCodec.MimeType, mediaCodec.PayloadType, codec.Parameters.Apt, len(consumerParams.Codecs), i)
					bFound = true
					rtxSupported = true
					break
				}
			}

			if !bFound {
				logger.Debugf("=============remove codec:%d, len:%d", i, len(consumerParams.Codecs))
				consumerParams.Codecs = append(consumerParams.Codecs[:i], consumerParams.Codecs[i+1:]...)
			}
		}
	}

	// codecs := make([]RtpCodecParameters, 0)
	// for i := 0; i < len(consumerParams.Codecs)-1; i++ {

	// 	exist := false
	// 	for _, j := range is {
	// 		if i == j {
	// 			exist = true
	// 			break
	// 		}
	// 	}
	// 	if !exist {
	// 		codecs = append(codecs, consumerParams.Codecs[i])
	// 	}
	// }

	// consumerParams.Codecs = codecs

	if len(consumerParams.Codecs) == 0 || isRtxCodec(consumerParams.Codecs[0].MimeType) {
		logger.Debugf("============GetConsumerRtpParameters consumerParams.Codecs is nil or isRtxCodec=============")
		return nil
	}

	for _, v := range consumableRtpParameters.HeaderExtensions {

		match := true
		for _, k := range rtpCapabilities.HeaderExtensions {
			if !(k.PreferredId == v.Id && k.Uri == v.Uri) {
				match = false
				break
			}
		}

		if match {
			consumerParams.HeaderExtensions = append(consumerParams.HeaderExtensions, v)
		}
	}

	logger.Debugf("============GetConsumerRtpParameters:%+v=============", consumerParams)

	// Reduce codecs' RTCP feedback. Use Transport-CC if available, REMB otherwise.
	for _, v := range consumerParams.HeaderExtensions {
		if v.Uri == "http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01" {

			for _, k := range consumerParams.Codecs {

				feedbacks := make([]RtcpFeedback, 0)
				for _, l := range k.RtcpFeedback {
					if l.FeedbackType != "goog-remb" {
						feedbacks = append(feedbacks, l)
					}
				}

				k.RtcpFeedback = feedbacks
			}
		} else if v.Uri == "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time" {

			for _, k := range consumerParams.Codecs {

				feedbacks := make([]RtcpFeedback, 0)
				for _, l := range k.RtcpFeedback {
					if l.FeedbackType != "transport-cc" {
						feedbacks = append(feedbacks, l)
					}
				}

				k.RtcpFeedback = feedbacks
			}
		} else {

			for _, k := range consumerParams.Codecs {

				feedbacks := make([]RtcpFeedback, 0)
				for _, l := range k.RtcpFeedback {
					if l.FeedbackType != "transport-cc" && l.FeedbackType != "goog-remb" {
						feedbacks = append(feedbacks, l)
					}
				}

				k.RtcpFeedback = feedbacks
			}
		}
	}

	if !pipe {
		consumerEncoding := RtpEncodingParameters{
			Ssrc: utils.RandomNumberGenerator(100000000, 999999999),
		}

		if rtxSupported {
			consumerEncoding.RtxSsrc = RtxSsrc_t{Ssrc: consumerEncoding.Ssrc + 1}
		}

		// If any of the consumableParams.encodings has scalabilityMode, process it
		// (assume all encodings have the same value).
		for _, v := range consumableRtpParameters.Encodings {
			if len(v.ScalabilityMode) > 0 {
				scalabilityMode := v.ScalabilityMode

				// If there is simulast, mangle spatial layers in scalabilityMode.
				if len(consumableRtpParameters.Encodings) > 0 {
					scalabilityMode = fmt.Sprintf("S%dT3", len(consumableRtpParameters.Encodings))
				}

				consumerEncoding.ScalabilityMode = scalabilityMode
				break
			}
		}

		// Use the maximum maxBitrate in any encoding and honor it in the Consumer's
		// encoding.
		maxEncodingMaxBitrate := 0
		for _, v := range consumableRtpParameters.Encodings {
			if v.MaxBitrate > maxEncodingMaxBitrate {
				maxEncodingMaxBitrate = v.MaxBitrate
			}
		}

		if maxEncodingMaxBitrate > 0 {
			consumerEncoding.MaxBitrate = maxEncodingMaxBitrate
		}

		consumerParams.Encodings = append(consumerParams.Encodings, consumerEncoding)
	}

	return consumerParams
}
