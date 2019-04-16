package rtm

// https://rasolution-docs.ifunplus.cn/rtm/rtmserver/serverapi/

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack"
)

// Server 服务器连接
type Server struct {
	Conn   *net.TCPConn
	Config *Config
	seq    uint32
	mid    int64
	lock   sync.Mutex
}

type serverBase struct {
	PID  int32  `json:"pid" msgpack:"pid"`
	Sign string `json:"sign" msgpack:"sign"`
	Salt int64  `json:"salt" msgpack:"salt"`
}

type serverBaseMsg struct {
	MType int8   `json:"mtype" msgpack:"mtype"`
	From  int64  `json:"from" msgpack:"from"`
	MID   int64  `json:"mid" msgpack:"mid"`
	Msg   string `json:"msg" msgpack:"msg"`
	Attrs string `json:"attrs" msgpack:"attrs"`
}

type serverSendmsg struct {
	serverBase
	serverBaseMsg
	To int64 `msgpack:"to"`
}

type serverSendroommsg struct {
	serverBase
	serverBaseMsg
	Rid int64 `json:"rid" msgpack:"rid"`
}

type serverSendmsgs struct {
	serverBase
	serverBaseMsg
	Tos []int64 `msgpack:"tos"`
}

type serverGettoken struct {
	serverBase
	UID int64 `msgpack:"uid"`
}

type serverGettokenReturn struct {
	Token string `msgpack:"token"`
}

// NewServer 建立新 Server 端
func NewServer() *Server {
	if globalConfig == nil {
		return nil
	}
	return &Server{
		Config: globalConfig,
	}
}

// Connect 创建连接
func (s *Server) Connect() (err error) {

	addr, err := net.ResolveTCPAddr(`tcp`, s.Config.ServerGate)
	if err != nil {
		return
	}

	conn, err := net.DialTCP(`tcp`, nil, addr)
	if err != nil {
		return
	}

	s.Conn = conn

	return
}

func (s *Server) genBase() serverBase {
	pid := s.Config.ProjectID
	salt, sign := signature(pid, s.Config.SignatureSecretKey)
	return serverBase{
		PID:  pid,
		Salt: salt,
		Sign: sign,
	}
}

func (s *Server) genBaseMsg() serverBaseMsg {
	return serverBaseMsg{
		MType: 37,
		MID:   genMID(),
	}
}

// Sendmsg 发送消息
func (s *Server) Sendmsg(from, to int64, msg string) (err error) {

	data := serverSendmsg{}

	data.serverBase = s.genBase()
	data.serverBaseMsg = s.genBaseMsg()

	data.From = from
	data.To = to
	data.Msg = msg

	return s.send(`sendmsg`, &data)
}

// Sendroommsg 发送 room 消息
func (s *Server) Sendroommsg(from, rid int64, msg string) (err error) {

	fmt.Println(`Sendroommsg`, from, rid, msg)

	data := serverSendroommsg{}

	data.serverBase = s.genBase()
	data.serverBaseMsg = s.genBaseMsg()

	data.From = from
	data.Rid = rid
	data.Msg = msg

	return s.send(`sendroommsg`, &data)
}

// ServerGettoken ...
func ServerGettoken(uid int64) (token string, err error) {

	s := NewServer()
	err = s.Connect()
	if err != nil {
		return
	}
	defer s.Conn.Close()

	err = s.Gettoken(uid)
	if err != nil {
		return
	}

	j, err := s.Read()
	if err != nil {
		return
	}

	r := serverGettokenReturn{}
	err = msgpack.Unmarshal(j.Content, &r)
	if err != nil {
		return
	}

	return r.Token, nil
}

// ServerSendmsg ...
func ServerSendmsg(from, to int64, msg string) (err error) {
	s := NewServer()
	err = s.Connect()
	if err != nil {
		return
	}
	err = s.Sendmsg(from, to, msg)
	s.Conn.Close()
	return
}

// ServerSendmsgs ...
func ServerSendmsgs(from int64, tos []int64, msg string) (err error) {
	s := NewServer()
	err = s.Connect()
	if err != nil {
		return
	}
	err = s.Sendmsgs(from, tos, msg)
	s.Conn.Close()
	return
}

// Sendmsgs 发送多人消息
func (s *Server) Sendmsgs(from int64, tos []int64, msg string) (err error) {

	data := serverSendmsgs{}

	data.serverBase = s.genBase()
	data.serverBaseMsg = s.genBaseMsg()

	data.From = from
	data.Tos = tos
	data.Msg = msg

	return s.send(`sendmsgs`, &data)
}

// Gettoken 获取 auth token
func (s *Server) Gettoken(uid int64) (err error) {

	data := serverGettoken{}
	data.serverBase = s.genBase()
	data.UID = uid

	return s.send(`gettoken`, &data)
}

func (s *Server) send(method string, data interface{}) (err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	// debug, err := json.Marshal(data)
	// fmt.Println(`debug send`, string(debug), data)

	jsonb, err := msgpack.Marshal(data)
	if err != nil {
		return
	}

	// fmt.Println(`json len`, len(jsonb))

	// xp, err := json.MarshalIndent(data, ``, "\t")
	// fmt.Println(string(xp))

	seq := uint32(time.Now().UnixNano())

	buf := getSendBuffer(MtypeTwoWay, uint8(len(method)))

	binary.Write(&buf, binary.LittleEndian, uint32(len(jsonb))) // size of msg
	binary.Write(&buf, binary.LittleEndian, uint32(seq))        // seq

	buf.WriteString(method) // method
	buf.Write(jsonb)        // msg

	b := buf.Bytes()
	_, err = s.Conn.Write(b)

	fmt.Println(`server send`, err)

	return
}

func (s *Server) Read() (r *Read, err error) {

	ra, _, err := read(s.Conn, nil)
	if err != nil {
		return
	}

	r = ra[0]

	return
}
