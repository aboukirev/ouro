package rtsp

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"
)

// TODO: Figure out sending out RTCP keep-alive packets.

type (
	// Conn encapsulates low level network connection and hides buffering and packetization.
	Conn struct {
		conn      *net.TCPConn  // Main in/out connection in regular RTSP, GET connection in RTSP over HTTP
		post      *net.TCPConn  // POST connection in RTSP over HTTP
		rdr       *bufio.Reader // For line-oriented text protocol
		connected bool          // Indicates that POST has been issued for RTSP over HTTP connection
		guid      string
		Proto     int
		Timeout   time.Duration
		URL       *url.URL    // Parsed out original URI with user credentials.
		BaseURI   string      // Formatted URI without user credentials.
		addr      *net.IPAddr // Local address for UDP listeners
		sinks     []udpsink   // UDP listeners, 2 per media stream: data and control
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

	addr, err := net.ResolveTCPAddr("tcp", url1.Host)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	url2 := *url1
	url2.User = nil

	c := &Conn{
		conn:      conn,
		rdr:       bufio.NewReaderSize(conn, 2048),
		connected: false,
		guid:      fmt.Sprintf("%016x", rand.Uint64()),
		Proto:     ProtoTCP,
		Timeout:   time.Millisecond * 2000, //time.Second * 2,
		URL:       url1,
		BaseURI:   url2.String(),
	}

	if c.Proto == ProtoHTTP {
		// Issue GET and read response.  Afterwards we'll be getting incoming RTSP/RTP/RTCP responses in a stream.
		cmd := ConnectHTTP("GET", c.BaseURI, c.guid)
		if _, err := c.conn.Write(cmd); err != nil {
			return nil, err
		}
		// Read response but only headers and prepare to receive content stream of unknown length.
		if err := ReceiveHTTP(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Conn) Write(p []byte) (n int, err error) {
	if c.Proto == ProtoHTTP {
		if c.Timeout > 0 {
			c.post.SetWriteDeadline(time.Now().Add(c.Timeout))
		}
		// Issue POST for RTSP over HTTP.  Before first command or after temporary disconnect (after sending PLAY, for instance).
		if !c.connected {
			cmd := ConnectHTTP("POST", c.BaseURI, c.guid)
			if _, err := c.post.Write(cmd); err != nil {
				return 0, err
			}
			c.connected = true
		}
		w := base64.NewEncoder(base64.StdEncoding, c.post)
		defer w.Close()
		return w.Write(p)
	}
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
