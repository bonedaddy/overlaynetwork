package overlaynetwork

import (
	"context"
	"crypto/sha256"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-routing"
	"github.com/multiformats/go-multihash"
	"sync"
	"time"
)

// Pubsub is a wrapper around the libp2p
type Pubsub struct {
	ps *pubsub.PubSub
	rt routing.IpfsRouting
	ht host.Host
}

// Publish will publish the provided data to the peers subscribed to the topic
func (p *Pubsub) Publish(ctx context.Context, topic string, data []byte) error {
	return p.ps.Publish(topic, data)
}

// Subscribe will subscribe you to  the given topic
func (p *Pubsub) Subscribe(ctx context.Context, topic string) (*pubsub.Subscription, error) {
	sub, err := p.ps.Subscribe(topic)
	if err != nil {
		return nil, err
	}

	go func() {
		h := sha256.Sum256([]byte("gossipsub:" + topic))
		encoded, err := multihash.Encode(h[:], multihash.SHA2_256)
		if err != nil {
			return
		}
		mh, err := multihash.Cast(encoded)
		if err != nil {
			return
		}
		id := cid.NewCidV1(cid.Raw, mh)
		go p.rt.Provide(ctx, id, true)
		p.connectToPubSubPeers(ctx, id)
	}()
	return sub, nil
}

// GetTopics returns the list of topics were currently subscribed to
func (p *Pubsub) GetTopics() []string {
	return p.ps.GetTopics()
}

// ListPeers returns the list of peers subscribed to a given topic
func (p *Pubsub) ListPeers(topic string) []peer.ID {
	return p.ps.ListPeers(topic)
}

func (p *Pubsub) connectToPubSubPeers(ctx context.Context, cid *cid.Cid) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	provs := p.rt.FindProvidersAsync(ctx, cid, 10)
	var wg sync.WaitGroup
	for prov := range provs {
		wg.Add(1)
		go func(pi peerstore.PeerInfo) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			err := p.ht.Connect(ctx, pi)
			if err != nil {
				log.Info("pubsub discover: ", err)
				return
			}
			log.Info("connected to pubsub peer:", pi.ID)
		}(prov)
	}
	wg.Wait()
}
