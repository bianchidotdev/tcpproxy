package core

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Proxy struct {
	Targets  []string
	Listener net.Listener
	Quit     chan interface{}
}

func MakeProxy(app *AppConfig, port int) *Proxy {
	proxy := &Proxy{Targets: app.Targets, Quit: make(chan interface{})}

	var err error
	proxy.Listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	// asynchronously start listening and return proxy
	go proxy.listenAndProxy(app)
	return proxy
}

func (p *Proxy) listenAndProxy(app *AppConfig) {
	for {
		conn, err := p.Listener.Accept()
		if err != nil {
			select {
			// wait for signal to stop accepting connections
			case <-p.Quit:
				return
			default:
				log.Println("accept error", err)
			}
		}

		// select random target - other strategies could be implemented here, like round robin
		rand.Seed(time.Now().UnixNano())
		// one goroutine per connection (would be nice to have max thread counts)
		go p.handleConnection(conn, rand.Intn(len(app.Targets)), 0)
	}
}

func (p *Proxy) Close(wg *sync.WaitGroup) error {
	// send close signal
	close(p.Quit)
	// close listener - it's possible we might want to wait for existing connections to drain
	err := p.Listener.Close()
	// mark wait group as done for this particular listener
	wg.Done()
	return err
}

func (p *Proxy) selectNextTarget(currentTarget int) int {
	// iterate through sequentially
	nextTarget := currentTarget + 1
	if nextTarget >= len(p.Targets) {
		nextTarget = 0
	}
	return nextTarget
}

func (p *Proxy) handleConnection(clientConn net.Conn, currentTarget int, attempts int) {
	log.Printf("accepted: %s directing to %s", clientConn.RemoteAddr(), p.Targets[currentTarget])

	// start connection with target server
	serverConn, err := net.Dial("tcp", p.Targets[currentTarget])
	if err != nil {
		log.Printf("unable to connect to %s: %s", p.Targets[currentTarget], err)

		nextTarget := p.selectNextTarget(currentTarget)
		// retry up to the # of targets available
		if attempts < len(p.Targets) {
			p.handleConnection(clientConn, nextTarget, attempts+1)
		} else {
			log.Println("Failed to find any healthy targets. Closing connection...")
			clientConn.Close()
		}
		return
	}

	// Ensure we close the server connection
	defer serverConn.Close()

	// Make channels and wait for each individual tcp stream threads to return
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		io.Copy(clientConn, serverConn)
		clientConn.(*net.TCPConn).CloseWrite()
		wg.Done()
	}()
	go func() {
		io.Copy(serverConn, clientConn)
		serverConn.(*net.TCPConn).CloseWrite()
		wg.Done()
	}()

	wg.Wait()

	// Close client connection
	if err = clientConn.Close(); err != nil {
		log.Printf("Close error: %s", err)
	}

	log.Printf("Closing connection at %s", clientConn.RemoteAddr())
}
