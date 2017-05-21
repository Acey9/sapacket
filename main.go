package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/Acey9/apacket/logp"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"
)

const version = "0.1"

var spacket Sapacket

type Sapacket struct {
	ListenIP   string
	ListenPort uint16
	CertFile   string
	KeyFile    string
	Logging    *logp.Logging
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
		conn.Close()
	}()

	for {
		conn.SetDeadline(time.Now().Add(180 * time.Second))
		pkt, err := ReadPacket(conn)
		if err != nil {
			logp.Err("%s read pkt err: %v", conn.RemoteAddr(), err)
			return
		}
		if pkt.Type != PACKET {
			logp.Err("%s pkt type", conn.RemoteAddr())
			return
		}
		logp.Info("pkt %s", pkt.Body)
	}
}

func optParse() {
	var logging logp.Logging
	var fileRotator logp.FileRotator
	var rotateEveryKB uint64
	var keepFiles int
	var port uint

	flag.StringVar(&spacket.ListenIP, "b", "0.0.0.0", "Listen address")
	flag.UintVar(&port, "p", 15444, "Listen port")
	spacket.ListenPort = uint16(port)

	flag.StringVar(&logging.Level, "l", "info", "logging level")
	flag.StringVar(&fileRotator.Path, "lp", "", "log path")
	flag.StringVar(&fileRotator.Name, "n", "sapacket.log", "log name")
	flag.Uint64Var(&rotateEveryKB, "r", 10240, "rotate every KB")
	flag.IntVar(&keepFiles, "k", 7, "number of keep files")

	flag.StringVar(&spacket.CertFile, "cf", "", "X509 cert file")
	flag.StringVar(&spacket.KeyFile, "kf", "", "X509 key file")

	printVersion := flag.Bool("V", false, "version")

	flag.Parse()
	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if spacket.CertFile == "" || spacket.KeyFile == "" {
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
	fmt.Println("vim-go")
}
