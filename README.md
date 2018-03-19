This project is an attempt to build a comprehensive library to connect to IP cameras and consume video and audio streams.
The ultimate goal is to create a robust streaming proxy converting IP camera stream into something directly usable from a browser.

At this time the following functionality has been implemented:
- Connecting to camera over RTSP.
- OPTIONS, DESCRIBE, SETUP, PLAY, PAUSE, and TEARDOWN commands.
- Handling Basic and Digest authentication.
- Simplistic parsing of SDP data.
- Parsing and building Transport header.
- Handling RTSP state machine with CSeq.
- Receiving RTP and RTCP packets over TCP.
- Unwrapping RTP/RTCP packets from RTSP message.
- Parsing RTP packets for h.264 NAL units.
- Handling NAL aggrgates, fragments, DONs (Decoding Order Number) and timestamps.
- Receiving and parsing basic RTCP packets.
- Initial work on UDP listeners for RTP over UDP.

I wanted to get real response data before I start writing tests for packets and messages.  That is in the plans.
It is hard to test RTSP as it operates as a state machine: RTSP messages (text protocol) and RTP packets (binary protocol) are coming through the same connection in an unpredictable order. 

I am building just client functionality.  I plan to finish RTP/RTCP over UDP, add sending keep-alive RTCP packets, implement RTSP over HTTP eventually.
Then the plans include transforming media streams to output HLS or MPEG-DASH.

There are many more intricacies involved in handling protocols proper.  For instance, packetisation mode from SDP can help with hadling RTP payload.  Sequence number returned in response to PLAY command can be used to find initial RTP packet to start streaming with, etc.  A better SDP parser would be useful.  Sorting out of order NAL units in MTAP NAL could be useful if there are cameras sending MTAPs.