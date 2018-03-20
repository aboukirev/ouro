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
		addr    *net.IPAddr
		sinks   []udpsink
	}

	// Maintains UDP connection for RTP and RTCP channel pair.
	udpsink struct {
		*net.UDPConn
		done  chan int
		index byte
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

// AddSink creates a listener on UDP port for RTP data or control channel.
func (c *Conn) AddSink(index byte, port int) error {
	if c.addr == nil {
		var err error
		if c.addr, err = net.ResolveIPAddr("ip", "127.0.0.1"); err != nil {
			return err
		}
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: c.addr.IP, Port: port, Zone: ""})
	if err != nil {
		return err
	}
	sink := udpsink{
		UDPConn: conn,
		done:    make(chan int),
		index:   index,
	}
	c.sinks = append(c.sinks, sink)
	return nil
}

// Start initiates processing of incoming packets on all UDP listeners created thus far.
func (c *Conn) Start(datach, ctrlch chan ChannelData) {
	var ch chan ChannelData
	for i, s := range c.sinks {
		if (i % 2) == 0 {
			ch = datach
		} else {
			ch = ctrlch
		}
		go func(sink udpsink) {
			var buf [2048]byte
			for {
				select {
				case <-sink.done:
					break
				}
				n, _, err := sink.ReadFromUDP(buf[:])
				if err != nil {
					break
				}
				if ch != nil {
					ch <- ChannelData{Channel: sink.index, Payload: append([]byte{}, buf[:n]...)}
				}
			}
			sink.Close()
			// TODO: Notify caller about closed connection.
		}(s)
	}
}

// Stop tells all UDP listeners to suspend processing and close.  The list of listerners is reset.
func (c *Conn) Stop() {
	for _, sink := range c.sinks {
		select {
		case sink.done <- 1:
		}
	}
	c.sinks = c.sinks[:0]
}
