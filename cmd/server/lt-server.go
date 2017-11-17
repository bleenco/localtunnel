package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bleenco/localtunnel"
)

var port = flag.String("p", "1234", "http listen port")
var domain = flag.String("d", "local.host", "server domain")

func main() {
	flag.Parse()
	stop := make(chan os.Signal, 1)

	httpServer := localtunnel.SetupServer(*port, *domain)

	go func() {
		fmt.Printf("Listening on http://0.0.0.0%s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("Error: %s", err)
			return
		}
	}()

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Printf("Shutting down...")
	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Server stopped.")
	}
}
