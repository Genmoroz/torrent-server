package downloader

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

const pstr = "BitTorrent protocol"

type (
	Peers struct {
		peerIPs   []string
		mux       sync.Mutex
		peersChan chan *Peer
	}

	Peer struct {
		conn   net.Conn
		ip     string
		choked bool
	}

	handshakeMessage struct {
		pstr     string
		infoHash [20]byte
		peerID   [20]byte
	}
)

func ConnectToPeer(network, ip string, port uint16, infoHash, peerID [20]byte) (*Peer, error) {
	address := fmt.Sprintf("%s:%d", ip, port)
	log.Printf("resolving TCP, network: %s, address: %s", network, address)
	addr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("dialing TCP, network: %s, address: %s", network, address)
	conn, err := net.DialTCP(network, nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial with timeout: %w", err)
	}

	log.Printf("handshaking with Peer, network: %s, address: %s", network, address)
	if err = doHandshake(conn, infoHash, peerID); err != nil {
		return nil, fmt.Errorf("failed to do handshake: %w", err)
	}

	return &Peer{
		conn: conn,
		ip:   ip,
	}, nil
}

func doHandshake(conn *net.TCPConn, infoHash, peerID [20]byte) error {
	expected := handshakeMessage{
		pstr:     pstr,
		infoHash: infoHash,
		peerID:   peerID,
	}

	if err := writeHandshakeMessage(conn, expected); err != nil {
		return fmt.Errorf("failed to write handshake message: %w", err)
	}

	actual, err := readHandshakeMessage(conn)
	if err != nil {
		return fmt.Errorf("failed to read handshake message: %w", err)
	}

	if !bytes.Equal(actual.infoHash[:], infoHash[:]) {
		return errors.New("infoHash's are not equal")
	}

	return nil
}

func writeHandshakeMessage(conn *net.TCPConn, msg handshakeMessage) error {
	_, err := conn.Write(prepareHandshakeMessage(msg))
	return err
}

func readHandshakeMessage(conn *net.TCPConn) (*handshakeMessage, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrLen := int(lengthBuf[0])

	if pstrLen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return nil, err
	}

	handshakeBuf := make([]byte, pstrLen+48)
	_, err = io.ReadFull(conn, handshakeBuf)
	if err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuf[pstrLen+8:pstrLen+28])
	copy(peerID[:], handshakeBuf[pstrLen+28:])

	return &handshakeMessage{
		pstr:     string(handshakeBuf[0:pstrLen]),
		infoHash: infoHash,
		peerID:   peerID,
	}, nil
}
func prepareHandshakeMessage(msg handshakeMessage) []byte {
	buf := make([]byte, len(msg.pstr)+49)
	buf[0] = byte(len(msg.pstr))
	offset := 1
	offset += copy(buf[offset:], msg.pstr)
	offset += copy(buf[offset:], make([]byte, 8)) // 8 reserved bytes
	offset += copy(buf[offset:], msg.infoHash[:])
	offset += copy(buf[offset:], msg.peerID[:])
	return buf
}

func (p *Peers) addPeer(peerIP string, peer *Peer) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, k := range p.peerIPs {
		if k == peerIP {
			return fmt.Errorf("peer is already exist with peerIP: %s", peerIP)
		}
	}

	p.peersChan <- peer
	p.peerIPs = append(p.peerIPs, peerIP)
	return nil
}

func (p *Peers) removePeerIP(peerIP string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	for index, ip := range p.peerIPs {
		if ip == peerIP {
			p.peerIPs = append(p.peerIPs[:index], p.peerIPs[index+1:]...)
			return nil
		}
	}

	return fmt.Errorf("peerIP is not presented in peerIPs")
}

func (p *Peers) existPeerIP(peerIP string) bool {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, ip := range p.peerIPs {
		if ip == peerIP {
			return true
		}
	}

	return false
}

func (p *Peer) SendRequest(index, begin, length int64) error {
	req := FormatRequest(index, begin, length)
	_, err := p.conn.Write(req.Serialize())
	return err
}

type messageID uint8

const (
	// MsgChoke chokes the receiver
	MsgChoke messageID = 0
	// MsgUnchoke unchokes the receiver
	MsgUnchoke messageID = 1
	// MsgInterested expresses interest in receiving data
	MsgInterested messageID = 2
	// MsgNotInterested expresses disinterest in receiving data
	MsgNotInterested messageID = 3
	// MsgHave alerts the receiver that the sender has downloaded a piece
	MsgHave messageID = 4
	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgBitfield messageID = 5
	// MsgRequest requests a block of data from the receiver
	MsgRequest messageID = 6
	// MsgPiece delivers a block of data to fulfill a request
	MsgPiece messageID = 7
	// MsgCancel cancels a request
	MsgCancel messageID = 8
)

// Message stores ID and payload of a message
type Message struct {
	ID      messageID
	Payload []byte
}

// FormatRequest creates a REQUEST message
func FormatRequest(index, begin, length int64) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint64(payload[0:4], uint64(index))
	binary.BigEndian.PutUint64(payload[4:8], uint64(begin))
	binary.BigEndian.PutUint64(payload[8:12], uint64(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
// Interprets `nil` as a keep-alive message
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	length := uint32(len(m.Payload) + 1) // +1 for id
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}
