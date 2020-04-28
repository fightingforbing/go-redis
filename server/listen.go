package server

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

type Listener struct {
	Ln      net.Listener
	File      *os.File
	Fd      int
}

func newListen(ln net.Listener) *Listener {
	listener := &Listener{}
	listener.Ln = ln
	listener.system()
	return listener
}

func (ln *Listener) system() error {
	var err error
	switch netln := ln.Ln.(type) {
	case *net.TCPListener:
		ln.File, err = netln.File()
	}
	if err != nil {
		ln.close()
		return err
	}
	ln.Fd = int(ln.File.Fd())
	return syscall.SetNonblock(ln.Fd, true)
}

func (ln *Listener) close() {
	if ln.Fd != 0 {
		syscall.Close(ln.Fd)
	}
	if ln.File != nil {
		ln.File.Close()
	}
	if ln.Ln != nil {
		ln.Ln.Close()
	}
}

func listenToPort(port int) {
	fmt.Printf("监听端口:%d\n",port)
	var err error
	listen, err := net.Listen("tcp", fmt.Sprintf("localhost:%d",port))
	if err != nil {
		panic(err)
	}
	ln := newListen(listen)
	fmt.Printf("监听套接字 fd=%d\n",ln.Fd)
	Server.Ipfd = append(Server.Ipfd, ln.Fd)
}