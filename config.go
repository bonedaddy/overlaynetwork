package overlaynetwork

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"errors"
	"github.com/gcash/bchd/chaincfg"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/yawning/bulb"
	"github.com/yawning/bulb/utils/pkcs1"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// NodeConfig contains basic configuration information that we'll need to
// start our node.
type NodeConfig struct {
	// Params represents the Bitcoin Cash network that this node will be using.
	Params *chaincfg.Params

	// Port specifies the port use for incoming connections. Defaults to 8005.
	Port uint16

	// DisableDnsSeeds will disable querying the DNS seeds for bootstrap addresses
	DisableDNSSeeds bool

	// BootstrapPeers is an optional list of peers to use for bootstrapping
	// the DHT and connecting to the network.
	BootstrapPeers []peerstore.PeerInfo

	// PrivateKey is the key to initialize the node with. Typically
	// this will be persisted somewhere and loaded from disk on
	// startup.
	PrivateKey crypto.PrivKey

	// DataDir is the path to a directory to store node data.
	DataDir string

	// ConnectionType specifies Clearnet, Tor, or Dualstack mode.
	ConnectionType TorMode

	// TorControlPort specifies the Tor control port to connect to if using
	// a Tor mode. If no port is specified we will try figure out which port
	// to connect to.
	TorControlPort uint16

	// TorControlPassword is the Tor control password which is required if using
	// a Tor mode.
	TorControlPassword string

	// TorListeningPort is the port on which to listen for incoming onion connections.
	// Defaults to 8006.
	TorListeningPort uint16
}

// TorMode specifies whether and how the overlay uses the Tor network
type TorMode int

const (
	// TmClearnet mode will not use Tor at all and all connections will
	// be over the clear internet.
	TmClearnet TorMode = 0

	// TmTorOnly mode will only make outgoing and incoming connections over Tor.
	TmTorOnly  TorMode = 1

	// TmDualStack mode will accept incoming connections over Tor in addition to
	// over the clear internet. All outgoing connections will use the clear internet
	// by default unless connecting to a TorOnly node.
	TmDualStack TorMode = 2
)

// getTorControlPort returns the default Tor control port if Tor is running on the
// default port or an error.
func getTorControlPort() (uint16, error) {
	conn, err := bulb.Dial("tcp4", "127.0.0.1:9151")
	if err == nil {
		conn.Close()
		return 9151, nil
	}
	conn, err = bulb.Dial("tcp4", "127.0.0.1:9051")
	if err == nil {
		conn.Close()
		return 9051, nil
	}
	return 0, errors.New("tor control unavailable")
}

// createHiddenServiceKey will generate a new RSA key and onion address and save it to the repo
func createHiddenServiceKey(repoPath string) (onionAddr string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", err
	}
	id, err := pkcs1.OnionAddr(&priv.PublicKey)
	if err != nil {
		return "", err
	}

	f, err := os.Create(path.Join(repoPath, id+".onion_key"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	privKeyBytes, err := pkcs1.EncodePrivateKeyDER(priv)
	if err != nil {
		return "", err
	}

	block := pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKeyBytes}
	err = pem.Encode(f, &block)
	if err != nil {
		return "", err
	}
	return id, nil
}

// maybeCreateHiddenServiceKey will generate a new key pair if one does not already exist
func maybeCreateHiddenServiceKey(repoPath string) (onionAddr string, err error) {
	d, err := os.Open(repoPath)
	if err != nil {
		return "", err
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".onion_key" {
			addr := strings.Split(file.Name(), ".onion_key")
			return addr[0], nil
		}
	}

	return createHiddenServiceKey(repoPath)
}