package rtp

func SupportedRtpCapabilities() *RtpCapabilities {

	return &RtpCapabilities{
		Codecs: []interface{}{
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/opus",
				ClockRate: 48000,
				Channels:  2,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:                 "audio",
				MimeType:             "audio/PCMU",
				PreferredPayloadType: 0,
				ClockRate:            8000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:                 "audio",
				MimeType:             "audio/PCMA",
				PreferredPayloadType: 8,
				ClockRate:            8000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/ISAC",
				ClockRate: 32000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/ISAC",
				ClockRate: 16000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:                 "audio",
				MimeType:             "audio/G722",
				PreferredPayloadType: 9,
				ClockRate:            8000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/iLBC",
				ClockRate: 8000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/SILK",
				ClockRate: 24000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/SILK",
				ClockRate: 16000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/SILK",
				ClockRate: 12000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/SILK",
				ClockRate: 8000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:                 "audio",
				MimeType:             "audio/CN",
				PreferredPayloadType: 13,
				ClockRate:            32000,
			},
			RtpCodecCapability{
				Kind:                 "audio",
				MimeType:             "audio/CN",
				PreferredPayloadType: 13,
				ClockRate:            16000,
			},
			RtpCodecCapability{
				Kind:                 "audio",
				MimeType:             "audio/CN",
				PreferredPayloadType: 13,
				ClockRate:            8000,
			},
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/telephone-event",
				ClockRate: 48000,
			},
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/telephone-event",
				ClockRate: 32000,
			},

			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/telephone-event",
				ClockRate: 16000,
			},
			RtpCodecCapability{
				Kind:      "audio",
				MimeType:  "audio/telephone-event",
				ClockRate: 8000,
			},
			RtpCodecCapability{
				Kind:      "video",
				MimeType:  "video/VP8",
				ClockRate: 90000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "nack"},
					{FeedbackType: "nack", Parameter: "pli"},
					{FeedbackType: "ccm", Parameter: "fir"},
					{FeedbackType: "goog-remb"},
					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "video",
				MimeType:  "video/VP9",
				ClockRate: 90000,
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "nack"},
					{FeedbackType: "nack", Parameter: "pli"},
					{FeedbackType: "ccm", Parameter: "fir"},
					{FeedbackType: "goog-remb"},
					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "video",
				MimeType:  "video/H264",
				ClockRate: 90000,
				Parameters: RTPHeaderParameter{
					Packetization_mode:      1,
					Level_asymmetry_allowed: 1,
				},
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "nack"},
					{FeedbackType: "nack", Parameter: "pli"},
					{FeedbackType: "ccm", Parameter: "fir"},
					{FeedbackType: "goog-remb"},
					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "video",
				MimeType:  "video/H264",
				ClockRate: 90000,
				Parameters: RTPHeaderParameter{
					Packetization_mode:      0,
					Level_asymmetry_allowed: 1,
				},
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "nack"},
					{FeedbackType: "nack", Parameter: "pli"},
					{FeedbackType: "ccm", Parameter: "fir"},
					{FeedbackType: "goog-remb"},
					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "video",
				MimeType:  "video/H265",
				ClockRate: 90000,
				Parameters: RTPHeaderParameter{
					Packetization_mode:      1,
					Level_asymmetry_allowed: 1,
				},
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "nack"},
					{FeedbackType: "nack", Parameter: "pli"},
					{FeedbackType: "ccm", Parameter: "fir"},
					{FeedbackType: "goog-remb"},
					{FeedbackType: "transport-cc"},
				},
			},
			RtpCodecCapability{
				Kind:      "video",
				MimeType:  "video/H265",
				ClockRate: 90000,
				Parameters: RTPHeaderParameter{
					Packetization_mode:      0,
					Level_asymmetry_allowed: 1,
				},
				RtcpFeedback: []RtcpFeedback{

					{FeedbackType: "nack"},
					{FeedbackType: "nack", Parameter: "pli"},
					{FeedbackType: "ccm", Parameter: "fir"},
					{FeedbackType: "goog-remb"},
					{FeedbackType: "transport-cc"},
				},
			},
		},
		HeaderExtensions: []RtpHeaderExtension{
			{
				Kind:             "audio",
				Uri:              "urn:ietf:params:rtp-hdrext:sdes:mid",
				PreferredId:      1,
				PreferredEncrypt: false,
				Direction:        "sendrecv",
			},
			{
				Kind:             "video",
				Uri:              "urn:ietf:params:rtp-hdrext:sdes:mid",
				PreferredId:      1,
				PreferredEncrypt: false,
				Direction:        "sendrecv",
			},
			{
				Kind:             "video",
				Uri:              "urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
				PreferredId:      2,
				PreferredEncrypt: false,
				Direction:        "recvonly",
			},
			{
				Kind:             "video",
				Uri:              "urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
				PreferredId:      3,
				PreferredEncrypt: false,
				Direction:        "recvonly",
			},
			{
				Kind:             "audio",
				Uri:              "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time",
				PreferredId:      4,
				PreferredEncrypt: false,
				Direction:        "sendrecv",
			},
			{
				Kind:             "video",
				Uri:              "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time",
				PreferredId:      4,
				PreferredEncrypt: false,
				Direction:        "sendrecv",
			},
			// NOTE: For audio we just enable transport-wide-cc-01 when receiving media.
			{
				Kind:             "audio",
				Uri:              "http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01",
				PreferredId:      5,
				PreferredEncrypt: false,
				Direction:        "recvonly",
			},
			{
				Kind:             "video",
				Uri:              "http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01",
				PreferredId:      5,
				PreferredEncrypt: false,
				Direction:        "sendrecv",
			},
			// NOTE: Remove this once framemarking draft becomes RFC.
			{
				Kind:             "video",
				Uri:              "http://tools.ietf.org/html/draft-ietf-avtext-framemarking-07",
				PreferredId:      6,
				PreferredEncrypt: false,
				Direction:        "sendrecv",
			},
			{
				Kind:             "video",
				Uri:              "urn:ietf:params:rtp-hdrext:framemarking",
				PreferredId:      7,
				PreferredEncrypt: false,
				Direction:        "sendrecv",
			},
			{
				Kind:             "audio",
				Uri:              "urn:ietf:params:rtp-hdrext:ssrc-audio-level",
				PreferredId:      10,
				PreferredEncrypt: false,
				Direction:        "sendrecv",
			},
			{
				Kind:             "video",
				Uri:              "urn:3gpp:video-orientation",
				PreferredId:      11,
				PreferredEncrypt: false,
				Direction:        "sendrecv",
			},
			{
				Kind:             "video",
				Uri:              "urn:ietf:params:rtp-hdrext:toffset",
				PreferredId:      12,
				PreferredEncrypt: false,
				Direction:        "sendrecv",
			},
		},
		FecMechanisms: []string{},
	}
}
