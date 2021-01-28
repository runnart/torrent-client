package p2p

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

// Peer encodes connection info for a peer.
type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}

// Unmarshal parses peer IP addresses and ports from a buffer.
func Unmarshal(peersBin []byte) ([]Peer, error) {
	// 4 bytes reserved for IP
	// 2 bytes reserved for port number in big-endian
	// Example:
	//		peers: [192|0|2|123|0x1A|0xE1|...] - 192.0.2.123:6881, ...
	//			   [ 	 IP	   |   PORT  |...]
	const peerSize = 6
	numPeers := len(peersBin) / peerSize
	if len(peersBin)%peerSize != 0 {
		return nil, fmt.Errorf("received malformed peers")
	}
	peers := make([]Peer, 0, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers = append(peers, Peer{
			IP:   peersBin[offset : offset+4],
			Port: binary.BigEndian.Uint16(peersBin[offset+4 : offset+6]),
		})
	}
	return peers, nil
}
