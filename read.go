package rtm

import (
	"bytes"
	"encoding/binary"
	"net"
)

func read(conn *net.TCPConn, pad []byte) (r []*Read, remain []byte, err error) {

	recv := make([]byte, 4096)

	if pad != nil {
		recv = append(pad, recv...)
	}

	size, err := conn.Read(recv)
	if err != nil {
		return
	}

	r, remain, err = readByte(recv[:size:size])

	return
}

func readByte(line []byte) (r []*Read, remain []byte, err error) {

	for {

		var rt *Read
		var isBreak = false

		rt, remain, isBreak, err = readParse(line)

		if isBreak {
			return
		}

		r = append(r, rt)

		if remain == nil {
			break
		}

		line = remain
	}

	return
}

func readParse(recv []byte) (r *Read, remain []byte, isBreak bool, err error) {

	r, err = readParseInit(recv)
	if err != nil {
		return
	}

	dataLen := len(recv)

	length := 0
	if r.Mtype == MtypeAnswer {
		if recv[7] != 0 {
			err = ErrAnswerStatus
		}
	} else {
		length = int(recv[7])
	}
	methodEnd := 16 + length

	var payloadLen uint32
	buf := bytes.NewBuffer(recv[8:12])
	binary.Read(buf, binary.LittleEndian, &payloadLen)

	fullLen := methodEnd + int(payloadLen)

	if fullLen > dataLen {
		return nil, recv, true, nil
	}

	// fmt.Println(`sample`, recv[:fullLen:fullLen])

	r.Content = recv[methodEnd:fullLen]

	if r.Mtype != MtypeOneWay {
		r.Seq = recv[12:16]
	}

	if length > 0 {
		r.Method = string(recv[16:methodEnd])
	}

	if fullLen < dataLen {
		remain = recv[fullLen:dataLen:dataLen]
	}

	return
}

func readParseInit(recv []byte) (r *Read, err error) {

	var mtype Mtype

	switch recv[6] {
	case 0,
		1,
		2:

		mtype = Mtype(recv[6])

	default:

		err = ErrUnknownMtype
		return
	}

	r = &Read{
		Mtype: mtype,
	}

	return
}
