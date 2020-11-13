package statsd

import (
	"fmt"
	"io"
	"net"
	"time"
)

const (
	timeoutDuration = 10 * time.Millisecond
)

// SafeConn is an implementation of the io.WriteCloser that wraps a net.Conn type
// its purpose is to perform a guard as a part of each Write call to first check if
// the connection is still up by performing a small read. The use case of this is to
// protect against the case where a TCP connection comes disconnected and the Write
// continues to retry for up to 15 minutes before determining that the connection has
// been broken off.
type SafeConn struct {
	netConn net.Conn
}

func NewSafeConn(network string, address string) (*SafeConn, error) {
	newConn, err := dialTimeout(network, address, 5*time.Second)
	if err != nil {
		return nil, err
	}

	c := &SafeConn{
		netConn: newConn,
	}

	return c, nil
}

func (s *SafeConn) Write(p []byte) (n int, err error) {
	// check if connection is closed
	if s.connIsClosed() {
		return 0, fmt.Errorf("connection is closed")
	}

	return s.netConn.Write(p)
}

func (s *SafeConn) Close() error {
	return s.netConn.Close()
}

func (s *SafeConn) connIsClosed() bool {
	err := s.netConn.SetReadDeadline(time.Now().Add(timeoutDuration))
	if err != nil {
		return true
	}

	one := make([]byte, 1)
	_, err = s.netConn.Read(one)
	return err == io.EOF
}
