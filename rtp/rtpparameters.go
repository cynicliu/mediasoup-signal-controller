package rtp

import "encoding/json"

type RtpCodecCapability struct {
	/**
	 * Media kind.
	 */
	Kind string `json:"kind"`
	/**
	 * The codec MIME media type/subtype (e.g. 'audio/opus', 'video/VP8').
	 */
	MimeType string `json:"mimeType"`

	/**
	 * The preferred RTP payload type.
	 */
	PreferredPayloadType int `json:"preferredPayloadType"`

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
	 * Codec specific parameters. Some parameters (such as 'packetization_mode'
	 * and 'profile-level-id' in H264 or 'profile-id' in VP9) are critical for
	 * codec matching.
	 */
	//Parameters RTPHeaderParameter `json:"parameters"`
	Parameters RTPHeaderParameter `json:"parameters"`

	/**
	 * Transport layer and codec-specific feedback messages for this codec.
	 */
	RtcpFeedback []RtcpFeedback `json:"rtcpFeedback"`

	Apt int `json:"apt"`
}

type RtxCodecParameter struct {
	Apt int `json:"apt"`
}
type RtxCodecCapability struct {
	Kind                 string             `json:"kind"`
	MimeType             string             `json:"mimeType"`
	PreferredPayloadType int                `json:"preferredPayloadType"`
	ClockRate            int                `json:"clockRate"`
	Parameters           RTPHeaderParameter `json:"parameters"`
	RtcpFeedback         []RtcpFeedback     `json:"rtcpFeedback"`
}

type RtpHeaderExtension struct {
	/**
	 * Media kind. If empty string, it's valid for all kinds.
	 * Default any media kind.
	 */
	//kind?: MediaKind | '';
	Kind string `json:"kind"`

	/*
	 * The URI of the RTP header extension, as defined in RFC 5285.
	 */
	Uri string `json:"uri"`

	/**
	 * The preferred numeric identifier that goes in the RTP packet. Must be
	 * unique.
	 */
	PreferredId int `json:"preferredId"`

	/**
	 * If true, it is preferred that the value in the header be encrypted as per
	 * RFC 6904. Default false.
	 */
	PreferredEncrypt bool `json:"preferredEncrypt"`

	/**
	 * If 'sendrecv', mediasoup supports sending and receiving this RTP extension.
	 * 'sendonly' means that mediasoup can send (but not receive) it. 'recvonly'
	 * means that mediasoup can receive (but not send) it.
	 */
	//direction RtpHeaderExtensionDirection
	Direction string `json:"direction"`
}

type RtpParameters struct {
	/**
	 * Supported media and RTX codecs.
	 */
	Codecs []RtpCodecCapability `json:"codecs"`
	/**
	 * Supported RTP header extensions.
	 */
	HeaderExtensions []RtpHeaderExtension `json:"headerExtensions"`

	/**
	 * Transmitted RTP streams and their settings.
	 */
	Encodings []RtpEncodingParameters `json:"encodings"`

	/**
	 * Parameters used for RTCP.
	 */
	Rtcp RtcpParameters `json:"rtcp"`

	AppData json.RawMessage `json:"appData"`
}

type RtpCapabilities struct {
	/**
	 * Supported media and RTX codecs.
	 * may be RtpCodecCapability or RtxCodecCapability
	 */
	// Codecs []RtpCodecCapability `json:"codecs"`
	Codecs []interface{} `json:"codecs"`
	/**
	 * Supported RTP header extensions.
	 */
	HeaderExtensions []RtpHeaderExtension `json:"headerExtensions"`
	/**
	 * Supported FEC mechanisms.
	 */
	FecMechanisms []string `json:"fecMechanisms"`
}

type RtpCapabilities_In struct {
	/**
	 * Supported media and RTX codecs.
	 * may be RtpCodecCapability or RtxCodecCapability
	 */
	Codecs []RtpCodecCapability `json:"codecs"`
	/**
	 * Supported RTP header extensions.
	 */
	HeaderExtensions []RtpHeaderExtension `json:"headerExtensions"`
	/**
	 * Supported FEC mechanisms.
	 */
	FecMechanisms []string `json:"fecMechanisms"`
}

type RtcpFeedback struct {
	/**
	 * RTCP feedback type.
	 */
	FeedbackType string `json:"type"`

	/**
	 * RTCP feedback parameter.
	 */
	Parameter string `json:"parameter"`
}

type RtxSsrc_t struct {
	Ssrc int `json:"ssrc"`
}

type RtpEncodingParameters struct {
	/**
	 * The media SSRC.
	 */
	Ssrc int `json:"ssrc"`

	/**
	 * The RID RTP extension value. Must be unique.
	 */
	Rid string `json:"rid"`

	/**
	 * Codec payload type this encoding affects. If unset, first media codec is
	 * chosen.
	 */
	CodecPayloadType int `json:"codecPayloadType"`

	/**
	 * RTX stream information. It must contain a numeric ssrc field indicating
	 * the RTX SSRC.
	 */
	RtxSsrc RtxSsrc_t `json:"rtx"`

	/**
	 * It indicates whether discontinuous RTP transmission will be used. Useful
	 * for audio (if the codec supports it) and for video screen sharing (when
	 * static content is being transmitted, this option disables the RTP
	 * inactivity checks in mediasoup). Default false.
	 */
	Dtx bool `json:"dtx"`

	/**
	 * Number of spatial and temporal layers in the RTP stream (e.g. 'L1T3').
	 * See webrtc-svc.
	 */
	ScalabilityMode string `json:"scalabilityMode"`

	/**
	 * Others.
	 */
	ScaleResolutionDownBy int `json:"scaleResolutionDownBy"`
	MaxBitrate            int `json:"maxBitrate"`
}

type RTPHeaderParameter struct {
	Packetization_mode      int `json:"packetization-mode"`
	Level_asymmetry_allowed int `json:"level-asymmetry-allowed"`

	Profile_level_id string `json:"profile-level-id"`
	Profile_id       int    `json:"profile-id"`
	Apt              int    `json:"apt"`
}
type RtpHeaderExtensionParameters struct {
	/**
	 * The URI of the RTP header extension, as defined in RFC 5285.
	 */
	Uri string `json:"uri"`

	/**
	 * The numeric identifier that goes in the RTP packet. Must be unique.
	 */
	Id int `json:"id"`

	/**
	 * If true, the value in the header is encrypted as per RFC 6904. Default false.
	 */
	Encrypt bool `json:"encrypt"`

	/**
	 * Configuration parameters for the header extension.
	 */
	Parameters RTPHeaderParameter `json:"parameters"`
}
