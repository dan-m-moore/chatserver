package main

import (
	"chatserver/config"
	"chatserver/model"
	"chatserver/model/actions"
	"chatserver/model/subs"
	"chatserver/telnetapi"
	"chatserver/webapi"
	"flag"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"strconv"

	gotelnet "github.com/reiver/go-telnet"
)

func main() {
	// All configuration options are contained in the config file
	configFilePath := flag.String("c", "", "config file path")
	flag.Parse()

	// The config file path is required
	if *configFilePath == "" {
		flag.Usage()
		log.Fatalln("error: config file path must be provided")
	}

	// Parse the config file
	config, err := config.ParseFile(*configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Print the parsed config
	log.Println("Welcome to chatserver!")
	log.Println("----------------------")
	log.Println("Serving telnet on port", config.TelnetPort)
	log.Println("Serving web client on port", config.WebPort)
	log.Println("Web client path:", config.WebClientPath)
	log.Println("Log file path:", config.LogFilePath)

	// Create the actions Replayer and Logger as needed (determined by the log file path)
	var actionsReplayer model.ActionsReplayer
	var actionsLogger actions.Actor
	if config.LogFilePath != "" {
		// If the file doesn't exist, then don't try to replay it
		_, err := os.Stat(config.LogFilePath)
		if err == nil {
			actionsReplayer, err = actions.NewReplayer(config.LogFilePath)
			if err != nil {
				log.Fatal(err)
			}
		}

		actionsLogger, err = actions.NewLogger(config.LogFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Create/Initialize the model
	subsEngine := subs.NewEngine()
	model, err := model.NewModel(actionsReplayer, actionsLogger, subsEngine)
	if err != nil {
		log.Fatal(err)
	}

	// Serve telnet
	telnetHandler := telnetapi.NewConnectionHandler(model, subsEngine)
	telnetPort := ":" + strconv.Itoa(config.TelnetPort)
	go func() {
		err := gotelnet.ListenAndServe(telnetPort, telnetHandler)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Set up JSON RPC
	err = rpc.RegisterName("chatserver", webapi.NewInstance(model))
	if err != nil {
		log.Fatal(err)
	}
	webapiHandler := webapi.NewConnectionHandler(subsEngine)

	// Serve HTTP
	http.Handle("/", http.FileServer(http.Dir(config.WebClientPath)))
	http.Handle("/ws", webapiHandler)
	webPort := ":" + strconv.Itoa(config.WebPort)
	err = http.ListenAndServe(webPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
