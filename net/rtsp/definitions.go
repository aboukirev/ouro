package rtsp

import (
	"errors"
)

const (
	// constant            method            direction        object     requirement

	// VerbDescribe        DESCRIBE          C->S             P,S        recommended
	VerbDescribe = "DESCRIBE"
	// VerbAnnounce        ANNOUNCE          C->S, S->C       P,S        optional
	VerbAnnounce = "ANNOUNCE"
	// VerbGetParameter    GET_PARAMETER     C->S, S->C       P,S        optional
	VerbGetParameter = "GET_PARAMETER"
	// VerbOptions         OPTIONS           C->S, S->C       P,S        required (S->C: optional)
	VerbOptions = "OPTIONS"
	// VerbPause           PAUSE             C->S             P,S        recommended
	VerbPause = "PAUSE"
	// VerbPlay            PLAY              C->S             P,S        required
	VerbPlay = "PLAY"
	// VerbRecord          RECORD            C->S             P,S        optional
	VerbRecord = "RECORD"
	// VerbRedirect        REDIRECT          S->C             P,S        optional
	VerbRedirect = "REDIRECT"
	// VerbSetup           SETUP             C->S             S          required
	VerbSetup = "SETUP"
	// VerbSetParameter    SET_PARAMETER     C->S, S->C       P,S        optional
	VerbSetParameter = "SET_PARAMETER"
	// VerbTeardown        TEARDOWN          C->S             P,S        required
	VerbTeardown = "TEARDOWN"
)

const (
	// ProtoTCP requests TCP as lower protocol.
	ProtoTCP = 0
	// ProtoUnicast requests UDP unicast as lower protocol.
	ProtoUnicast = 1
	// ProtoMulticast requests UDP multicast as lower protocol.
	ProtoMulticast = 2
	// ProtoHTTP opens HTTP connections with GET and POST, transmits and receives base64 encoded RTSP messages over it.
	ProtoHTTP = 3
)

const (
	// StageInit indicates initialization stage.
	StageInit = iota
	// StageReady indicates that setup of all transports is complete.
	StageReady
	// StagePlay indicates that playback is in progress.
	StagePlay
	// StagePause indicates that playback is paused.
	StagePause
	// StageDone indicates that connection is closing.
	StageDone
)

const (
	// HeaderAccept is HTTP Accept header.
	HeaderAccept = "Accept"
	// HeaderAuthenticate is HTTP WWW-Authenticate header.
	HeaderAuthenticate = "WWW-Authenticate"
	// HeaderAuthorization is HTTP Accept header.
	HeaderAuthorization = "Authorization"
	// HeaderContentLength is HTTP Content-Length header.
	HeaderContentLength = "Content-Length"
	// HeaderCSeq is RTSP CSeq header.
	HeaderCSeq = "CSeq"
	// HeaderPublic is HTTP Public header.
	HeaderPublic = "Public"
	// HeaderSession is RTSP Session header.
	HeaderSession = "Session"
	// HeaderTransport is RTSP Transport header.
	HeaderTransport = "Transport"
	// HeaderUserAgent is RTSP User-Agent header.
	HeaderUserAgent = "User-Agent"
	// HeaderXSessionCookie is HTTP X-SessionCookie header for RTSP over HTTP.
	HeaderXSessionCookie = "X-SessionCookie"
	// HeaderPragma is HTTP Pragma header.
	HeaderPragma = "Pragma"
	// HeaderCacheControl is HTTP Cache-Control header.
	HeaderCacheControl = "Cache-Control"
)

const (
	// RtspOK indicates that request has succeeded.
	RtspOK = 200
	// RtspNoContent indicates that server has successfully fulfilled the request and that there is no additional content to send in the response payload body.
	RtspNoContent = 204
	// RtspNotModified indicates that client, which made the request conditional, already has a valid representation.
	RtspNotModified = 304
	// RtspUnauthorized indicates that clients should authorize to be served complete response to current request.
	RtspUnauthorized = 401
	// RtspLowOnStorageSpace indicates insufficient storage space on server to satisfy record request.
	RtspLowOnStorageSpace = 250
	// RtspMethodNotAllowed indicates that method specified in the request is not allowed for the resource identified by the request URI.
	RtspMethodNotAllowed = 405
	// RtspParameterNotUnderstood indicates that recipient of the request does not support one or more parameters contained in the request.
	RtspParameterNotUnderstood = 451
	// RtspConferenceNotFound indicates that conference indicated by a Conference header field is unknown to the media server.
	RtspConferenceNotFound = 452
	// RtspNotEnoughBandwidth indicates that request was refused because there was insufficient bandwidth.
	RtspNotEnoughBandwidth = 453
	// RtspSessionNotFound indicates that RTSP session identifier in the Session header is missing,
	RtspSessionNotFound = 454
	// RtspMethodNotValidInThisState indicates that client or server cannot process this request in its current state.
	RtspMethodNotValidInThisState = 455
	// RtspHeaderFieldNotValidForResource indicates that server could not act on a required request header.
	RtspHeaderFieldNotValidForResource = 456
	// RtspInvalidRange indicates that Range value given is out of bounds.
	RtspInvalidRange = 457
	// RtspParameterIsReadOnly indicates that parameter to be set by SET_PARAMETER can be read but not modified.
	RtspParameterIsReadOnly = 458
	// RtspAggregateOperationNotAllowed indicates that requested method may not be applied on the URL in question since it is an aggregate (presentation) URL.
	RtspAggregateOperationNotAllowed = 459
	// RtspOnlyAggregateOperationAllowed indicates that requested method may not be applied on the URL in question since it is not an aggregate (presentation) URL.
	RtspOnlyAggregateOperationAllowed = 460
	// RtspUnsupportedTransport indicates that Transport field did not contain a supported transport specification.
	RtspUnsupportedTransport = 461
	// RtspDestinationUnreachable indicates that data transmission channel could not be established because the client address could not be reached.
	RtspDestinationUnreachable = 462
	// RtspOptionNotSupported indicates that option given in the Require or the Proxy-Require fields was not supported.
	RtspOptionNotSupported = 551
)

// Errors
var (
	errMalformedResponse = errors.New("Malformed response header")
	errInvalidStatus     = errors.New("Missing or malformed status code")
	errNotSupported      = errors.New("Unsupported protocol version")
	errNoCredentials     = errors.New("Resource URI has no credentials to perform authorization")
	errOutOfMemory       = errors.New("Out of memory")
	errBadResponse       = errors.New("Bad or unexpected response")
	errTimeout           = errors.New("Network timeout")
	errNoConnection      = errors.New("Connection to RTSP source is required")
	errPacketTooShort    = errors.New("Packet is too short")
	errInvalidParameter  = errors.New("Invalid parameter")
)

const (
	// Agent is user agent string for this application.
	// FIXME: Make this configurable.
	Agent = "Vigil/1.0"
)
