package main

import (
	"log"
	"net/url"
	"os"

	"github.com/ryanjyoder/couchdb"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Not enough arguments. Must provide directory to seed from")
	}
	dbURL, err := url.Parse("http://localhost:5984/")
	check(err)

	client, err := couchdb.NewAuthClient("admin", "admin", dbURL)
	check(err)

	db := client.Use("snaps")
	desginDocs, err := client.Parse(os.Args[1])
	check(err)
	db.Seed(desginDocs)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
