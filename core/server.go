package core

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

func Serve(config *ProxyConfig) {
	// set up wait group to ensure we don't end the program while listeners are still running
	wg := &sync.WaitGroup{}
	var proxies []Proxy

	for _, app := range config.Apps {
		for _, port := range app.Ports {
			wg.Add(1)
			proxy := MakeProxy(app, port)
			proxies = append(proxies, *proxy)
		}
	}

	// Wait for a SIGINT (perhaps triggered by user with CTRL-C)
	// It'd be nice to do a force quit on a second ctrl-c, but looks like you can use ctrl-\
	// Close all existing server connections
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		log.Println("Received an interrupt, stopping services...")
		for _, proxy := range proxies {
			log.Println("Closing listener at", proxy.Listener.Addr())
			proxy.Close(wg)
		}
	}()

	wg.Wait()
}
