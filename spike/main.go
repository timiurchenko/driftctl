package main

import (
	"encoding/json"
	"io"
	"log"
	"os"

	tfjson "github.com/hashicorp/terraform-json"
)

func main() {
	state := &tfjson.State{}
	file, err := os.Open("driftctl.json")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer file.Close()
	b, _ := io.ReadAll(file)
	err = json.Unmarshal(b, state)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
