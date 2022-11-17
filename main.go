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
	Host         string `toml:"host"`
	Port         uint   `toml:"port"`
	User         string `toml:"user"`
	Password     string `toml:"password"`
	IdentityFile string `toml:"identity_file"`
}

func main() {
	var (
		bind         string
		listen       uint
		host         string
		port         uint
		user         string
		password     string
		identityFile string
		configPath   string
	)
	bindUsage := "the bind address"
	bindDefault := ""
	listenUsage := "the listen port"
	listenDefault := uint(8080)
	hostUsage := "the target host (required if no config file)"
	portUsage := "the port to connect"
	portDefualt := uint(22)
	userUsage := "the login user (required if no config file)"
	passwordUsage := "the login password"
	identityFileUsage := "the identity file"
	configPathUsage := "the path of config file (ignore other args if a config file exists)"
	configPathDefualt := "./config.toml"

	flag.StringVar(&bind, "b", bindDefault, bindUsage)
	flag.UintVar(&listen, "l", listenDefault, listenUsage)
	flag.StringVar(&host, "t", "", hostUsage)
	flag.UintVar(&port, "p", portDefualt, portUsage)
	flag.StringVar(&user, "u", "", userUsage)
	flag.StringVar(&password, "s", "", passwordUsage)
	flag.StringVar(&identityFile, "i", "", identityFileUsage)
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
		if password == "" && identityFile == "" {
			log.Fatal("password can not be empty")
		}
		addr := fmt.Sprintf("%s:%d", host, port)
		handler = &sshHandler{addr: addr, user: user, secret: password}
	} else if err != nil {
		log.Fatal("could not parse config file: ", err)
	} else {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		if password != "" {
			handler = &sshHandler{addr: addr, user: cfg.User, secret: cfg.Password}
		} else {
			handler = &sshHandler{addr: addr, user: cfg.User, keyfile: cfg.IdentityFile}
		}
	}

	http.Handle("/", http.FileServer(http.Dir("./front/")))
	http.HandleFunc("/web-socket/ssh", handler.webSocket)
	addr := fmt.Sprintf("%s:%d", bind, listen)
	log.Fatal(http.ListenAndServe(addr, nil))
}
