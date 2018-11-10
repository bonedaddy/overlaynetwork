# Echo client/server example

This example shows how you can use the overlay network to make direct connections between two peers and
run a custom protocol (an echo protocol in this case). Connections will automatically be encrypted and 
authenticated. All you need to do is register your protocol handler and you're good to go.
## Build

From `go-libp2p` base folder:

```
> go build ./examples/echo
```

## Usage

```
> ./echo -l 10000
2017/03/15 14:11:32 listening for connections
2017/03/15 14:11:32 Now run "./echo -l 10001 -d /ip4/127.0.0.1/tcp/10000/p2p/12D3KooWAP8mog99pPrF3Sq3WEnzsM2UoucKgfCGiQ2en4wK9SPD" on a different terminal
```

The listener libp2p host will print its `Multiaddress`, which indicates how it can be reached (ip4+tcp) and its randomly generated ID (`12D3KooWA...`)

Now, launch another node that talks to the listener:

```
> ./echo -l 10001 -d /ip4/127.0.0.1/tcp/10000/p2p/12D3KooWAP8mog99pPrF3Sq3WEnzsM2UoucKgfCGiQ2en4wK9SPD
```

The new node with send the message `"Hello, world!"` to the listener, which will in turn echo it over the stream and close it. The listener logs the message, and the sender logs the response.

## Details

Notice that all we need to do is create a basic `NodeConfig` and then `OverlayNode`. To register our echo protocol we simply do
```go
node.Host.SetStreamHandler("/bitcoincash/echo/1.0.0", func(s net.Stream) {
	// code here
})
```

	