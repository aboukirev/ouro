package main

import (
	// "encoding/hex"
	"flag"
	"log"
	"os"
	"time"

	"github.com/daneshvar/ouro/net/h264"
	"github.com/daneshvar/ouro/net/rtcp"
	"github.com/daneshvar/ouro/net/rtp"
	"github.com/daneshvar/ouro/net/rtsp"
)

func handleStates(s *rtsp.Session) {
	state := rtsp.StageInit
	nalsink := h264.NewNALSink()
	tmr := time.NewTimer(time.Minute)
	tmr.Stop()
	for {
		select {
		case <- tmr.C:
			log.Printf("Timer: stage=%d\n", state)
			if state == rtsp.StagePlay {
				if err := s.Pause(); err != nil {
					log.Fatal(err)
				}
			} else if state == rtsp.StagePause {
				log.Println("Stage: Closing")
				if err := s.Teardown(); err != nil {
					log.Fatal(err)
				}
			} else if state == rtsp.StageDone {
				return
			}
		case state = <- s.State:
			switch state {
			case rtsp.StageReady:
				log.Println("Stage: Ready")
				if err := s.Play(); err != nil {
					log.Fatal(err)
				}
			case rtsp.StagePlay:
				log.Println("Stage: Playing")
				tmr.Reset(time.Second * 5)
			case rtsp.StagePause:
				log.Println("Stage: Pausing")
				tmr.Reset(time.Second * 5)
			case rtsp.StageDone:
				log.Println("Stage: Done")
				tmr.Reset(time.Second * 5)
			default:
				log.Printf("Stage: %d\n", state)
			}
		case pkt := <- s.Data:
            if pkt.Channel % 2 == 0 {
				var p *rtp.Packet
				var err error
				if p, err = rtp.Unpack(pkt.Payload); err != nil {
					log.Println(err)
					return
				}
			
				log.Printf("RTP [%d] PT=%d, CC=%d, M=%t, SN=%d\r\n", pkt.Channel, p.PT(), p.CC(), p.M(), p.SN)
			
				buf := p.PL
				if buf != nil {
					err := nalsink.Push(buf, p.TS)
					if err != nil {
						log.Println(err)
						// log.Println(hex.Dump(buf))
						return
					}
					for _, nal := range nalsink.Units {
						log.Printf("NAL Zero=%t, RefIdc=%d, Type=%d, Size=%d\r\n", nal.ZeroBit(), nal.RefIdc(), nal.Type(), len(nal.Data))
						// log.Println(hex.Dump(nal.Data))
						// TODO: Detect IDR (Type == 5)
						// TODO: Feed video packets to HLS/MP4/DASH emitter.
					}
				}
			} else {
				var p *rtcp.Packet
				var err error
				if p, err = rtcp.Unpack(pkt.Payload); err != nil {
					log.Println(err)
					// log.Panicln(hex.Dump(pkt.Payload))
					return
				}
			
				log.Printf("RTCP [%d] PT=%d, LN=%d, C=%d\r\n", pkt.Channel, p.PT, p.LN, p.C())
			}
		}
	}
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Panic("Wrong number of arguments.  Exactly one is expected.")
	}
	url := flag.Arg(0)

	log.SetOutput(os.Stdout)
	sess := rtsp.NewSession()
	err := sess.Open(url, rtsp.ProtoTCP)
	if err != nil {
		log.Fatalln(err)
	}
	handleStates(sess)
}
