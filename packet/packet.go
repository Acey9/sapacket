package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	//"fmt"
	//"github.com/Acey9/apacket/logp"
	"net"
)

const (
	OTHER = iota
	PACKET
	LOGIN
	LOGINSUCC
	HEARTBEAT
)

const MAXPKTLEN = 1024 * 1024 * 4

type Pkt struct {
	Len  uint32
	Type uint8
	Body []byte
}

func (pkt *Pkt) pack() ([]byte, error) {
	_body := pkt.Body
	_len := uint32(len(_body) + 5)

	if _len > MAXPKTLEN {
		err := errors.New("packet length exceeds the maximum")
		return nil, err
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(_len))
	binary.Write(buf, binary.BigEndian, pkt.Type)
	binary.Write(buf, binary.BigEndian, _body)

	return buf.Bytes(), nil
}

func Pack(_type uint8, body []byte) ([]byte, error) {
	p := Pkt{0, _type, body}
	pkt, err := p.pack()
	if err != nil {
		return nil, err
	}
	return pkt, nil
}

func Unpack(buf []byte) *Pkt {
	_len := binary.BigEndian.Uint32(buf[0:4])
	_type := uint8(buf[4])
	body := buf[5:]
	return &Pkt{_len, _type, body}
}

func WritePacket(conn net.Conn, buf []byte) error {
	bufLen := len(buf)
	bufPos := 0
	for {
		n, err := conn.Write(buf[bufPos : bufPos+bufLen])
		if err != nil {
			return err
		}
		bufLen -= n
		if bufLen <= 0 {
			break
		}

		bufPos += n
	}
	return nil
}

func ReadPacket(conn net.Conn) (*Pkt, error) {

	headLen := uint32(5)
	head := make([]byte, headLen)
	n, err := conn.Read(head)
	if err != nil || uint32(n) != headLen {
		return nil, err
	}
	//logp.Debug("packet", "ReadPacket.len: %d", n)

	pktLen := binary.BigEndian.Uint32(head[0:4])
	//logp.Debug("packet", "pktLen: %d", pktLen)
	if pktLen > MAXPKTLEN {
		err := errors.New("packet length exceeds the maximum")
		return nil, err
	}

	buf := make([]byte, pktLen)
	bufPos := headLen
	buf[0] = head[0]
	buf[1] = head[1]
	buf[2] = head[2]
	buf[3] = head[3]
	buf[4] = head[4]

	ptype := buf[4]

	//logp.Debug("packet", "packet.type: %v", ptype)
	if ptype != PACKET && ptype != HEARTBEAT && ptype != LOGIN && ptype != LOGINSUCC {
		return nil, errors.New("pkt type error")
	}

	last_len := pktLen - headLen
	for {
		//logp.Debug("packet", "bufPos: %d\tlast_len: %d", bufPos, last_len)
		n, err := conn.Read(buf[bufPos : bufPos+last_len])
		if err != nil {
			return nil, err
		}
		bufPos += uint32(n)
		last_len -= uint32(n)
		if last_len <= 0 {
			break
		}
	}
	//logp.Debug("packet", "ReadPacket.len: %d", n)
	return Unpack(buf), nil
}

/*
func main() {
	res := Pack(HEARTBEAT, "\x43")
	fmt.Printf("% X\n", res)
	up := Unpack(res)
	fmt.Println(up)

	res = Pack(HEARTBEAT, "Spring is great")
	fmt.Printf("% X\n", res)
	up = Unpack(res)
	fmt.Println(up)
}
*/
