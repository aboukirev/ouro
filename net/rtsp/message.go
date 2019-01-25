package rtsp

import (
	"bytes"
	"net/textproto"
	"strconv"
	"strings"
)

type (
	// An MessageHeader represents a MIME-style header mapping  keys to values.
	MessageHeader map[string]string

	// Headers facilitates passing additional headers to the RTSP command formatter.
	Headers map[string]string

	// Request encapsulates verb, uri, and headers for RTSP command.
	Request struct {
		Verb    string
		Cseq    int
		URI     string
		Auth    string
		Session string
		Header  MessageHeader
	}

	// Response encapsulates RTSP response.  It is modeled by http.Response but includes only what is needed to handle RTSP.
	Response struct {
		Proto         string
		Status        string
		StatusCode    int
		ContentLength int64
		Cseq          int
		Header        MessageHeader
		Body          []byte
	}

	// Queue tracks outgoing requests that have not yet been responded to.
	Queue map[int]*Request
)

var (
	crnl  = []byte{'\r', '\n'}
	colsp = []byte{':', ' '}
)

// Set sets the header associated with key to the given value.
func (h MessageHeader) Set(key string, value string) {
	h[textproto.CanonicalMIMEHeaderKey(key)] = value
}

// Get gets the first value associated with the given key.
func (h MessageHeader) Get(key string) string {
	if h == nil {
		return ""
	}
	v, ok := h[textproto.CanonicalMIMEHeaderKey(key)]
	if ok {
		return v
	}
	return ""
}

// Del deletes the values associated with key.
func (h MessageHeader) Del(key string) {
	delete(h, textproto.CanonicalMIMEHeaderKey(key))
}

// Pack request into RTSP message.
func (r *Request) Pack() []byte {
	buf := &bytes.Buffer{}
	buf.WriteString(r.Verb)
	buf.WriteByte(' ')
	buf.WriteString(r.URI)
	buf.WriteString(" RTSP/1.0")
	buf.Write(crnl)
	buf.WriteString(HeaderCSeq)
	buf.Write(colsp)
	buf.WriteString(strconv.Itoa(r.Cseq))
	buf.Write(crnl)
	for key, val := range r.Header {
		buf.WriteString(key)
		buf.Write(colsp)
		buf.WriteString(val)
		buf.Write(crnl)
	}
	if r.Session != "" && r.Verb != VerbOptions && r.Verb != VerbSetup {
		buf.WriteString(HeaderSession)
		buf.Write(colsp)
		buf.WriteString(r.Session)
		buf.Write(crnl)
	}
	if r.Auth != "" {
		buf.WriteString(HeaderAuthorization)
		buf.Write(colsp)
		buf.WriteString(r.Auth)
		buf.Write(crnl)
	}
	buf.WriteString(HeaderUserAgent)
	buf.Write(colsp)
	buf.WriteString(Agent)
	buf.Write(crnl)
	buf.WriteString(HeaderContentLength)
	buf.Write(colsp)
	buf.WriteByte('0') // We do not have any content in Client -> Server commands.
	buf.Write(crnl)
	buf.Write(crnl)
	return buf.Bytes()
}

// Unpack RTSP message into response structure.
func Unpack(rdr *Conn) (*Response, error) {
	line, err := rdr.ReadLine()
	if err != nil {
		return nil, err
	}
	i := strings.IndexByte(line, ' ')
	if i == -1 {
		return nil, errMalformedResponse
	}
	if line[:i] != "RTSP/1.0" {
		return nil, errNotSupported
	}
	r := &Response{
		Header: make(MessageHeader),
	}
	r.Proto = line[:i]
	r.Status = strings.TrimSpace(line[i+1:])
	statusCode := r.Status
	if i := strings.IndexByte(r.Status, ' '); i != -1 {
		statusCode = r.Status[:i]
	}
	if len(statusCode) != 3 {
		return nil, errInvalidStatus
	}
	r.StatusCode, err = strconv.Atoi(statusCode)
	if err != nil || r.StatusCode < 0 {
		return nil, errInvalidStatus
	}

	// Parse the response headers.
	for {
		line, err = rdr.ReadLine()
		if err != nil {
			return nil, err
		}
		if line = strings.TrimSpace(line); line == "" {
			break
		}
		keyval := strings.SplitN(line, ":", 2)
		if len(keyval) != 2 {
			return nil, errMalformedResponse
		}
		r.Header.Set(strings.TrimSpace(keyval[0]), strings.TrimSpace(keyval[1]))
	}

	r.Cseq, _ = strconv.Atoi(r.Header.Get(HeaderCSeq))

	r.ContentLength, err = strconv.ParseInt(r.Header.Get(HeaderContentLength), 10, 64)
	if err != nil {
		err = nil
		r.ContentLength = 0
	}

	if r.ContentLength > 0 {
		r.Body, err = rdr.ReadBytes(int(r.ContentLength))
	}

	return r, err
}

// Pack response into RTSP message.
func (r Response) Pack() []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("RTSP/1.0")
	buf.WriteByte(' ')
	buf.WriteString(r.Status)
	buf.Write(crnl)
	for key, val := range r.Header {
		buf.WriteString(key)
		buf.Write(colsp)
		buf.WriteString(val)
		buf.Write(crnl)
	}
	buf.Write(crnl)
	if r.ContentLength > 0 {
		buf.Write(r.Body)
		buf.Write(crnl)
	}
	return buf.Bytes()
}
