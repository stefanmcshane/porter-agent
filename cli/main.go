package main

import (
	"fmt"
	"log"

	"k8s.io/client-go/rest"
)

func main() {
	conf, err := rest.InClusterConfig()

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(conf.Host)

	for {
		fmt.Println("example log file!")
	}
}
