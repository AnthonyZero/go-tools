package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type NewCodecFunc func(closer io.ReadWriteCloser) Codec //编码函数

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

var NewCodecFuncMap map[Type]NewCodecFunc
var _ Codec = (*GobCodec)(nil) //check

type Header struct {
	ServiceMethod string //服务名和方法名
	Seq           uint64
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(header *Header) error
	ReadBody(body interface{}) error
	Write(header *Header, body interface{}) error
}

type GobCodec struct {
	conn io.ReadWriteCloser //conn 是由构建函数传入，通常是通过 TCP 或者 Unix 建立 socket 时得到的链接实例
	buf  *bufio.Writer      //buf 是为了防止阻塞而创建的带缓冲的 Writer，一般这么做能提升性能。
	dec  *gob.Decoder       //gob 的 Decoder 和 Encoder
	enc  *gob.Encoder
}

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

func (g *GobCodec) Close() error {
	return g.conn.Close()
}

func (g *GobCodec) ReadHeader(header *Header) error {
	return g.dec.Decode(header)
}

func (g *GobCodec) ReadBody(body interface{}) error {
	return g.dec.Decode(body)
}

func (g *GobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	if err := g.enc.Encode(header); err != nil {
		log.Printf("rpc codec: gob error encoding header: %v", err)
		return err
	}
	if err := g.enc.Encode(body); err != nil {
		log.Printf("rpc codec: gob error encoding body: %v", err)
		return err
	}
	return nil
}
