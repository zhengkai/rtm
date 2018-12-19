package rtm

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"

	"github.com/zhengkai/mpp"
)

// Client 客户端连接
type Client struct {
	Conn     *net.TCPConn
	Config   *ConfigClient
	Endpoint string
	seq      uint32
	mid      int64
	remain   []byte // 上一次读取不完整的信息，在下一次读取时填到头部
}

type clientWhich struct {
	PID      int32  `json:"pid"`
	UID      int64  `json:"mid"`
	What     string `json:"what"`
	AddrType string `json:"addrType"`
	Version  int32  `json:"version"`
}

type clientWhichReturn struct {
	Endpoint string `json:"endpoint"`
}

// Read 从 rtm 读到的数据
type Read struct {
	Mtype   Mtype
	Content []byte
	Method  string
	Seq     []byte
}

type clientBaseMsg struct {
	MID   int64  `json:"mid"`
	MType int8   `json:"mtype"`
	Msg   string `json:"msg"`
	Attrs string `json:"attrs"`
}

type clientSendmsg struct {
	clientBaseMsg
	To int64 `json:"to"`
}

// ClientPushmsg ...
type ClientPushmsg struct {
	From  int64  `json:"from"`
	Mtype int8   `json:"mtype"`
	Ftype int8   `json:"ftype"`
	MID   int64  `json:"mid"`
	Msg   string `json:"msg"`
	Attrs string `json:"attrs"`
}

func (c *Client) genBaseMsg() clientBaseMsg {
	return clientBaseMsg{
		MType: 37,
		MID:   genMID(),
	}
}

// NewClient 创建信 Client 端
func NewClient(uid int64) *Client {
	if globalConfig == nil {
		return nil
	}

	return &Client{
		Config: &ConfigClient{
			ProjectID:  globalConfig.ProjectID,
			UID:        uid,
			ClientGate: globalConfig.ClientGate,
		},
	}
}

// Connect 创建连接
func (c *Client) Connect() (err error) {

	c.remain = nil

	err = c.connect(c.Config.ClientGate)
	if err != nil {
		return
	}

	err = c.which()
	if err != nil {
		return
	}

	err = c.auth()
	return
}

func (c *Client) auth() (err error) {

	token, err := ServerGettoken(c.Config.UID)
	if err != nil {
		return
	}

	d := clientAuth{
		PID:     c.Config.ProjectID,
		UID:     c.Config.UID,
		Token:   token,
		Version: 1,
	}

	err = c.send(`auth`, &d)
	return
}

func (c *Client) which() (err error) {

	d := clientWhich{
		PID:      c.Config.ProjectID,
		UID:      c.Config.UID,
		What:     `rtmGated`,
		AddrType: `ipv4`,
		Version:  1,
	}

	err = c.send(`which`, &d)

	data, err := c.Read()
	c.Conn.Close()
	if err != nil {
		return
	}

	r := clientWhichReturn{}
	err = json.Unmarshal(data[0].Content, &r)
	if err != nil {
		return
	}

	c.Endpoint = r.Endpoint

	err = c.connect(r.Endpoint)

	return
}

type clientAuth struct {
	PID     int32  `json:"pid"`
	UID     int64  `json:"uid"`
	Token   string `json:"token"`
	Version int32  `json:"version"`
}

func (c *Client) connect(endpoint string) (err error) {

	addr, err := net.ResolveTCPAddr(`tcp`, endpoint)
	if err != nil {
		return
	}

	// fmt.Println(`client addr`, endpoint, addr)

	conn, err := net.DialTCP(`tcp`, nil, addr)
	if err != nil {
		return
	}

	c.Conn = conn
	return
}

func jsonMarshal(data interface{}) (r []byte, err error) {
	if data == nil {
		r = []byte(`{}`)
	} else {
		r, err = json.Marshal(data)
	}
	return
}

func (c *Client) send(method string, data interface{}) (err error) {

	jsonb, err := jsonMarshal(data)
	if err != nil {
		return
	}

	buf := getSendBuffer(MtypeTwoWay, uint8(len(method)))

	binary.Write(&buf, binary.LittleEndian, uint32(len(jsonb))) // size of msg

	c.seq++
	if c.seq >= 4294967295 {
		c.seq = 1
	}

	binary.Write(&buf, binary.LittleEndian, uint32(c.seq)) // seq

	buf.WriteString(method) // method
	buf.Write(jsonb)        // msg

	b := buf.Bytes()

	_, err = c.Conn.Write(b)

	return
}

// Sendmsg 发送消息
func (c *Client) Sendmsg(to int64, msg string) (err error) {

	data := clientSendmsg{}

	data.clientBaseMsg = c.genBaseMsg()

	data.Msg = msg
	data.To = to

	return c.send(`sendmsg`, &data)
}

func (c *Client) answer(seq []byte, data interface{}) (err error) {

	jsonb, err := jsonMarshal(data)
	if err != nil {
		return
	}

	buf := getSendBuffer(MtypeAnswer, 0)

	binary.Write(&buf, binary.LittleEndian, uint32(len(jsonb))) // size of msg

	buf.Write(seq)   // seq
	buf.Write(jsonb) // msg

	b := buf.Bytes()

	_, err = c.Conn.Write(b)

	return
}

// Read 读
func (c *Client) Read() (ra []*Read, err error) {

	var remain []byte
	ra, remain, err = read(c.Conn, c.remain)
	if err != nil {
		return
	}
	c.remain = remain

	if len(remain) > 0 {
		fmt.Println(`remain`, len(remain), remain, c)
	}

	for _, r := range ra {

		if !c.parse(r) {
			continue
		}

		if r.Method == `pushmsg` {
			buf := mpp.ToJson(r.Content)
			r.Content = buf.Bytes()
		}
	}

	return
}

func (c *Client) parse(r *Read) (isReturn bool) {

	if r.Method == `ping` {
		c.answer(r.Seq, nil)
		return false
	}

	return true
}

// GetPushmsg ...
func GetPushmsg(b []byte) (r ClientPushmsg, err error) {
	r = ClientPushmsg{}
	err = json.Unmarshal(b, &r)
	return
}