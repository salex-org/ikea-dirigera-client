package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/salex-org/ikea-dirigera-client/pkg/client"
	"github.com/zalando/go-keyring"
)

var (
	currentContext = Context{
		Name:    "salex-lab",
		Address: "192.168.1.148",
		Port:    8443,
	}
)

type Context struct {
	Name          string
	Address       string
	Port          int
	Authorization client.Authorization
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ikea <command>")
		return
	}

	switch os.Args[1] {
	case "scan":
		listHubs()
	case "authorize":
		err := authorize(&currentContext)
		if err != nil {
			return
		}
		err = addContext(currentContext)
		if err != nil {
			return
		}
	case "devices":
		err := initialize(&currentContext)
		if err != nil {
			log.Fatal(err)
		}
		listDevices(currentContext)
	case "listen":
		err := initialize(&currentContext)
		if err != nil {
			log.Fatal(err)
		}
		listenForEvents(currentContext)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
	}
}

func listHubs() {
	hubs, err := client.Scan()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d hubs:\n", len(hubs))
	for i, hub := range hubs {
		fmt.Printf("%d. Hub %s at %s port %d\tSerial-# %s\tFirmware %s\n", i+1, hub.HostName, hub.Address, hub.Port, hub.SerialNumber, hub.FirmwareVersion)
	}
}

func listenForEvents(cliContext Context) {
	// Notification context for reacting on process termination - used by shutdown function
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Waiting group used to await finishing the shutdown process when stopping
	var wait sync.WaitGroup

	dirigeraClient := client.Connect(cliContext.Address, cliContext.Port, &cliContext.Authorization)
	dirigeraClient.RegisterEventHandler(func(event client.Event) {
		fmt.Printf("Event received: %v\n", event)
	}, "deviceStateChanged")

	// Loop function for event listening
	fmt.Printf("Start listening for events in %s...\n", cliContext.Name)
	wait.Add(1)
	go func() {
		defer wait.Done()
		err := dirigeraClient.ListenForEvents()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Shutdown function waiting for the SIGTERM notification to stop event listening
	wait.Add(1)
	go func() {
		defer wait.Done()
		<-ctx.Done()
		fmt.Printf("\n\U0001F6D1 Stop listening for events\n")
		err := dirigeraClient.StopEventListening()
		if err != nil {
			log.Fatal(err)
		}
	}()

	wait.Wait()
	fmt.Printf("\U0001F3C1 Shutdown finished\n")
	os.Exit(0)
}

func listDevices(cliContext Context) {
	dirigeraClient := client.Connect(cliContext.Address, cliContext.Port, &cliContext.Authorization)
	devices, err := dirigeraClient.ListDevices()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d devices in %s:\n", len(devices), cliContext.Name)
}

func initialize(context *Context) error {
	var err error
	context.Authorization.AccessToken, err = keyring.Get("ikea-sh-"+context.Name, "access-token")
	if err != nil {
		return fmt.Errorf("error reading access token from keyring: %w", err)
	}
	context.Authorization.TLSFingerprint, err = keyring.Get("ikea-sh-"+context.Name, "tls-fingerprint")
	if err != nil {
		return fmt.Errorf("error reading TLS fingerprint from keyring: %w", err)
	}

	return nil
}

func generateClientName() string {
	hostname, err := os.Hostname()
	hostname = strings.SplitN(hostname, ".", 2)[0]
	if err != nil {
		return uuid.New().String()
	}
	me, err := user.Current()
	if err != nil {
		return uuid.New().String()
	}

	return fmt.Sprintf("%s@%s", me.Username, hostname)
}

func authorize(context *Context) error {
	clientName := generateClientName()
	fmt.Printf("Adding new user %s to hub at %s\n", clientName, context.Address)
	var err error
	context.Authorization, err = client.Authorize(context.Address, context.Port, clientName, func() {
		fmt.Printf("Please press the button on the backside of the hub.")
	}, func() {
		fmt.Printf(".")
	})
	if err != nil {
		fmt.Printf("failed: %v\n", err)

		return fmt.Errorf("error authorizing new client: %w", err)
	}
	fmt.Printf("success\n")

	return nil
}

func addContext(context Context) error {
	err := keyring.Set("ikea-sh-"+context.Name, "access-token", context.Authorization.AccessToken)
	if err != nil {
		return fmt.Errorf("error storing access token in keyring: %w", err)
	}
	err = keyring.Set("ikea-sh-"+context.Name, "tls-fingerprint", context.Authorization.TLSFingerprint)
	if err != nil {
		return fmt.Errorf("error storing TLS fingerprint in keyring: %w", err)
	}
	fmt.Printf("Access token and TLS fingerprint added to keyring\n")

	return nil
}
