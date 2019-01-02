package rtsp

import (
	"testing"
)

func TestDigestAuth(t *testing.T) {
	digest, err := NewDigest("/dir/index.html", `Digest realm="testrealm@host.com", qop="auth", nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093", opaque="5ccc069c403ebaf9f0171e9517f40e41"`)
	if err != nil {
		t.Error(err)
	}
	authfunc := digest.Authenticate("Mufasa", "Circle Of Life")
	digest.cnonce = "6443776e86b984dd"
	auth := authfunc("OPTIONS", nil)
	if auth != `Digest username="Mufasa", realm="testrealm@host.com", nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093", uri="/dir/index.html", algorithm="md5", opaque="5ccc069c403ebaf9f0171e9517f40e41", qop="auth", nc=00000001, cnonce="6443776e86b984dd", response="15f8e0d8b404b53a52e8cb7fa89988ee"` {
		t.Error(auth)
	}
}

func TestBasicAuth(t *testing.T) {
	basic, err := NewDigest("/dir/index.html", `Basic realm="testrealm@host.com"`)
	if err != nil {
		t.Error(err)
	}
	authfunc := basic.Authenticate("Mufasa", "Circle Of Life")
	auth := authfunc("OPTIONS", nil)
	if auth != "Basic TXVmYXNhOkNpcmNsZSBPZiBMaWZl" {
		t.Error(auth)
	}
}
