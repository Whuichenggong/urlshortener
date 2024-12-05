package main

import "github.com/Whuichenggong/urlshortener/urlshortener/application"

func main() {
	a := application.Application{}
	if err := a.Init("./config/config.yaml"); err != nil {
		panic(err)
	}
	a.RunServer()
}
