package rtm

import (
	"encoding/binary"
	"fmt"
	"net"
)

func read(conn *net.TCPConn, pad []byte) (r []*Read, remain []byte, err error) {

	recv := make([]byte, 2000) // rtm 每个包最多就是 mtu 大小，我这里看到的是 1448

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

	// log.Println(`read`, size, len(remain), len(pad))

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

func readParse(v []byte) (r *Read, remain []byte, err error) {

	dataSize := len(v)

	// http://wiki.ifunplus.cn/display/core/FPNN+protocol

	if dataSize < 12 {
		remain = v
		return
	}

	r, err = readParseInit(v)
	if err != nil {
		return nil, nil, err
	}

	// fmt.Println(`readParse`, v)

	headSize, methodSize, fullSize, err := readParseSize(v, r.Mtype)
	if err != nil {
		return nil, nil, err
	}

	if fullSize > dataSize { // 长度不够，意味着要等下一行数据才能拼出来
		return nil, v, nil
	}

	if headSize > 12 {
		r.Seq = v[12:headSize]
	}

	beforePayload := headSize + methodSize

	if beforePayload < fullSize {
		r.Content = v[beforePayload:fullSize]
	}

	if methodSize > 0 {
		r.Method = string(v[headSize:beforePayload])
	}

	if fullSize < dataSize {
		// fmt.Println(`make remain`, fullSize, dataSize)
		remain = v[fullSize:]
	}

	return
}

func readParseInit(v []byte) (r *Read, err error) {

	var mtype Mtype

	switch v[6] {
	case 0,
		1,
		2:

		mtype = Mtype(v[6])

	default:

		err = ErrUnknownMtype
		return
	}

	r = &Read{
		Mtype: mtype,
	}

	return
}

func readParseSize(v []byte, mtype Mtype) (headSize, methodSize int, fullSize int, err error) {

	headSize = 12 // Method Name 之前的字节数
	if mtype != MtypeOneWay {
		headSize += 4
	}

	if mtype == MtypeAnswer {
		if v[7] != 0 {
			fmt.Println(`ErrAnswerStatus`, v[7])
			err = ErrAnswerStatus
			return
		}
	} else {
		methodSize = int(v[7]) // Method Name 字节数
	}

	// payload 字节数
	payloadSize := binary.LittleEndian.Uint32(v[8:12])

	// fmt.Println(`calc fullSize`, MtypeAnswer, v[0:12], headSize, methodSize, payloadSize)

	fullSize = headSize + methodSize + int(payloadSize)

	return
}
