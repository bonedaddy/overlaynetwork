package overlaynetwork

import (
	"context"
	"fmt"
	"github.com/gcash/bchd/chaincfg"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/opts"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p-routing"
	"net"
	"path"
)

var (
	// ProtocolDHTMainnet defines the protocol ID for the DHT. We are prefixing it
	// with /bitcoincash/ to avoid the DHT accidentally merging with other
	// libp2p DHTs. The /mainnet/ path is used to segregate the network from testnet.
	ProtocolDHTMainnet = protocol.ID("/bitcoincash/mainnet/kad/1.0.0")

	// ProtocolDHTTestnet3 defines the protocol ID for the DHT. We are prefixing it
	// with /bitcoincash/ to avoid the DHT accidentally merging with other
	// libp2p DHTs. The /testnet3/ path is used to segregate the network from mainnet.
	ProtocolDHTTestnet3 = protocol.ID("/bitcoincash/testnet3/kad/1.0.0")
)

// OverlayNode represents our node in the overlay network. It is
// capable of making direct connections to other peers in the overlay and
// maintaining a kademlia DHT for the purpose of resolving peerIDs into
// network addresses.
type OverlayNode struct {

	// Params represents the Bitcoin Cash network that this node will be using.
	Params *chaincfg.Params

	// Host is the main libp2p instance which handles all our networking.
	// It will require some configuration to set it up. Once set up we
	// can register new protocol handlers with it.
	Host host.Host

	// Routing is a routing implementation which implements the PeerRouting,
	// ContentRouting, and ValueStore interfaces. In practice this will be
	// a Kademlia DHT.
	Routing routing.IpfsRouting

	// PubSub is an instance of gossipsub which uses the DHT save lists of
	// subscribers to topics which publishers can find via a DHT query and
	// publish messages to the topic using a gossip mechanism.
	PubSub *Pubsub

	// PrivateKey is the identity private key for this node
	PrivateKey crypto.PrivKey

	// Datastore is a datastore implementation that we will use to store routing
	// data.
	Datastore datastore.Datastore

	bootstrapPeers   []peerstore.PeerInfo
	disableDNSSeeeds bool
}

// NewOverlayNode is a constructor for our Node object
func NewOverlayNode(config *NodeConfig) (*OverlayNode, error) {
	opts := []libp2p.Option{
		// Listen on all interface on both IPv4 and IPv6.
		// If we're going to enable other transports such as Tor or QUIC we would do it here.

		// TODO: users who start in Tor mode will have their privacy blown if they use this
		// before getting around to implementing Tor. For now we should probably check if
		// the wallet was started in Tor mode and panic if payment channels are enabled.
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.Port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/tcp/%d", config.Port)),
		libp2p.Identity(config.PrivateKey),
	}

	// This function will initialize a new libp2p host with our options plus a bunch of default options
	// The default options includes default transports, muxers, security, and peer store.
	peerHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// Create a leveldb datastore
	dstore, err := leveldb.NewDatastore(path.Join(config.DataDir, "libp2p"), nil)
	if err != nil {
		return nil, err
	}

	protocol := ProtocolDHTTestnet3
	if config.Params == &chaincfg.MainNetParams {
		protocol = ProtocolDHTMainnet
	}

	// Create the DHT instance. It needs the host and a datastore instance.
	routing, err := dht.New(
		context.Background(), peerHost,
		dhtopts.Datastore(dstore),
		dhtopts.Protocols(protocol),
		dhtopts.Validator(record.NamespacedValidator{
			"pk":     record.PublicKeyValidator{},
			"sha256": &Sha256Validator{},
		}),
	)

	ps, err := pubsub.NewGossipSub(context.Background(), peerHost)
	if err != nil {
		return nil, err
	}

	node := &OverlayNode{
		Params:           config.Params,
		Host:             peerHost,
		Routing:          routing,
		PubSub:           &Pubsub{ps: ps, ht: peerHost, rt: routing},
		PrivateKey:       config.PrivateKey,
		Datastore:        dstore,
		bootstrapPeers:   config.BootstrapPeers,
		disableDNSSeeeds: config.DisableDNSSeeds,
	}
	return node, nil
}

// StartOnlineServices will bootstrap the peer host using the provided bootstrap peers. Once the host
// has been bootstrapped it will proceed to bootstrap the DHT.
func (n *OverlayNode) StartOnlineServices() error {
	peers := n.bootstrapPeers
	if !n.disableDNSSeeeds {
		// TODO: we don't want to do this in the clear if we're using Tor. We need to
		// investigate if we can lookup a TXT record over Tor.
		peerChan := SeedFromDNS(n.Params, net.LookupTXT)
		for pi := range peerChan {
			peers = append(peers, pi)
		}
	}
	return Bootstrap(n.Routing.(*dht.IpfsDHT), n.Host, bootstrapConfigWithPeers(peers))
}

// Shutdown will cancel the context shared by the various components which will shut them all down
// disconnecting all peers in the process.
func (n *OverlayNode) Shutdown() {
	n.Host.Close()
}
