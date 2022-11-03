package main

import (
	"encoding/json"
	"os"
	"proj1/server"
	"strconv"
)

func main() {
	config := server.Config{}

	// Parallel version
	if len(os.Args) == 2 {
		config.Mode = "p"
		num, _ := strconv.Atoi(os.Args[1])
		config.ConsumersCount = num
	} else { // Sequential version
		config.Mode = "s"
	}

	// Set encoder and decoder
	config.Decoder = json.NewDecoder(os.Stdin)
	config.Encoder = json.NewEncoder(os.Stdout)

	// Run server
	server.Run(config)
}
