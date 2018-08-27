package webrtc

import (
	"fmt"
	"sync"
	"time"

	"strconv"

	"github.com/pions/webrtc"
	"github.com/pions/webrtc/pkg/datachannel"
)

func createClient(name string, offerCh chan webrtc.RTCSessionDescription, answerCh chan webrtc.RTCSessionDescription, master bool) {
	peerConnection, err := webrtc.New(webrtc.RTCConfiguration{
		ICEServers: []webrtc.RTCICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	if err != nil {
		panic(err)
	}

	// peerConnection.OnICEConnectionStateChange = func(connectionState ice.ConnectionState) {
	// 	fmt.Printf("Connection State has changed %s \n", connectionState.String())
	// }

	datachannels := make([]*webrtc.RTCDataChannel, 0)
	var dataChannelsLock sync.RWMutex

	peerConnection.Ondatachannel = func(d *webrtc.RTCDataChannel) {
		dataChannelsLock.Lock()
		datachannels = append(datachannels, d)
		dataChannelsLock.Unlock()
		fmt.Printf("New DataChannel %s %d\n", d.Label, d.ID)

		d.Lock()
		defer d.Unlock()
		d.Onmessage = func(payload datachannel.Payload) {
			switch p := payload.(type) {
			case *datachannel.PayloadString:
				fmt.Printf("Message '%s' from DataChannel '%s' payload '%s'\n", p.PayloadType().String(), d.Label, string(p.Data))
			case *datachannel.PayloadBinary:
				fmt.Printf("Message '%s' from DataChannel '%s' payload '% 02x'\n", p.PayloadType().String(), d.Label, p.Data)
			default:
				fmt.Printf("Message '%s' from DataChannel '%s' no payload \n", p.PayloadType().String(), d.Label)
			}
		}
	}
	if master {
		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			panic(err)
		}
		// fmt.Println("", name, "webrtc offer ", offer.Sdp)
		offerCh <- offer

		answer := <-answerCh
		if err := peerConnection.SetRemoteDescription(answer); err != nil {
			panic(err)
		}
	} else {
		offer := <-offerCh
		if err := peerConnection.SetRemoteDescription(offer); err != nil {
			panic(err)
		}
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}
		//fmt.Println("", name, "webrtc answer ", answer.Sdp)
		answerCh <- answer
	}

	fmt.Println("Random messages will now be sent to any connected DataChannels every 5 seconds")
	counter := uint64(0)
	for {
		time.Sleep(5 * time.Second)
		message := name + " " + strconv.FormatUint(counter, 10)
		fmt.Printf("%s is sending '%s' \n", name, message)
		dataChannelsLock.RLock()
		for _, d := range datachannels {
			err := d.Send(datachannel.PayloadString{Data: []byte(message)})
			if err != nil {
				panic(err)
			}
		}
		dataChannelsLock.RUnlock()
		counter++
	}
}

func WebRTC_main() {
	webrtc.RegisterDefaultCodecs()
	offer := make(chan webrtc.RTCSessionDescription)
	answer := make(chan webrtc.RTCSessionDescription)
	go createClient("Bob", offer, answer, false)
	createClient("Alice", offer, answer, true)
}
