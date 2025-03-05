package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	log.Println("Сервер запущен")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}
