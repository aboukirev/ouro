package rtsp

import (
	"bufio"
	"encoding/binary"
	"net"
	"net/url"
	"time"
)

// TODO: Figure out sending out RTCP keep-alive packets.

type (
	// Conn encapsulates low level network connection and hides buffering and packetization.
	Conn struct {
		conn    net.Conn
		rdr     *bufio.Reader
		Timeout time.Duration
		URL     *url.URL // Parsed out original URI with user credentials.
		BaseURI string   // Formatted URI without user credentials.
	}

	// UDPConn maintains UDP connection for RTP and RTCP channel pair.
	UDPConn struct {
		data *net.UDPConn // UDP connection for data channel.
		ctrl *net.UDPConn // UDP connection for control channel.
		ch   byte         // Data channel number.  Control channel is ch + 1.
	}
)

var be = binary.BigEndian

// Dial opens RTSP connection and starts a session.
func Dial(uri string) (*Conn, error) {
	url1, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if _, _, err = net.SplitHostPort(url1.Host); err != nil {
		url1.Host = url1.Host + ":554"
	}

	conn, err := net.Dial("tcp", url1.Host)
	if err != nil {
		return nil, err
	}

	url2 := *url1
	url2.User = nil

	return &Conn{
		conn:    conn,
		rdr:     bufio.NewReaderSize(conn, 2048),
		Timeout: time.Millisecond * 2000, //time.Second * 2,
		URL:     url1,
		BaseURI: url2.String(),
	}, nil
}

func (c *Conn) Write(p []byte) (n int, err error) {
	if c.Timeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.Timeout))
	}
	return c.conn.Write(p)
}

// Peek returns the next n bytes without advancing the reader.
func (c *Conn) Peek(n int) ([]byte, error) {
	if c.Timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.Timeout))
	}
	return c.rdr.Peek(n)
}

// ReadLine reads until the first occurrence of newline,
// returning a string containing the data up to and including newline.
func (c *Conn) ReadLine() (string, error) {
	if c.Timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.Timeout))
	}
	return c.rdr.ReadString('\n')
}

// ReadUint16 reads and returns an unsigned 2-byte integer.
func (c *Conn) ReadUint16() (data uint16, err error) {
	if c.Timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.Timeout))
	}
	err = binary.Read(c.rdr, be, &data)
	return
}

// Discard skips the next n bytes, returning the number of bytes discarded.
func (c *Conn) Discard(n int) (discarded int, err error) {
	return c.rdr.Discard(n)
}

// ReadByte reads and returns a single byte.
func (c *Conn) ReadByte() (byte, error) {
	if c.Timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.Timeout))
	}
	return c.rdr.ReadByte()
}

// ReadBytes reads requested number of bytes, returning a slice containing the data.
func (c *Conn) ReadBytes(n int) ([]byte, error) {
	b, err := c.Peek(n)
	if err == nil {
		c.rdr.Discard(n)
	}
	return b, err
}

// NewUDPConn starts listening on a pair of UDP sockets for RTP and RTCP packets.
func (c *UDPConn) NewUDPConn(ch byte, port int) (err error) {
	c.ch = ch
	var addr *net.IPAddr
	if addr, err = net.ResolveIPAddr("ip", "127.0.0.1"); err != nil {
		return err
	}
	if c.data, err = net.ListenUDP("udp", &net.UDPAddr{IP: addr.IP, Port: port, Zone: ""}); err != nil {
		return err
	}
	if c.ctrl, err = net.ListenUDP("udp", &net.UDPAddr{IP: addr.IP, Port: port + 1, Zone: ""}); err != nil {
		c.data.Close()
		return err
	}
	return
}

// Process incoming UDP data sending packets to respecive channels.
func (c *UDPConn) Process(datach, ctrlch chan ChannelData) {
	go func() {
		var buf [2048]byte
		for {
			// TODO: Check for flag to close connection.
			n, _, err := c.data.ReadFromUDP(buf[:])
			if err != nil {
				break
			}
			if datach != nil {
				datach <- ChannelData{Channel: c.ch, Payload: append([]byte{}, buf[:n]...)}
			}
		}
		c.data.Close()
		// TODO: Notify caller about closed connection.
	}()
	go func() {
		var buf [2048]byte
		for {
			// TODO: Check for flag to close connection.
			n, _, err := c.ctrl.ReadFromUDP(buf[:])
			if err != nil {
				break
			}
			if ctrlch != nil {
				ctrlch <- ChannelData{Channel: c.ch + 1, Payload: append([]byte{}, buf[:n]...)}
			}
		}
		c.ctrl.Close()
		// TODO: Notify caller about closed connection.
	}()
}
