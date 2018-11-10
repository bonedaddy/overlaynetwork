# Pubsub example

This example shows how you can use the overlay network's pubsub ability to publish a message to a topic.
Subscribers to the topic set themselves as a subscriber in the DHT allowing other subscribers to find them
and connect to each other forming a mesh. Publishers can then publish to the topic and the message will be
gossiped between all the subscribers. Messages are signed and authenticated against the Peer ID of the peer
who sent the message. 

## Build

From `overlaynetwork` base folder:

```
> go build ./examples/pubsub
```

## Usage

```
> ./dht -l 10000
2017/03/15 14:11:32 listening for connections
2017/03/15 14:11:32 Now run "./pubsub -l 10001 -d /ip4/127.0.0.1/tcp/10000/p2p/12D3KooWAP8mog99pPrF3Sq3WEnzsM2UoucKgfCGiQ2en4wK9SPD" on a different terminal
```

The listener libp2p host will print its `Multiaddress`, which indicates how it can be reached (ip4+tcp) and its randomly generated ID (`12D3KooWA...`)

Now, launch another node that talks to the listener:

```
> ./dht -l 10001 -d /ip4/127.0.0.1/tcp/10000/p2p/12D3KooWAP8mog99pPrF3Sq3WEnzsM2UoucKgfCGiQ2en4wK9SPD
```

The new node will broadcast "I love pizza!" to the "pizza" topic. 

## Details

Notice that all we need to do is create a basic `NodeConfig` and then `OverlayNode`. No additional configuration is needed to
use pubsub.
