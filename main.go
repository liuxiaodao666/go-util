package main

import (
	"net/http"

	"github.com/liuxiaodao666/go-util/logger"
	_ "github.com/liuxiaodao666/go-util/pprof"
)

func main() {
	logger.Errorf("mock %s", "error")
	http.Get("https://www.baidu.com")

}
