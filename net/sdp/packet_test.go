package sdp

import (
	"testing"
)

func TestParse(t *testing.T) {
	feeds := Parse([]byte(`
v=0
o=- 1 1 IN IP4
s=hysxrtpsion
c=IN IP4 0.0.0.0
t=0 0
a=control:*
a=range:npt=0-
m=video 0 RTP/AVP 96
a=rtpmap:96 H264/90000
a=control:trackID=0
a=fmtp:96 packetization-mode=1; sprop-parameter-sets=Z2QAKKzoBQBbkA==,aO48sA==; profile-level-id=640028
m=audio 0 RTP/AVP 97
b=AS:16
a=control:trackID=1
a=rtpmap:97 MPEG4-GENERIC/16000/1
a=fmtp:97 profile-level-id=1;mode=AAC-hbr;sizelength=13;indexlength=3;indexdeltalength=3;config=1408
m=application 0 RTP/AVP 111
a=control:trackID=2
a=rtpmap:111 X-KATA/1000
a=fmtp:111 octet-align=1
b=AS:2
`))
	t.Logf("%v", feeds)
}
