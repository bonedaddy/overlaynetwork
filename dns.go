package overlaynetwork

import (
	"fmt"
	"github.com/gcash/bchd/chaincfg"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"sync"
)

type LookupTXTFunc func(name string) (txt []string, err error)

// SeedFromDNS uses DNS seeding to populate the address manager with peers using the DNS TXT record.
func SeedFromDNS(chainParams *chaincfg.Params, lookupFn LookupTXTFunc) <-chan peerstore.PeerInfo {
	ch := make(chan peerstore.PeerInfo)
	go func() {
		var wg sync.WaitGroup
		// TODO: we should modify the DNSSeed object to add a bool for whether or not it supports
		// resolving overlay network addresses.
		for _, dnsseed := range chainParams.DNSSeeds {
			wg.Add(1)
			go func(host string) {
				txt, err := lookupFn(host)
				if err != nil {
					log.Infof("DNS discovery failed on seed %s: %v", host, err)
					return
				}
				for _, t := range txt {
					pi, err := ParseBootstrapPeer(t)
					if err != nil {
						return
					}
					ch <- pi
				}
				wg.Done()
			}(dnsseed.Host)
		}
		wg.Wait()
		close(ch)
	}()
	return ch
}

// ParseBootstrapPeer parses a DNS TXT record into a PeerInfo object
func ParseBootstrapPeer(addr string) (peerstore.PeerInfo, error) {
	p2pAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return peerstore.PeerInfo{}, err
	}

	pid, err := p2pAddr.ValueForProtocol(ma.P_P2P)
	if err != nil {
		return peerstore.PeerInfo{}, err
	}

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		return peerstore.PeerInfo{}, err
	}

	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/p2p/%s", peer.IDB58Encode(peerid)))
	targetAddr := p2pAddr.Decapsulate(targetPeerAddr)

	return peerstore.PeerInfo{
		Addrs: []ma.Multiaddr{
			targetAddr,
		},
		ID: peerid,
	}, nil
}
