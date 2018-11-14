package h264

import (
	"testing"
)

func TestAvailable(t *testing.T) {
	data := []byte{0x4D, 0x32}
	r := NewBitReader(data)
	n := r.Available()
	if n != 16 {
		t.Errorf("%d bits available, expected 16.", n)
	}
}

func TestRead(t *testing.T) {
	data := []byte{0x4D, 0x32} // 0100 1101 0011 0010
	r := NewBitReader(data)
	v, err := r.Read(3)
	if err != nil {
		t.Error(err)
	}
	if v != 0x2 {
		t.Errorf("Read value %x, expected 0x2", v)
	}
	v, err = r.Read(2)
	if v != 0x1 {
		t.Errorf("Read value %x, expected 0x1", v)
	}
	v, err = r.Read(7)
	if v != 0x53 {
		t.Errorf("Read value %x, expected 0x53", v)
	}
	n := r.Available()
	if n != 4 {
		t.Errorf("%d bits available, expected 4", n)
	}
	v, err = r.Read(2)
	if v != 0x0 {
		t.Errorf("Read value %x, expected 0x0", v)
	}
	v, err = r.Read(1)
	if v != 0x1 {
		t.Errorf("Read value %x, expected 0x1", v)
	}
	v, err = r.Read(1)
	if v != 0x0 {
		t.Errorf("Read value %x, expected 0x0", v)
	}
	n = r.Available()
	if n != 0 {
		t.Errorf("%d bits available, expected 0", n)
	}
	v, err = r.Read(1)
	if err == nil {
		t.Error("Expected error reading past the end of the buffer, got nothing")
	}
}

func TestSignedGolomb(t *testing.T) {
	data := []byte{
		0x80, // 1
		0x40, // 010
		0x60, // 011
		0x20, // 00100
		0x38, // 00111
	}
	expected := []int32{0, 1, -1, 2, -3}
	for i := range data {
		r := NewBitReader(data[i : i+1])
		v, err := r.ReadSignedGolomb()
		if err != nil {
			t.Error(err)
		}
		if v != expected[i] {
			t.Errorf("Read value %x, expected %x", v, expected[i])
		}
	}
}
