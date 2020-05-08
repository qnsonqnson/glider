package ss

import (
	"errors"
	"net"

	"github.com/nadoo/glider/common/pool"
	"github.com/nadoo/glider/common/socks"
)

// PktConn .
type PktConn struct {
	net.PacketConn

	writeAddr net.Addr // write to and read from addr

	tgtAddr   socks.Addr
	tgtHeader bool
}

// NewPktConn returns a PktConn
func NewPktConn(c net.PacketConn, writeAddr net.Addr, tgtAddr socks.Addr, tgtHeader bool) *PktConn {
	pc := &PktConn{
		PacketConn: c,
		writeAddr:  writeAddr,
		tgtAddr:    tgtAddr,
		tgtHeader:  tgtHeader}
	return pc
}

// ReadFrom overrides the original function from net.PacketConn
func (pc *PktConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if !pc.tgtHeader {
		return pc.PacketConn.ReadFrom(b)
	}

	buf := pool.GetBuffer(len(b))
	defer pool.PutBuffer(buf)

	n, raddr, err := pc.PacketConn.ReadFrom(buf)
	if err != nil {
		return n, raddr, err
	}

	tgtAddr := socks.SplitAddr(buf)
	if tgtAddr == nil {
		return n, raddr, errors.New("can not get addr")
	}
	copy(b, buf[len(tgtAddr):])

	//test
	if pc.writeAddr == nil {
		pc.writeAddr = raddr
	}

	if pc.tgtAddr == nil {
		pc.tgtAddr = tgtAddr
	}

	return n - len(tgtAddr), raddr, err
}

// WriteTo overrides the original function from net.PacketConn
func (pc *PktConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if !pc.tgtHeader {
		return pc.PacketConn.WriteTo(b, addr)
	}

	buf := pool.GetBuffer(len(pc.tgtAddr) + len(b))
	pool.PutBuffer(buf)

	copy(buf, pc.tgtAddr)
	copy(buf[len(pc.tgtAddr):], b)

	return pc.PacketConn.WriteTo(buf, pc.writeAddr)
}
