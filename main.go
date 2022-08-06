package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"github.com/stokkelol/pravdabot/cmd"
)

func main() {
	_ = godotenv.Load()
	viper.AutomaticEnv()
	client, err := cmd.New()
	if err != nil {
		log.Fatalf("Error creating service: %s", err.Error())
	}
	client.Run()
}
