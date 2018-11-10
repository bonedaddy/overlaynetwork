package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/overlaynetwork"
	golog "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peerstore"
	gologging "github.com/whyrusleeping/go-logging"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
)

func main() {
	// LibP2P code uses golog to log messages. They log with different
	// string IDs (i.e. "swarm"). We can control the verbosity level for
	// all loggers with:
	golog.SetAllLoggers(gologging.INFO) // Change to DEBUG for extra info

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	// First let's create a new identity key pair for our node. If this was your
	// application you would likely save this private key to a database and load
	// it from the db on subsequent start ups.
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	// Next we'll create the node config
	cfg := overlaynetwork.NodeConfig{
		PrivateKey: privKey,

		// We'll use testnet for this example.
		Params: &chaincfg.TestNet3Params,

		// We'll also disable querying DNS seeds to prevent this example from actually
		// connecting to the overlay network. For this example, we're just going to
		// connect two nodes locally.
		DisableDNSSeeds: true,

		Port: uint16(*listenF),

		// You would also set a directory here to use as the data directory for this node.
		// For this we will just use a temp directory.
		DataDir: path.Join(os.TempDir(), strconv.Itoa(*listenF)),
	}

	// Now create our node object
	node, err := overlaynetwork.NewOverlayNode(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// This is where we will register our custom protocol. In this case we are using
	// /bitcoincash/echo/1.0.0. Not that the /bitcoincash/ prefix denotes that this
	// protocol is intended to run on the bitcoin cash overlay network.
	//
	// The function we will set here will fire whenever our node accepts a new stream.
	// Note that a stream is NOT a connection. Once a node has established an open
	// connection he may open an unlimited number of streams within that connection
	// and multiplex data over those streams. This callback fires only when a new
	// stream is accepted. It's up to you to decide how you want to handle streams
	// in your protocol. One stream per connection? One stream per message? One stream
	// per request/reply message pair? It's up to you.
	//
	// Finally it's up to you to decide what wire serialization you are going to use
	// and how it will be delimited.
	node.Host.SetStreamHandler("/bitcoincash/echo/1.0.0", func(s net.Stream) {
		log.Println("Got a new stream!")
		buf := bufio.NewReader(s)
		str, err := buf.ReadString('\n')
		if err != nil {
			s.Close()
		}
		log.Printf("read: %s\n", str)
		s.Write([]byte(str))
		s.Close()
	})

	// If this is the listening echo node then just hang here.
	if *target == "" {
		log.Println("listening for connections")
		fullAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", *listenF, node.Host.ID().Pretty())
		log.Printf("Now run \"./echo -l %d -d %s\" on a different terminal\n", *listenF+1, fullAddr)
		select {}
	}
	/**** This is where the listener code ends ****/

	// Parse the target address
	peerInfo, err := overlaynetwork.ParseBootstrapPeer(*target)
	if err != nil {
		log.Fatal(err)
	}

	// We have a peer ID and a targetAddr so we add it to the peerstore
	// so LibP2P knows how to contact it.
	node.Host.Peerstore().AddAddr(peerInfo.ID, peerInfo.Addrs[0], peerstore.PermanentAddrTTL)

	log.Println("opening stream")
	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set above because
	// we use the same /bitcoincash/echo/1.0.0 protocol
	s, err := node.Host.NewStream(context.Background(), peerInfo.ID, "/bitcoincash/echo/1.0.0")
	if err != nil {
		log.Fatalln(err)
	}

	_, err = s.Write([]byte("Hello, world!\n"))
	if err != nil {
		log.Fatalln(err)
	}

	out, err := ioutil.ReadAll(s)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("read reply: %q\n", out)
}
