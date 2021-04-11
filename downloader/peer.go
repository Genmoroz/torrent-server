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

type (
	Peers struct {
		peerIPs   []string
		mux       sync.Mutex
		peersChan chan *Peer
	}

	Peer struct {
		conn       net.Conn
		ip         string
		choked     bool
		downloaded uint32
		backlog    uint32
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

const pstr = "BitTorrent protocol"

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

func (p *Peer) ReadMessage() error {
	lengthBuf := make([]byte, 4)
	if err := readFull(p.conn, lengthBuf); err != nil {
		return err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// keep-alive message
	if length == 0 {
		return nil
	}

	messageBuf := make([]byte, length)
	if err := readFull(p.conn, messageBuf); err != nil {
		return err
	}

	msg := &Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}

	switch msg.ID {
	case MsgUnchoke:
		p.choked = false
		return nil
	case MsgChoke:
		p.choked = true
		return nil
	case MsgHave:
		index, err := parseHaveMessage(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case MsgPiece:
		n, err := message.ParsePiece(p.index, state.buf, msg)
		if err != nil {
			return err
		}
		p.downloaded += n
		p.backlog--
	default:
		return fmt.Errorf("handler for messageID [%d] does not exist", msg.ID)
	}

	return nil
}

func readFull(r io.Reader, buf []byte) error {
	_, err := io.ReadFull(r, buf)
	return err
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

func (p *Peer) SendRequest(index, begin, length uint32) error {
	req := FormatRequest(index, begin, length)
	_, err := p.conn.Write(req.Serialize())
	return err
}

func parseHaveMessage(msg *Message) (uint32, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("expected HAVE (ID %d), got ID %d", MsgHave, msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("expected payload length 4, got length %d", len(msg.Payload))
	}

	return binary.BigEndian.Uint32(msg.Payload), nil
}
