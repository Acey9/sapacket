package test

import (
	"testing"
	//"github.com/stretchr/testify/assert"
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"github.com/Acey9/sapacket/packet"
	"time"
)

func TestLoginSucc(t *testing.T) {
	token := "54321"
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:5444", conf)
	if err != nil {
		t.Error("Bad result", err)
		return
	}

	defer conn.Close()

	conn.SetDeadline(time.Now().Add(30 * time.Second))
	login, err := packet.Pack(packet.LOGIN, []byte(token))
	if err != nil {
		t.Error("Bad result", err)
		return
	}
	err = packet.WritePacket(conn, login)
	if err != nil {
		//logp.Err("login faield. %v", err)
		t.Error("Bad result", err)
		return
	}

	conn.SetDeadline(time.Now().Add(30 * time.Second))
	pkt, err := packet.ReadPacket(conn)
	if err != nil {
		//logp.Err("login faield. %v", err)
		t.Error("Bad result", err)
		return
	}

	if pkt.Type != packet.LOGINSUCC {
		//logp.Err("login faield. %v", err)
		t.Error("Bad result", "login faield")
		return
	}
	t.Log("TestLoginSucc, succ")
}

func TestLoginFailed(t *testing.T) {
	token := "543210"
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:5444", conf)
	if err != nil {
		t.Error("Bad result", err)
		return
	}

	defer conn.Close()

	conn.SetDeadline(time.Now().Add(30 * time.Second))
	login, err := packet.Pack(packet.LOGIN, []byte(token))
	if err != nil {
		t.Error("Bad result", err)
		return
	}
	err = packet.WritePacket(conn, login)
	if err != nil {
		//logp.Err("login faield. %v", err)
		t.Error("Bad result", err)
		return
	}

	conn.SetDeadline(time.Now().Add(30 * time.Second))
	pkt, err := packet.ReadPacket(conn)
	if err != nil {
		//logp.Err("login faield. %v", err)
		t.Log("TestLoginFailed, succ")
		return
	}
	if pkt.Type == packet.LOGINSUCC {
		//logp.Err("login faield. %v", err)
		t.Error("Bad result", "packet.LOGINSUCC")
		return
	}
	t.Error("Bad result", "LOGINSUCC")
}

func TestPack(t *testing.T) {
	_, err := packet.Pack(packet.PACKET, []byte("abcd"))
	if err != nil {
		t.Error("Bad result", err)
	}
	t.Log("TestPack", "normal pack succ")

	var buf bytes.Buffer

	for i := 1; i < 20000; i++ {
		buf.Write([]byte("hello"))
	}
	_, err = packet.Pack(packet.PACKET, buf.Bytes())
	if err != nil {
		t.Log("TestPack", "abnormal pack succ:", err)
		return
	}
	t.Error("Bad result", err)
}

func TestSendData(t *testing.T) {
	token := "54321"
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:5444", conf)
	if err != nil {
		t.Error("Bad result", err)
		return
	}

	defer conn.Close()

	conn.SetDeadline(time.Now().Add(30 * time.Second))
	login, err := packet.Pack(packet.LOGIN, []byte(token))
	if err != nil {
		t.Error("Bad result", err)
		return
	}
	err = packet.WritePacket(conn, login)
	if err != nil {
		//logp.Err("login faield. %v", err)
		t.Error("Bad result", err)
		return
	}

	conn.SetDeadline(time.Now().Add(30 * time.Second))
	pkt, err := packet.ReadPacket(conn)
	if err != nil {
		//logp.Err("login faield. %v", err)
		t.Error("Bad result", err)
		return
	}

	if pkt.Type != packet.LOGINSUCC {
		//logp.Err("login faield. %v", err)
		t.Error("Bad result", "login faield")
		return
	}

	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write([]byte("abcd"))
	w.Close()

	fmt.Println("normal pkt.body len: ", len(buf.Bytes()))
	msg, err := packet.Pack(packet.PACKET, buf.Bytes())
	if err != nil {
		t.Error("Bad result", err)
	}
	err = packet.WritePacket(conn, msg)
	if err != nil {
		t.Error("Bad result", err)
	}
	fmt.Println("send abcd end")

	bigbuf := new(bytes.Buffer)
	binary.Write(bigbuf, binary.BigEndian, 100)
	binary.Write(bigbuf, binary.BigEndian, 1)

	for i := 1; i < 200; i++ {
		binary.Write(bigbuf, binary.BigEndian, []byte("hello"))
	}
	fmt.Println("abnormal pkt len: ", len(bigbuf.Bytes()))

	var bbuf bytes.Buffer
	w2 := zlib.NewWriter(&bbuf)
	w2.Write(bigbuf.Bytes())
	w2.Close()

	n, err := conn.Write(bbuf.Bytes())
	if err != nil {
		t.Error(n, err)
	}
}
