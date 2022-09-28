package main

import (
	"log"

	"k8s.io/client-go/rest"
)

func main() {
	conf, err := rest.InClusterConfig()

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(conf.Host)
}
