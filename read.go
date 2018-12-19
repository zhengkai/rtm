package rtm

import (
	"bytes"
	"encoding/binary"
	"net"
)

func read(conn *net.TCPConn, pad []byte) (r []*Read, remain []byte, err error) {

	recv := make([]byte, 4096)

	size, err := conn.Read(recv)
	if err != nil {
		return
	}

	if len(pad) > 0 {
		recv = append(pad, recv[:size]...)
	} else {
		recv = recv[:size:size]
	}

	r, remain, err = readByte(recv)

	if err != nil {
		conn.Close()
	}

	return
}

func readByte(line []byte) (r []*Read, remain []byte, err error) {

	for {

		var rp *Read

		rp, remain, err = readParse(line)

		if err != nil {
			break
		}

		if rp == nil {
			break
		} else {
			r = append(r, rp)
		}
		if remain == nil {
			break
		}

		line = remain
	}

	return
}

func readParse(recv []byte) (r *Read, remain []byte, err error) {

	dataSize := len(recv)

	// http://wiki.ifunplus.cn/display/core/FPNN+protocol

	if dataSize < 12 {
		remain = recv
		return
	}

	r, err = readParseInit(recv)
	if err != nil {
		return nil, nil, err
	}

	// fmt.Println(`readParse`, recv)

	headSize, methodSize, fullSize, err := readParseSize(recv, r.Mtype)
	if err != nil {
		return nil, nil, err
	}

	if fullSize > dataSize { // 长度不够，意味着要等下一行数据才能拼出来
		return nil, recv, nil
	}

	if headSize > 12 {
		r.Seq = recv[12:headSize]
	}

	beforePayload := headSize + methodSize

	if beforePayload < fullSize {
		r.Content = recv[beforePayload:fullSize]
	}

	if methodSize > 0 {
		r.Method = string(recv[headSize:beforePayload])
	}

	if fullSize < dataSize {
		// fmt.Println(`make remain`, fullSize, dataSize)
		remain = recv[fullSize:]
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

func readParseSize(recv []byte, mtype Mtype) (headSize, methodSize int, fullSize int, err error) {

	headSize = 12 // Method Name 之前的字节数
	if mtype != MtypeOneWay {
		headSize += 4
	}

	if mtype == MtypeAnswer {
		if recv[7] != 0 {
			err = ErrAnswerStatus
			return
		}
	} else {
		methodSize = int(recv[7]) // Method Name 字节数
	}

	var payloadSize uint32 // payload 字节数
	buf := bytes.NewBuffer(recv[8:12])
	binary.Read(buf, binary.LittleEndian, &payloadSize)

	// fmt.Println(`calc fullSize`, MtypeAnswer, recv[0:12], headSize, methodSize, payloadSize)

	fullSize = headSize + methodSize + int(payloadSize)

	return
}
