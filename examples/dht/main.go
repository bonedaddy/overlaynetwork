package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/overlaynetwork"
	golog "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peerstore"
	gologging "github.com/whyrusleeping/go-logging"
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

	// If the target address is provided let's add it as a bootstrap peer in the config
	if *target != "" {
		// Parse the target address
		peerInfo, err := overlaynetwork.ParseBootstrapPeer(*target)
		if err != nil {
			log.Fatal(err)
		}
		cfg.BootstrapPeers = []peerstore.PeerInfo{peerInfo}
	}

	// Now create our node object
	node, err := overlaynetwork.NewOverlayNode(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// If this is the listening dht node then just hang here.
	if *target == "" {
		log.Println("listening for connections")
		fullAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", *listenF, node.Host.ID().Pretty())
		log.Printf("Now run \"./dht -l %d -d %s\" on a different terminal\n", *listenF+1, fullAddr)
		select {}
	}
	/**** This is where the listener code ends ****/

	// Ok now we can bootstrap the node. This could take a little bit if we we're
	// running on a live network.
	err = node.StartOnlineServices()
	if err != nil {
		log.Fatal(err)
	}

	// Create a DHT entry. In this instance the key must be the hex encoded sha256 hash
	// of the value.
	value := []byte("Hello World!")
	valueHash := sha256.Sum256(value)
	key := hex.EncodeToString(valueHash[:])

	// Put the key/value to the DHT
	fmt.Println("Putting value to the DHT")
	err = node.Routing.PutValue(context.Background(), fmt.Sprintf("/sha256/%s", key), value)
	if err != nil {
		log.Fatal(err)
	}

	// Look up the key in the DHT and see if we get back the same value.
	returnedValue, err := node.Routing.GetValue(context.Background(), fmt.Sprintf("/sha256/%s", key))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got value from DHT: %s\n", string(returnedValue))
}