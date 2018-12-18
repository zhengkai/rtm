package rtm

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"
)

var (
	nsKeep = int64(1000000) // 基于纳秒做随机数时的随机范围

	// ErrUnknownMtype 协议返回错误的 mtype
	ErrUnknownMtype = errors.New(`unknown fpnn mtype`)

	// ErrAnswerStatus  协议返回错误的 mtype
	ErrAnswerStatus = errors.New(`something error in answer`)
)

// Mtype FPNN 协议返回类型
type Mtype uint8

const (
	// MtypeOneWay 单路
	MtypeOneWay = 0

	// MtypeTwoWay 双路
	MtypeTwoWay = 1

	// MtypeAnswer 回答
	MtypeAnswer = 2
)

func signature(pid int32, secret string) (salt int64, hash string) {

	salt = rand.Int63()

	s := fmt.Sprintf(`%d:%s:%d`, pid, secret, salt)
	// fmt.Println(s)

	hasher := md5.New()
	hasher.Write([]byte(s))
	hash = fmt.Sprintf(`%X`, hasher.Sum(nil))

	return
}

func genMID() int64 {
	return time.Now().UTC().UnixNano()/nsKeep*nsKeep + rand.Int63n(nsKeep)
}

func getSendBuffer(mtype Mtype, sizeOrStatus uint8) (buf bytes.Buffer) {

	buf.Write([]byte(`FPNN`))
	buf.WriteByte(1)            // VERSION
	buf.WriteByte(0x40)         // FLAG
	buf.WriteByte(uint8(mtype)) // mtype answer
	buf.WriteByte(sizeOrStatus) // size or status
	return buf
}

func read(conn *net.TCPConn) (r *Read, err error) {

	recv := make([]byte, 4096)

	size, err := conn.Read(recv)
	if err != nil {
		return
	}

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

	length := 0
	if mtype == MtypeAnswer {
		if recv[7] != 0 {
			err = ErrAnswerStatus
		}
	} else {
		length = int(recv[7])
	}
	methodEnd := 16 + length

	r.Content = recv[methodEnd:size]

	if mtype != MtypeOneWay {
		r.Seq = recv[12:16]
	}

	if length > 0 {
		r.Method = string(recv[16:methodEnd])
	}

	return
}
