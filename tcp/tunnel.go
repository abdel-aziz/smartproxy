package tcp

import (
	"io"

	"git.blueboard.it/blueboard/proxy/common"
)

var BUF_CONNECT_SIZE = len([]byte(CONNECT_OK))

func (p *Proxy) initialBody(msg []byte) error {
	index, err := common.EndHeaders(msg)
	if err != nil {
		return err
	}
	initialBody := msg[index:]
	if len(initialBody) > 0 {
		_, err = p.connSrv.Write(initialBody)
		return err
	}
	return nil
}

func (p *Proxy) pipe(dst io.WriteCloser, src io.ReadCloser, errs chan error) {
	for {
		msg := make([]byte, BUF_SIZE)
		nb, err := src.Read(msg)
		if err != nil {
			if p.err == nil {
				p.err = err
			}
			errs <- err
			return
		}
		p.received += nb
		p.sent += nb
		_, err = dst.Write(msg[:nb])
		if err != nil {
			if p.err == nil {
				p.err = err
			}
			errs <- err
			return
		}
	}
}
