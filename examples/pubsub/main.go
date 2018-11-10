package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/overlaynetwork"
	golog "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	gologging "github.com/whyrusleeping/go-logging"
	"io"
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
		log.Printf("Now run \"./pubsub -l %d -d %s\" on a different terminal\n", *listenF+1, fullAddr)

		// Subscribe to the topic "pizza"
		sub, err := node.PubSub.Subscribe(context.Background(), "pizza")
		if err != nil {
			log.Fatal(err)
		}
		for {
			msg, err := sub.Next(context.Background())
			if err == io.EOF || err == context.Canceled {
				break
			} else if err != nil {
				break
			}
			pid, err := peer.IDFromBytes(msg.From)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Received pubsub message: %s from %s\n", string(msg.Data), pid.Pretty())
		}
	}
	/**** This is where the listener code ends ****/

	// Ok now we can bootstrap the node. This could take a little bit if we we're
	// running on a live network.
	err = node.StartOnlineServices()
	if err != nil {
		log.Fatal(err)
	}

	// Publish to the topic "pizza"
	fmt.Println("Publishing message to topic..")
	err = node.PubSub.Publish(context.Background(), "pizza", []byte("I love pizza!"))
	if err != nil {
		log.Fatal(err)
	}
	// hang
	select {}
}
