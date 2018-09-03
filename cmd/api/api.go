package main

import (
	"log"

	"github.com/ryanjyoder/inasnap"
)

func main() {
	configs := inasnap.Configs{
		CouchURL:      "http://localhost:5984/",
		CouchUser:     "admin",
		CouchPassword: "admin",
		CouchDBName:   "snaps",
		ListenPort:    "8080",
	}
	s, err := inasnap.NewAPIServer(configs)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.Run())

}
