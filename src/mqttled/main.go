package main

import (
	"fmt"
	"os/exec"
	"flag"
	"net"
	"os"
	"log"
  "bytes"
	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
)

var conns = flag.Int("conns", 10, "how many conns (0 means infinite)")
var host = flag.String("host", "localhost:1883", "hostname of broker")
var user = flag.String("user", "", "username")
var pass = flag.String("pass", "", "password")
var dump = flag.Bool("dump", false, "dump messages?")
var wait = flag.Int("wait", 10, "ms to wait between client connects")
var pace = flag.Int("pace", 60, "send a message on average once every pace seconds")

var payload proto.Payload
var topic = "discovery"
var listenertopic = "actuator/usbled"


func Scroll(text string) {
	fmt.Println("received:" + text)
	go func() {
		snap := os.Getenv("SNAP")
		cmd := exec.Command(snap+"/dcled", "-r", "-m", text)
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
  }()
}

func main() {
	flag.Parse()

	payload = proto.BytesPayload([]byte(listenertopic))

	conn, err := net.Dial("tcp", *host)
	if err != nil {
		fmt.Fprint(os.Stderr, "dial: ", err)
		return
	}
	cc := mqtt.NewClientConn(conn)
	cc.Dump = *dump
	cc.ClientId = "mqttled"



	if err := cc.Connect(*user, *pass); err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}



	// Sender
	cc.Publish(&proto.Publish{
		Header:    proto.Header{Retain: false},
		TopicName: topic,
		Payload:   proto.BytesPayload([]byte(listenertopic)),
	})

	tq := []proto.TopicQos{
		{Topic: listenertopic, Qos: proto.QosAtLeastOnce},
	}

	// Receiver
	  cc.Subscribe(tq)
		for m := range cc.Incoming {
			buf := new(bytes.Buffer)
			m.Payload.WritePayload(buf)
			msg := buf.String()
			Scroll(msg)
		}



}
