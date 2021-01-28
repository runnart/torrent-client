package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"

	"github.com/jackpal/bencode-go"
	tf "github.com/runnart/torrent-client/torrentfile"
)

type bencodeInfo struct {
	// Binary blob containing the SHA-1 hashes of each piece.
	Pieces string `bencode:"pieces"`
	// Exact size of a piece.
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

// hash calculate bencodeInfo SHA-1 sum
func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	if err := bencode.Marshal(&buf, *i); err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil

}

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // Length of SHA-1 hash in bytes
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		return nil, fmt.Errorf("received malformed pieces of length %d", len(buf))
	}

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, 0, numHashes)
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

// toTorrentFile convert's between bencodeTorrent and TorrentFile.
func (bto bencodeTorrent) toTorrentFile() (tf.TorrentFile, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return tf.TorrentFile{}, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return tf.TorrentFile{}, err
	}
	return tf.TorrentFile{
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		Announce:    bto.Announce,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}, nil
}

// OpenTorrentFile parses a torrent file.
func OpenTorrentFile(r io.Reader) (*bencodeTorrent, error) {
	bto := bencodeTorrent{}
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	return &bto, nil
}

func main() {
	f, err := os.Open("t.torrent")
	if err != nil {
		panic(err)
	}
	ff, err := OpenTorrentFile(f)
	if err != nil {
		panic(err)
	}
	torrent, err := ff.toTorrentFile()
	if err != nil {
		panic(err)
	}
	fmt.Println(torrent)
}
