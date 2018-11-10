# DHT example

This example shows how you can use the overlay network's DHT as a basic key/value store. In this instance the 
keys must be the hex encoded sha256 hash of the value. Notice no additional configuration is needed beyond what 
we used in the echo example. In this example we will spin up two nodes. One will listen on the network for DHT
events and the other will set and get the value.

## Build

From `overlaynetwork` base folder:

```
> go build ./examples/dht
```

## Usage

```
> ./dht -l 10000
2017/03/15 14:11:32 listening for connections
2017/03/15 14:11:32 Now run "./dht -l 10001 -d /ip4/127.0.0.1/tcp/10000/p2p/12D3KooWAP8mog99pPrF3Sq3WEnzsM2UoucKgfCGiQ2en4wK9SPD" on a different terminal
```

The listener libp2p host will print its `Multiaddress`, which indicates how it can be reached (ip4+tcp) and its randomly generated ID (`12D3KooWA...`)

Now, launch another node that talks to the listener:

```
> ./dht -l 10001 -d /ip4/127.0.0.1/tcp/10000/p2p/12D3KooWAP8mog99pPrF3Sq3WEnzsM2UoucKgfCGiQ2en4wK9SPD
```

The new node will insert the value `"Hello, world!"` into the DHT then retrieve it.

## Details

Notice that all we need to do is create a basic `NodeConfig` and then `OverlayNode`. No additional configuration is needed to
use the DHT.
