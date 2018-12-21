package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/zhengkai/rtm"
)

func server() {

	port := flag.Int("port", 21003, "http port")
	flag.Parse()

	addr := `127.0.0.1:` + strconv.Itoa(*port)

	mux := http.NewServeMux()
	mux.HandleFunc(`/api/token`, getToken)
	fmt.Printf("port = %s\n", addr)
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		fmt.Println(`http`, addr, `start fail:`)
		fmt.Println(err.Error())
	}
}

func getToken(w http.ResponseWriter, r *http.Request) {

	token, _ := rtm.ServerGettoken(654321)

	b := []byte(token)

	w.Write(b)
}
