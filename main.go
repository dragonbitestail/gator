package main

import (
    "fmt"
		"log"
		gtc "gator/internal/config"

)


func main() {

	config := readConfig()
	config.SetUser("dbt")
	config = readConfig()
	fmt.Printf("%+v\n", config)

}


func readConfig() *gtc.Config {
	config, err := gtc.Read()
	if err != nil {
		log.Fatal(err)
	}

	return config
}
