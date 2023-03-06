package libs

import (
	"fmt"
	"strconv"

	"github.com/cloudwebrtc/go-protoo/logger"
)

const (
	ProfileConstrainedBaseline = 1
	ProfileBaseline            = 2
	ProfileMain                = 3
	ProfileConstrainedHigh     = 4
	ProfileHigh                = 5

	// All values are equal to ten times the level number, except level 1b which is
	// special.
	Level1_b = 0
	Level1   = 10
	Level1_1 = 11
	Level1_2 = 12
	Level1_3 = 13
	Level2   = 20
	Level2_1 = 21
	Level2_2 = 22
	Level3   = 30
	Level3_1 = 31
	Level3_2 = 32
	Level4   = 40
	Level4_1 = 41
	Level4_2 = 42
	Level5   = 50
	Level5_1 = 51
	Level5_2 = 52
)

// For level_idc=11 and profile_idc=0x42, 0x4D, or 0x58, the constraint set3
// flag specifies if level 1b or level 1.1 is used.
const ConstraintSet3Flag = 0x10

type ProfileLevelId struct {
	profile int
	level   int
}

type BitPattern struct {
	mask        byte
	maskedValue byte
}

func (bt *BitPattern) isMatch(value byte) bool {
	return bt.maskedValue == (value & bt.mask)
}

type ProfilePattern struct {
	profile_idc byte
	profile_iop *BitPattern
	profile     int
}

func CreateProfilePatterns() []*ProfilePattern {

	return []*ProfilePattern{
		{
			profile_idc: 0x42,
			profile_iop: &BitPattern{
				mask:        ^byteMaskString('x', []byte("x1xx0000")),
				maskedValue: byteMaskString('1', []byte("x1xx0000")),
			},
			profile: ProfileConstrainedBaseline,
		},
		{
			profile_idc: 0x4D,
			profile_iop: &BitPattern{
				mask:        ^byteMaskString('x', []byte("1xxx0000")),
				maskedValue: byteMaskString('1', []byte("1xxx0000")),
			},
			profile: ProfileConstrainedBaseline,
		},
		{
			profile_idc: 0x58,
			profile_iop: &BitPattern{
				mask:        ^byteMaskString('x', []byte("11xx0000")),
				maskedValue: byteMaskString('1', []byte("11xx0000")),
			},
			profile: ProfileConstrainedBaseline,
		},
		{
			profile_idc: 0x42,
			profile_iop: &BitPattern{
				mask:        ^byteMaskString('x', []byte("x0xx0000")),
				maskedValue: byteMaskString('1', []byte("x0xx0000")),
			},
			profile: ProfileBaseline,
		},
		{
			profile_idc: 0x58,
			profile_iop: &BitPattern{
				mask:        ^byteMaskString('x', []byte("10xx0000")),
				maskedValue: byteMaskString('1', []byte("10xx0000")),
			},
			profile: ProfileBaseline,
		},
		{
			profile_idc: 0x4D,
			profile_iop: &BitPattern{
				mask:        ^byteMaskString('x', []byte("0x0x0000")),
				maskedValue: byteMaskString('1', []byte("0x0x0000")),
			},
			profile: ProfileMain,
		},
		{
			profile_idc: 0x64,
			profile_iop: &BitPattern{
				mask:        ^byteMaskString('x', []byte("00000000")),
				maskedValue: byteMaskString('1', []byte("00000000")),
			},
			profile: ProfileHigh,
		},
		{
			profile_idc: 0x64,
			profile_iop: &BitPattern{
				mask:        ^byteMaskString('x', []byte("00001100")),
				maskedValue: byteMaskString('1', []byte("00001100")),
			},
			profile: ProfileConstrainedHigh,
		},
	}
}

// Convert a string of 8 characters into a byte where the positions containing
// character c will have their bit set. For example, c = 'x', str = "x1xx0000"
// will return 0b10110000.
func byteMaskString(c byte, str []byte) byte {

	return ((btoi(str[0] == c) << 7) |
		(btoi(str[1] == c) << 6) |
		(btoi(str[2] == c) << 5) |
		(btoi(str[3] == c) << 4) |
		(btoi(str[4] == c) << 3) |
		(btoi(str[5] == c) << 2) |
		(btoi(str[6] == c) << 1) |
		(btoi(str[7] == c) << 0))
}

func btoi(b bool) byte {
	if b {
		return 1
	}
	return 0
}

/**
 * Parse profile level id that is represented as a string of 3 hex bytes.
 * Nothing will be returned if the string is not a recognized H264 profile
 * level id.
 *
 * @param {String} str - profile-level-id value as a string of 3 hex bytes.
 *
 * @returns {ProfileLevelId}
 */
func parseProfileLevelId(str string) *ProfileLevelId {
	// The string should consist of 3 bytes in hexadecimal format.
	if len(str) != 6 {
		return nil
	}

	profile_level_id_numeric, _ := strconv.Atoi(str)

	if profile_level_id_numeric == 0 {
		return nil
	}

	// Separate into three bytes.
	level_idc := profile_level_id_numeric & 0xFF
	profile_iop := byte(profile_level_id_numeric>>8) & 0xFF
	profile_idc := byte(profile_level_id_numeric>>16) & 0xFF

	// Parse level based on level_idc and constraint set 3 flag.
	var level int

	switch level_idc {
	case Level1_1:
		{
			if (profile_iop & ConstraintSet3Flag) != 0 {
				level = Level1_b
			} else {
				level = Level1_1
			}
			break
		}
	case Level1:
	case Level1_2:
	case Level1_3:
	case Level2:
	case Level2_1:
	case Level2_2:
	case Level3:
	case Level3_1:
	case Level3_2:
	case Level4:
	case Level4_1:
	case Level4_2:
	case Level5:
	case Level5_1:
	case Level5_2:
		{
			level = level_idc
			break
		}
	// Unrecognized level_idc.
	default:
		{
			// debug('parseProfileLevelId() | unrecognized level_idc:%s', level_idc);
			logger.Debugf("parseProfileLevelId() | unrecognized level_idc:%d", level_idc)
			return nil
		}
	}

	// Parse profile_idc/profile_iop into a Profile enum.
	//for (const pattern of ProfilePatterns)
	ProfilePatterns := CreateProfilePatterns()
	for _, pattern := range ProfilePatterns {
		if profile_idc == pattern.profile_idc && pattern.profile_iop.isMatch(profile_iop) {

			return &ProfileLevelId{
				profile: pattern.profile,
				level:   level,
			}
		}
	}

	logger.Errorf("parseProfileLevelId() | unrecognized profile_idc/profile_iop combination")

	return nil
}

func isLessLevel(a int, b int) bool {
	if a == Level1_b {
		return b != Level1 && b != Level1_b
	}

	if b == Level1_b {
		return a != Level1
	}

	return a < b
}

func minLevel(a int, b int) int {

	ret := isLessLevel(a, b)
	if ret {
		return a
	}
	return b
}

func isLevelAsymmetryAllowed(level_asymmetry_allowed int) bool {
	return level_asymmetry_allowed == 1
}

/**
 * Returns canonical string representation as three hex bytes of the profile
 * level id, or returns nothing for invalid profile level ids.
 *
 * @param {ProfileLevelId} profile_level_id
 *
 * @returns {String}
 */
func profileLevelIdToString(profileLevelId *ProfileLevelId) string {
	// Handle special case level == 1b.
	if profileLevelId.level == Level1_b {
		switch profileLevelId.profile {

		case ProfileConstrainedBaseline:
			{
				return "42f00b"
			}
		case ProfileBaseline:
			{
				return "42100b"
			}
		case ProfileMain:
			{
				return "4d100b"
			}
		// Level 1_b is not allowed for other profiles.
		default:
			{
				logger.Debugf("profileLevelIdToString() | Level 1_b not is allowed for profile:%d", profileLevelId.profile)
				return ""
			}
		}
	}

	var profile_idc_iop_string string

	switch profileLevelId.profile {
	case ProfileConstrainedBaseline:
		{
			profile_idc_iop_string = "42e0"
			break
		}
	case ProfileBaseline:
		{
			profile_idc_iop_string = "4200"
			break
		}
	case ProfileMain:
		{
			profile_idc_iop_string = "4d00"
			break
		}
	case ProfileConstrainedHigh:
		{
			profile_idc_iop_string = "640c"
			break
		}
	case ProfileHigh:
		{
			profile_idc_iop_string = "6400"
			break
		}
	default:
		{
			logger.Debugf("profileLevelIdToString() | unrecognized profile:%d", profileLevelId.profile)

			return ""
		}
	}

	levelStr := strconv.Itoa(int(profileLevelId.level))

	if len(levelStr) == 1 {
		levelStr = fmt.Sprintf("0%s", levelStr)
	}

	return fmt.Sprintf("%s%s", profile_idc_iop_string, levelStr)
}

/**
 * Parse profile level id that is represented as a string of 3 hex bytes
 * contained in an SDP key-value map. A default profile level id will be
 * returned if the profile-level-id key is missing. Nothing will be returned if
 * the key is present but the string is invalid.
 *
 * @param {Object} [params={}] - Codec parameters object.
 *
 * @returns {ProfileLevelId}
 */
func parseSdpProfileLevelId(profile_level_id string) *ProfileLevelId {
	if len(profile_level_id) == 0 {

		return &ProfileLevelId{
			profile: ProfileConstrainedBaseline,
			level:   Level3_1,
		}
	} else {
		return parseProfileLevelId(profile_level_id)
	}
}

/**
 * Returns true if the parameters have the same H264 profile, i.e. the same
 * H264 profile (Baseline, High, etc).
 *
 * @param {Object} [params1={}] - Codec parameters object.
 * @param {Object} [params2={}] - Codec parameters object.
 *
 * @returns {Boolean}
 */
func IsSameProfile(profile_level_id1 string, profile_level_id12 string) bool {
	profile_level_id_1 := parseSdpProfileLevelId(profile_level_id1)
	profile_level_id_2 := parseSdpProfileLevelId(profile_level_id12)

	// Compare H264 profiles, but not levels.
	return profile_level_id_1.profile == profile_level_id_2.profile
}

/**
 * Generate codec parameters that will be used as answer in an SDP negotiation
 * based on local supported parameters and remote offered parameters. Both
 * local_supported_params and remote_offered_params represent sendrecv media
 * descriptions, i.e they are a mix of both encode and decode capabilities. In
 * theory, when the profile in local_supported_params represent a strict superset
 * of the profile in remote_offered_params, we could limit the profile in the
 * answer to the profile in remote_offered_params.
 *
 * However, to simplify the code, each supported H264 profile should be listed
 * explicitly in the list of local supported codecs, even if they are redundant.
 * Then each local codec in the list should be tested one at a time against the
 * remote codec, and only when the profiles are equal should this function be
 * called. Therefore, this function does not need to handle profile intersection,
 * and the profile of local_supported_params and remote_offered_params must be
 * equal before calling this function. The parameters that are used when
 * negotiating are the level part of profile-level-id and level-asymmetry-allowed.
 *
 * @param {Object} [local_supported_params={}]
 * @param {Object} [remote_offered_params={}]
 *
 * @returns {String} Canonical string representation as three hex bytes of the
 *   profile level id, or null if no one of the params have profile-level-id.
 *
 * @throws {TypeError} If Profile mismatch or invalid params.
 */
func GenerateProfileLevelIdForAnswer(local_supported_profile_level_id string, remote_offered_params_profile_level_id string) (string, bool) {
	// If both local and remote params do not contain profile-level-id, they are
	// both using the default profile. In this case, don't return anything.

	// Parse profile-level-ids.
	n_local_supported_profile_level_id, _ := strconv.Atoi(local_supported_profile_level_id)
	n_remote_offered_params_profile_level_id, _ := strconv.Atoi(remote_offered_params_profile_level_id)
	local_profile_level_id := parseSdpProfileLevelId(local_supported_profile_level_id)
	remote_profile_level_id := parseSdpProfileLevelId(remote_offered_params_profile_level_id)

	// The local and remote codec must have valid and equal H264 Profiles.
	if local_profile_level_id == nil {
		logger.Errorf("invalid local_profile_level_id")

		return "", false
	}

	if remote_profile_level_id == nil {
		logger.Errorf("invalid remote_profile_level_id")
		return "", false
	}

	if local_profile_level_id.profile != remote_profile_level_id.profile {
		logger.Errorf("H264 Profile mismatch")
		return "", false
	}

	// Parse level information.
	level_asymmetry_allowed := isLevelAsymmetryAllowed(n_local_supported_profile_level_id) &&
		isLevelAsymmetryAllowed(n_remote_offered_params_profile_level_id)

	local_level := local_profile_level_id.level
	remote_level := remote_profile_level_id.level
	min_level := minLevel(local_level, remote_level)

	// Determine answer level. When level asymmetry is not allowed, level upgrade
	// is not allowed, i.e., the level in the answer must be equal to or lower
	// than the level in the offer.
	answer_level := n_local_supported_profile_level_id
	if !level_asymmetry_allowed {
		answer_level = min_level
	}

	logger.Debugf(
		"generateProfileLevelIdForAnswer() | result: [profile:%d, level:%d]",
		local_profile_level_id.profile, answer_level)

	// Return the resulting profile-level-id for the answer parameters.
	return profileLevelIdToString(
		&ProfileLevelId{
			profile: local_profile_level_id.profile,
			level:   answer_level,
		}), true
}
