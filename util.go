package rtm

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
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
	buf.WriteByte(0x80)         // FLAG 0x40 json 0x200 msgpack
	buf.WriteByte(uint8(mtype)) // mtype answer
	buf.WriteByte(sizeOrStatus) // size or status
	return buf
}
