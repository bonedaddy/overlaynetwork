# Bitcoin Cash Overlay Network

The Bitcoin Cash Overlay Network is a lightweight overlay network that is intended to be used by any 
Bitcoin Cash application that needs access to a peer-to-peer network for inter-app communication or 
data storage. 

The overlay network is based on Libp2p, a modular network stack that was spun off as part of the IPFS
project. If offers the following features out of the box:

- Extensible peer identities
- Encrypted and authenticated connections
- Protocol multiplexing
- Stream multiplexing
- Multi-transport
- DHT - Key/Value store
- DHT - Peer routing
- DHT - Content routing
- Pubsub

Using the overlay network in your app is dirt simple:
```go
privKey, _, _ := crypto.GenerateEd25519Key(rand.Reader)

cfg := overlaynetwork.NodeConfig{
    PrivateKey: privKey,
    Params: &chaincfg.MainnetParams,
    Port: uint16(4007),
    DataDir: path.Join(os.TempDir(), "overlaynetwork"),
}

node, _ := overlaynetwork.NewOverlayNode(&cfg)
```

From here just define and register your custom protocol:
```go
node.Host.SetStreamHandler("/bitcoincash/mycustomprotocol/1.0.0", func(s net.Stream) {
    // Handle reading from and writing to the stream here
})
```

Dialing other peers is as easy as:
```go
peerID, _ := peer.IDB58Decode("12D3KooWAP8mog99pPrF3Sq3WEnzsM2UoucKgfCGiQ2en4wK9SPD")
stream, _ := node.Host.NewStream(context.Background(), peerID, "/bitcoincash/mycustomprotocol/1.0.0")
// Write to the stream
```

Examples of apps that would benefit from connecting to the overlay network:
- Payment channel protocols
- Coin mixers
- Atomic swap apps
- P2P gambling apps
- Wallet-to-wallet communication