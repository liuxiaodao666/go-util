package main

import (
	"github.com/liuxiaodao666/go-util/logger"
	_ "github.com/liuxiaodao666/go-util/pprof"
	"net"
	"net/http"
)

func main() {
	logger.Errorf("mock %s", "error")
	http.Get("https://www.baidu.com")
	net.Dial()

}
