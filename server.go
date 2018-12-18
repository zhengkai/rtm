package rtm

// https://rasolution-docs.ifunplus.cn/rtm/rtmserver/serverapi/

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"time"
)

// Server 服务器连接
type Server struct {
	Conn   *net.TCPConn
	Config *Config
	seq    uint32
	mid    int64
}

type serverBase struct {
	PID  int32  `json:"pid"`
	Sign string `json:"sign"`
	Salt int64  `json:"salt"`
}

type serverBaseMsg struct {
	MType int8   `json:"mtype"`
	From  int64  `json:"from"`
	MID   int64  `json:"mid"`
	Msg   string `json:"msg"`
	Attrs string `json:"attrs"`
}

type serverSendmsg struct {
	serverBase
	serverBaseMsg
	To int64 `json:"to"`
}

type serverSendmsgs struct {
	serverBase
	serverBaseMsg
	Tos []int64 `json:"tos"`
}

type serverGettoken struct {
	serverBase
	UID int64 `json:"uid"`
}

type serverGettokenReturn struct {
	Token string `json:"token"`
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
	err = json.Unmarshal(j.Content, &r)
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

	jsonb, err := json.Marshal(data)
	if err != nil {
		return
	}

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

	return
}

func (s *Server) Read() (r *Read, err error) {

	r, err = read(s.Conn)

	return
}
