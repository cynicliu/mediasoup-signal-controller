package main

import (
	"fmt"
	"mediasoup-signal-controller/common"
)

func main() {

	str := "30:"

	nlen := common.NsPayloadLength([]byte(str), 0)
	fmt.Println(nlen)

	str = ""
	nlen = common.NsPayloadLength([]byte(str), 0)
	fmt.Println(nlen)
	str = ":"
	nlen = common.NsPayloadLength([]byte(str), 0)
	fmt.Println(nlen)

	str = "3;"
	nlen = common.NsPayloadLength([]byte(str), 0)
	fmt.Println(nlen)

	str = "xxx30"
	nlen = common.NsPayloadLength([]byte(str), 0)
	fmt.Println(nlen)

	str = "030:"
	nlen = common.NsPayloadLength([]byte(str), 0)
	fmt.Println(nlen)

	fmt.Println(common.NsWriteLength(12))

	str = "38:{\"event\":\"running\",\"targetId\":\"15103\"},"
	buf, nlen := common.NsPayload([]byte(str), 0)
	fmt.Println(string(buf), nlen)

	str = "{data:data}"

	buf, nlen = common.NsWrite([]byte(str), 0, len(str)-1)
	fmt.Println(string(buf), nlen, len(str))

	buffer := []byte("{data:data}")
	fmt.Println(string(buffer[0:len(buffer)]))
}
