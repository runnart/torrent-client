package torrentfile

import (
	"net/url"
	"strconv"
)

// TorrentFile describes torrent file content.
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

// buildTrackerURL build tracker url using peerID and port number.
func (t *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	u, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	qParams := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	u.RawQuery = qParams.Encode()
	return u.String(), nil
}
