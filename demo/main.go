package main

import (
	"fmt"
	"math/rand"
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

	go bot(123, 321)
	bot(321, 123)
}

func bot(id int64, target int64) {

	c := rtm.NewClient(id)
	err := c.Connect()
	if err != nil {
		fmt.Println(`client connect err`, err)
		return
	}

	i := 0

	go func() {
		for {

			time.Sleep(time.Second)

			c.Sendmsg(target, `test`)
		}
	}()

	fmt.Println(`start read`)
	for {
		r, err := c.Read()

		if err != nil {
			fmt.Println(`loop read end`)
			break
		}

		if r == nil || r.Method == `ping` {
			i++
			fmt.Println(`ping`, i)
			continue
		}

		if r.Method == `` {
			r.Method = `[answer]`
		}
		fmt.Println(`client read`, r.Method, len(r.Content))

		fmt.Println(string(r.Content))

		if err != nil {

			fmt.Println(`client err`, err)
			return
		}
		fmt.Println(`loop`)
	}

}
