package rtsp

// TODO: Could I use * instead of URI in RTSP commands?

import (
	"log"
	"strings"
	"sync"
	"time"
)

var (
	keepAliveTimeout = time.Second * 2
)

type (
	// Session maintains RTSP session and workflow.
	Session struct {
		*Conn
		sync.Mutex
		stage   int        // Current stage.
		State   chan int   // Stage signaller.
		auth    DigestAuth // Callback function to calculate digest authentication for a given verb/method.
		queue   Queue      // RTSP requests that we are waiting responses for
		done    chan struct{}
		session string
		feeds   []*Feed
		feedidx int
		verbs   map[string]struct{}
		last    time.Time
		cseq    int
	}
)

// NewSession returns new RTSP session manager.
func NewSession() *Session {
	return &Session{
		State: make(chan int),
		stage: StageInit,
		queue: make(Queue),
		verbs: make(map[string]struct{}, 11),
	}
}

// Open to an RTSP source.
func (s *Session) Open(uri string, proto int) error {
	conn, err := Dial(uri, proto)
	if err != nil {
		return err
	}
	s.Conn = conn
	go s.process()
	return s.Options()
}

func (s *Session) authorize(challenge string) error {
	digest, err := NewDigest(s.BaseURI, challenge)
	if err != nil {
		return err
	}
	if s.URL.User == nil {
		return errNoCredentials
	}
	username := s.URL.User.Username()
	password, _ := s.URL.User.Password()
	s.auth = digest.Authenticate(username, password)
	return nil
}

func (s *Session) enqueue(req *Request) {
	s.Lock()
	defer s.Unlock()
	req.Cseq = s.cseq
	req.Session = s.session
	s.queue[req.Cseq] = req
}

func (s *Session) dequeue(seq int) (req *Request, ok bool) {
	s.Lock()
	defer s.Unlock()
	req, ok = s.queue[seq]
	if ok {
		delete(s.queue, seq)
	}
	return
}

func (s *Session) command(verb, uri string, headers Headers) error {
	if s.Conn == nil {
		return errNoConnection
	}
	req := &Request{Verb: verb, URI: uri, Header: make(MessageHeader)}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if s.auth != nil {
		req.Auth = s.auth(req.Verb, nil)
	}
	s.cseq++
	s.enqueue(req)
	buf := req.Pack()

	log.Println(string(buf))

	if _, err := s.Write(buf); err != nil {
		return err
	}
	return nil
}

// Options handles client OPTIONS request in RTSP.
func (s *Session) Options() error {
	return s.command(VerbOptions, s.BaseURI, nil)
}

// Describe handles DESCRIBE request in RTSP and parses SDP data in response.
func (s *Session) Describe() error {
	return s.command(VerbDescribe, s.BaseURI, Headers{HeaderAccept: "application/sdp"})
}

// Setup issues SETUP command for all playable media.
func (s *Session) Setup() error {
	for i := range s.feeds {
		// Setup is done on a different URI that accounts for Control in media.
		uri := s.feeds[i].Control
		if !strings.HasPrefix(uri, "rtsp://") {
			uri = s.BaseURI + "/" + uri
		}
		err := s.command(VerbSetup, uri, Headers{HeaderTransport: s.feeds[i].TransportHeader()})
		if err != nil {
			return err
		}
		s.feeds[i].cseq = s.cseq
	}
	return nil
}

// Play handles client PLAY request in RTSP.
func (s *Session) Play() error {
	return s.command(VerbPlay, s.BaseURI, nil)
}

// Pause handles client PAUSE request in RTSP.
func (s *Session) Pause() error {
	return s.command(VerbPause, s.BaseURI, nil)
}

// Teardown handles client TEARDOWN request in RTSP.
func (s *Session) Teardown() error {
	// s.notify(StageDone)
	return s.command(VerbTeardown, s.BaseURI, nil)
}

// KeepAlive executes OPTIONS on a regular basis to keep connection alive.
// RTP over TCP does not need keep-alive messages according to RFC.
func (s *Session) KeepAlive() error {
	if s.Proto != ProtoHTTP && s.Proto != ProtoTCP && s.stage > StageInit && s.stage < StageDone {
		if s.last.IsZero() || time.Now().Sub(s.last) >= keepAliveTimeout {
			s.last = time.Now()
			// OPTIONS without session might not keep session alive.  GET_PARAMETER may be unsupported by the server.  Check verbs.
			return s.command(VerbOptions, s.BaseURI, Headers{HeaderSession: s.session})
		}
	}
	return nil
}

func (s *Session) receive() error {
	var ch byte
	var buf []byte
	b, err := s.Peek(1)
	if err != nil {
		return err
	}
	if b[0] == '$' {
		s.Discard(1)
		// RTSP allows for up to 8 transports so valid ch values are limited.
		if ch, err = s.ReadByte(); err != nil {
			return err
		}
		length, err := s.ReadUint16()
		if err != nil {
			return err
		}
		buf, err = s.ReadBytes(int(length))
		if err == nil {
			select {
			case s.Data <- RawPacket{Channel: ch, Payload: append([]byte{}, buf...)}:
			}
		}
		return err
	}
	return s.handleRtsp()
}

func (s *Session) process() {
	for {
		select {
		case <-s.done:
			return
		default:
			if err := s.KeepAlive(); err != nil && s.stage == StageDone {
				return
			}
			if err := s.receive(); err != nil && !isTimeoutOrTemp(err) {
				log.Println(err)
				return
			}
		}
	}
}

// Notify about stage change.
func (s *Session) notify(stage int) {
	s.stage = stage
	select {
	case s.State <- stage:
	}
}

func (s *Session) handleRtsp() (err error) {
	var rsp *Response
	if rsp, err = Unpack(s.Conn); err != nil {
		return err
	}

	log.Println(string(rsp.Pack()))

	req, ok := s.dequeue(rsp.Cseq)
	if !ok {
		// FIXME: Did not find matching request.  Maybe just drop response.
		return errBadResponse
	}

	if rsp.StatusCode == RtspUnauthorized {
		if err = s.authorize(rsp.Header.Get(HeaderAuthenticate)); err != nil {
			return err
		}
		req.Auth = s.auth(req.Verb, nil)
		// Treat this as retransmit of the same message.
		// If the following line is uncommented, it will break matching responses to transport setup commands based on checking CSeq.
		// s.cseq++
		s.enqueue(req)
		buf := req.Pack()

		log.Println(string(buf))

		_, err = s.Write(buf)
		return err
	}

	if sess := rsp.Header.Get(HeaderSession); sess != "" {
		if fields := strings.Split(sess, ";"); len(fields) > 0 {
			s.session = fields[0]
		}
	}

	switch req.Verb {
	case VerbOptions:
		if s.stage == StageInit {
			for _, v := range strings.Split(rsp.Header.Get(HeaderPublic), ", ") {
				s.verbs[v] = struct{}{}
			}
			return s.Describe()
		}
	case VerbDescribe:
		if rsp.Body != nil {
			if s.feeds, err = ParseFeeds(s.Proto, rsp.Body); err != nil {
				return err
			}
			return s.Setup()
		}
	case VerbPause:
		if rsp.StatusCode == RtspOK {
			s.notify(StagePause)
		}
	case VerbPlay:
		if rsp.StatusCode == RtspOK {
			s.notify(StagePlay)
		}
	case VerbTeardown:
		if rsp.StatusCode == RtspOK {
			s.notify(StageDone)
		}
	case VerbSetup:
		return s.handleSetup(rsp)
	}

	return
}

func (s *Session) handleSetup(rsp *Response) (err error) {
	if rsp.StatusCode == RtspOK {
		for _, f := range s.feeds {
			if f.cseq == rsp.Cseq {
				s.feedidx++
				if err = f.TransportSetup(rsp.Header.Get(HeaderTransport)); err != nil {
					return err
				}
				if !f.transp.IsTCP {
					ch := byte(s.feedidx) * 2
					if err := s.AddSink(ch, f.transp.Port.One); err != nil {
						return err
					}
					if err := s.AddSink(ch+1, f.transp.Port.Two); err != nil {
						return err
					}
				}
			}
		}
	} else if rsp.StatusCode != RtspUnauthorized {
		// Stream is not available even though SDP told us it is.
		s.feedidx++
	}
	if s.feedidx >= len(s.feeds) {
		s.notify(StageReady)
		s.Start()
	}
	return
}
