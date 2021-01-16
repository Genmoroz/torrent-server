package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"time"
	"torrent-server/client"
	"torrent-server/loader"
	"torrent-server/parser/bencode"
)

func main() {
	content, err := loader.ReadFile("./test.torrent")
	if err != nil {
		log.Fatal(err)
	}

	peerID := [20]byte{}
	_, err = rand.Read(peerID[:])
	if err != nil {
		log.Fatalln(err)
	}
	torrent, err := bencode.NewParser().Parse(content)
	url, err := client.PrepareTrackerURL(torrent, peerID, 6881)
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		pc, err := net.ListenPacket("udp", ":8080")
		if err != nil {
			log.Fatal(err)
		}
		defer pc.Close()

		for {
			buf := make([]byte, 1024)
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				continue
			}
			go serve(pc, addr, buf[:n])
		}
	}()

	conn, err := net.Dial("udp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}
	httpReq := fmt.Sprintf(
		`GET %s?%s HTTP/1.1`,
		url.Path,
		url.RawQuery,
	)
	n, err := fmt.Fprintf(conn, httpReq)
	//conn.Write()
	//n, err := conn.Write([]byte(httpReq))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("writen ", n, " bytes")

	status, err := bufio.NewReader(conn).ReadByte()
	if err != nil {
		log.Fatalln(err)
	}

	time.Sleep(99999 * time.Second)

	fmt.Println(status)
	//if err = client.Get(url); err != nil {
	//	log.Fatal(err)
	//}
}

func serve(pc net.PacketConn, addr net.Addr, buf []byte) {
	// 0 - 1: ID
	// 2: QR(1): Opcode(4)
	buf[2] |= 0x80 // Set QR bit

	pc.WriteTo(buf, addr)
}
