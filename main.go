package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"messer/server"
)

//go:embed dist
var wwwRoot embed.FS

var subCommand string

func init() {
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		subCommand = os.Args[1]
		os.Args = append(os.Args[0:1], os.Args[2:]...) // remove os.Args[1]
	}
}

func main() {
	subCmds := map[string]func(){
		"createUser": createUser,
	}

	if subCommand != "" {
		if cmd, ok := subCmds[subCommand]; ok {
			cmd()
			return
		}
		log.Fatalf("Command not found: %v", subCommand)
	}

	var help bool
	flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&help, "help", false, "show help")
	flag.Parse()

	if help {
		fmt.Println("Env: CONFIG_FILE=/path/to/config/file, default=.env")
		fmt.Println("Sub commands:")
		for k := range subCmds {
			fmt.Printf("\t%v\n", k)
		}
		return
	}

	conf := server.Config()
	log.Printf("config file: %v", conf.ConfigFile)
	if err := conf.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	srv, err := server.New(http.FS(wwwRoot))
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	errs := make(chan error, 2)
	go func() {
		log.Printf("starting server...")
		errs <- srv.Start()
	}()
	go func() {
		c := make(chan os.Signal, 3)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		errs <- srv.Stop()
	}()

	if err := <-errs; err != nil {
		log.Fatal(err)
	}
}

func createUser() {
	var username, password, name string
	flag.StringVar(&username, "username", "", "Username")
	flag.StringVar(&password, "password", "", "Password")
	flag.StringVar(&name, "name", "", "Name")
	flag.Parse()

	srv, err := server.New(http.FS(wwwRoot))
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	if err := srv.CreateUser(username, password, name); err != nil {
		log.Fatalf("failed to create user: %v", err)
	}
	log.Printf("user %v has been created", username)
}
