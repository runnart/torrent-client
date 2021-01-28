package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/runnart/torrent-client/bitfield"
	"github.com/runnart/torrent-client/handshake"
	"github.com/runnart/torrent-client/p2p"
)

// Client is a TCP connection with a peer.
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     p2p.Peer

	infoHash [20]byte
	peerID   [20]byte
}

func completeHandshake(conn net.Conn, infoHash, peerID [20]byte) (*handshake.Handshake, error) {
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.SetDeadline(time.Time{})
	}()

	req := NewHandshake(infoHash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}
	res, err := ReadHandshake(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("expected infohash %x got %x", res.InfoHash, infoHash)
	}
	return res, nil
}

func recvBitfield(conn net.Conn) (Bitfield, error) {
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.SetDeadline(time.Time{}) // disable deadline
	}()

	msg, err := ReadMessage(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, fmt.Errorf("expected bitfield but got %s", msg)
	}
	if msg.ID != MsgBitfield {
		return nil, fmt.Errorf("expected bitfield but got ID %d", msg.ID)
	}

	return msg.Payload, nil
}

// NewClient connects with a peer, completes a handshake, and receives a handshake.
func NewClient(peer Peer, peerID, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	if _, err := completeHandshake(conn, infoHash, peerID); err != nil {
		return nil, err
	}
	bf, err := recvBitfield(conn)
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bf,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil
}
