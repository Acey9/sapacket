package main

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/Acey9/apacket/logp"
	"github.com/Acey9/sapacket/packet"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"
)

const version = "1.0"

var spacket Sapacket

type Sapacket struct {
	ListenIP   string
	ListenPort uint16
	CertFile   string
	KeyFile    string
	Token      string
	Logging    *logp.Logging
	Timeout    uint16
}

func (this *Sapacket) sayHi() {
	fmt.Println("apacket server version: ", version)
}

func (this *Sapacket) start() {
	cer, err := tls.LoadX509KeyPair(this.CertFile, this.KeyFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	addr := bytes.Buffer{}
	addr.WriteString(this.ListenIP)
	addr.WriteString(":")
	addr.WriteString(strconv.Itoa(int(this.ListenPort)))
	server, err := tls.Listen("tcp", addr.String(), config)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}

		go this.initHandler(conn)
	}
	fmt.Println("Stopped accepting data")
}

func (this *Sapacket) initHandler(conn net.Conn) {

	defer func() {
		if err := recover(); err != nil {
			logp.Err("%v", err)
		}
		logp.Info("client %s lost", conn.RemoteAddr())
		conn.Close()
	}()

	conn.SetDeadline(time.Now().Add(60 * time.Second))
	pkt, err := packet.ReadPacket(conn)
	if err != nil {
		return
	}

	if pkt.Type != packet.LOGIN || string(pkt.Body) != this.Token {
		return
	}

	succ, err := packet.Pack(packet.LOGINSUCC, []byte(""))
	if err != nil {
		return
	}
	conn.SetDeadline(time.Now().Add(60 * time.Second))
	err = packet.WritePacket(conn, succ)
	if err != nil {
		logp.Err("client %s login err [%v]", conn.RemoteAddr(), err)
		return
	}

	logp.Info("client %s join", conn.RemoteAddr())

	for {

		conn.SetDeadline(time.Now().Add(time.Duration(this.Timeout) * time.Second))
		pkt, err = packet.ReadPacket(conn)
		if err != nil {
			logp.Err("client %s info [%v]", conn.RemoteAddr(), err)
			return
		}
		if pkt.Type != packet.PACKET {
			logp.Err("client %s error pkt type", conn.RemoteAddr())
			return
		}

		var out bytes.Buffer
		var in bytes.Buffer

		in.Write([]byte(pkt.Body))

		r, err := zlib.NewReader(&in)
		if err != nil {
			logp.Err("decode error [%v]", err)
			return
		}
		io.Copy(&out, r)
		r.Close()

		//fmt.Println(len(pkt.Body), len(out.String()))
		logp.Info("pkt %s", out.String())
	}
}

func optParse() {
	var logging logp.Logging
	var fileRotator logp.FileRotator
	var rotateEveryKB uint64
	var keepFiles int
	var port uint
	var timeout int

	flag.StringVar(&spacket.ListenIP, "b", "0.0.0.0", "Listen address")
	flag.UintVar(&port, "p", 5444, "Listen port")
	flag.StringVar(&spacket.Token, "a", "", "auth token")

	flag.StringVar(&logging.Level, "l", "info", "logging level")
	flag.StringVar(&fileRotator.Path, "lp", "", "log path")
	flag.StringVar(&fileRotator.Name, "n", "sapacket.log", "log name")
	flag.Uint64Var(&rotateEveryKB, "r", 10240, "rotate every KB")
	flag.IntVar(&keepFiles, "k", 7, "number of keep files")

	flag.IntVar(&timeout, "t", 900, "the client connection timed out. unit is second")

	flag.StringVar(&spacket.CertFile, "cf", "", "X509 cert file")
	flag.StringVar(&spacket.KeyFile, "kf", "", "X509 key file")

	printVersion := flag.Bool("V", false, "version")

	flag.Parse()
	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	spacket.ListenPort = uint16(port)
	spacket.Timeout = uint16(timeout)

	if spacket.CertFile == "" || spacket.KeyFile == "" || spacket.Token == "" {
		flag.Usage()
		os.Exit(1)
	}

	logging.Files = &fileRotator
	if logging.Files.Path != "" {
		tofiles := true
		logging.ToFiles = &tofiles

		rotateKB := rotateEveryKB * 1024
		logging.Files.RotateEveryBytes = &rotateKB
		logging.Files.KeepFiles = &keepFiles
	}
	spacket.Logging = &logging
}

func init() {
	optParse()
	logp.Init("sapacket", spacket.Logging)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	spacket.sayHi()
	spacket.start()

}
