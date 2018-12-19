package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/zhengkai/rtm"
)

func main() {

	rand.Seed(time.Now().UTC().UnixNano())

	rtm.SetConfig(&rtm.Config{
		ProjectID:          configProjectID,
		SignatureSecretKey: configSignatureSecretKey,
		ServerGate:         configServerGate,
		ClientGate:         configClientGate,
	})

	for i := int64(0); i < 10; i++ {
		go archer(20000 + i)
	}

	bot(10001)
}

// 负责发送消息
func archer(id int64) {

	target := int64(10001)

	c := rtm.NewClient(id)
	err := c.Connect()
	if err != nil {
		fmt.Println(`client connect err`, err)
		return
	}

	i := 0

	for {

		i++

		time.Sleep(time.Second)

		c.Sendmsg(target, `test `+strconv.Itoa(i))
	}
}

// 接收遍历显示所有消息
func bot(id int64) {

	c := rtm.NewClient(id)
	err := c.Connect()
	if err != nil {
		fmt.Println(`client connect err`, err)
		return
	}

	fmt.Println(`start read`)
	for {
		ra, err := c.Read()
		if err != nil {
			fmt.Println(`read error:`, err)
			break
		}

		for x, r := range ra {

			fmt.Println(`#`, x)

			if r.Method == `ping` || r.Method == `` {
				continue
			}

			if r.Method == `pushmsg` {
				s, _ := rtm.GetPushmsg(r.Content)
				fmt.Println(`pushmsg`, s.From, s.Msg)
			} else {
				fmt.Println(string(r.Content))
			}

			if err != nil {

				fmt.Println(`client err`, err)
				return
			}
		}
		// fmt.Println(`loop`)
	}
}
