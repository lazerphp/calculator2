package main

import (
	"log"
	app "orchestrator/internal/app"
)

func main() {
	log.Println("оркестратор в деле")
	app.Orchestrate()
}
