package common

import (
	"fmt"

	"github.com/cloudwebrtc/go-protoo/logger"
)

type NetString struct{}

/*
net strting example:------11:{data:data},  11 = len({data:data})
*/
func NsPayloadLength(buffer []byte, off int) int {

	ret := 0
	nLen := len(buffer)
	i := off
	for {
		//logger.Infof("i:%d", i)
		if i >= nLen {
			break
		}
		cc := buffer[i]
		if cc == 0x3a { // 0x3a = :
			if i == off {
				logger.Errorf("Invalid netstring with leading ':'")
				return -1
			}

			//logger.Infof("length:%d, i:%d, str=%s", ret, i, string(buffer))
			return ret
		}

		if cc < 0x30 || cc > 0x39 {
			logger.Errorf("Unexpected no number character: %c", cc)
			return -1
		}

		ret = ret*10 + (int)(cc-0x30)
		//logger.Infof("length:%d", ret)
		if ret == 0 {
			logger.Errorf("Invalid netstring with leading 0")
			return -1
		}
		i++
	}

	return ret
}

/*
  writelen = len(12:{data:data},)
*/
func NsWriteLength(nlen int) int {

	if nlen < 0 {
		return nlen
	}

	nslen := nlen
	for {
		if nlen < 10 {
			break
		}

		nslen += 1
		nlen /= 10
	}

	// nslen + 1 (last digit) + 1 (:) + 1 (,)
	return nslen + 3
}

func NsLength(buffer []byte, off int) int {
	return NsWriteLength(NsPayloadLength(buffer, off))
}

func NsPayload(buffer []byte, off int) ([]byte, int) {
	nlen := NsPayloadLength(buffer, off)

	if nlen < 0 {
		return nil, nlen
	}

	nslen := NsWriteLength(nlen)

	if len(buffer)-off-nslen < 0 {
		return nil, -1
	}

	start := off + nslen - nlen - 1
	retB := buffer[start : start+nlen]

	return retB, nlen
}

func NsWrite(buffer []byte, start int, end int) ([]byte, int) {

	nlen := end - start + 1
	nslen := NsWriteLength(nlen)

	var str string
	str = fmt.Sprintf("%d", nlen)
	str = str + ":"

	retB := []byte(str)
	temp := buffer[start : end+1]
	retB = append(retB, temp...)
	retB = append(retB, ',')

	return retB, nslen
}
