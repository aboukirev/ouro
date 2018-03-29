package rtsp

// Methods to facilitate RTSP over HTTP.

import (
	"bytes"
	"strconv"
	"strings"
)

// ConnectHTTP issues command for RTSP over HTTP wrapper.
func ConnectHTTP(verb, uri, cookie string) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString(verb)
	buf.WriteByte(' ')
	buf.WriteString(uri)
	buf.WriteString(" HTTP/1.0")
	buf.Write(crnl)
	buf.WriteString(HeaderXSessionCookie)
	buf.Write(colsp)
	buf.WriteString(cookie)
	buf.Write(crnl)
	buf.WriteString(HeaderAccept)
	buf.Write(colsp)
	buf.WriteString("application/x-rtsp-tunnelled")
	buf.Write(crnl)
	buf.WriteString(HeaderPragma)
	buf.Write(colsp)
	buf.WriteString("no-cache")
	buf.Write(crnl)
	buf.WriteString(HeaderCacheControl)
	buf.Write(colsp)
	buf.WriteString("no-store")
	buf.Write(crnl)
	if verb == "POST" {
		buf.WriteString(HeaderContentLength)
		buf.Write(colsp)
		buf.WriteString("32767") // Arbitrarily large number according to specification
		buf.Write(crnl)
	}
	buf.WriteString(HeaderUserAgent)
	buf.Write(colsp)
	buf.WriteString(Agent)
	buf.Write(crnl)
	buf.Write(crnl)
	return buf.Bytes()
}

// ReceiveHTTP reads headers form the response to GET.
func ReceiveHTTP(rdr *Conn) error {
	line, err := rdr.ReadLine()
	if err != nil {
		return err
	}
	i := strings.IndexByte(line, ' ')
	if i == -1 {
		return errMalformedResponse
	}
	if line[:i] != "HTTP/1.0" {
		return errNotSupported
	}
	status := strings.TrimSpace(line[i+1:])
	if i := strings.IndexByte(status, ' '); i != -1 {
		status = status[:i]
	}
	if len(status) != 3 {
		return errInvalidStatus
	}
	statusCode, err := strconv.Atoi(status)
	if err != nil || statusCode < 0 {
		return errInvalidStatus
	}
	// TODO: Verify expected 200 status.

	// Parse the response headers.
	for {
		line, err = rdr.ReadLine()
		if err != nil {
			return err
		}
		if line = strings.TrimSpace(line); line == "" {
			break
		}
		keyval := strings.SplitN(line, ":", 2)
		if len(keyval) != 2 {
			return errMalformedResponse
		}
		// TODO: Verify some of the response headers.
	}

	return nil
}
