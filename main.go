package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
)

type config struct {
	Host     string `toml:"host"`
	Port     uint   `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

func main() {
	var (
		host       string
		port       uint
		user       string
		password   string
		configPath string
	)
	hostUsage := "the target host (required if no config file)"
	portUsage := "the port to connect"
	portDefualt := uint(22)
	userUsage := "the login user (required if no config file)"
	passwordUsage := "the login password (required if no config file)"
	configPathUsage := "the path of config file (ignore other args if a config file exists)"
	configPathDefualt := "./config.toml"

	flag.StringVar(&host, "t", "", hostUsage)
	flag.UintVar(&port, "p", portDefualt, portUsage)
	flag.StringVar(&user, "u", "", userUsage)
	flag.StringVar(&password, "s", "", passwordUsage)
	flag.StringVar(&configPath, "c", configPathDefualt, configPathUsage)

	flag.Parse()

	var cfg config
	var handler *sshHandler
	if _, err := toml.DecodeFile(configPath, &cfg); errors.Is(err, os.ErrNotExist) {
		if host == "" {
			log.Fatal("host can not be empty")
		}
		if user == "" {
			log.Fatal("user can not be empty")
		}
		if password == "" {
			log.Fatal("password can not be empty")
		}
		addr := fmt.Sprintf("%s:%d", host, port)
		handler = &sshHandler{addr: addr, user: user, secret: password}
	} else if err != nil {
		log.Fatal("could not parse config file: ", err)
	} else {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		handler = &sshHandler{addr: addr, user: cfg.User, secret: cfg.Password}
	}

	http.HandleFunc("/web-socket/ssh", handler.webSocket)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
