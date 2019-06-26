package mpdconn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type MPDconn struct {
	conn net.Conn
	buf  *bufio.Reader
	url  string
}

func NewMPDconn(URL string) (*MPDconn, error) {

	m := new(MPDconn)

	m.url = URL

	err := m.EstablishConn()
	defer m.Close()

	if err != nil {
		return nil, err
	}

	return m, nil

}

func (m *MPDconn) EstablishConn() error {

	conn, err := net.Dial("tcp", m.url)
	if err != nil {
		return err
	}

	m.conn = conn

	m.buf = bufio.NewReader(m.conn)

	status, err := m.buf.ReadString('\n')

	s := strings.Split(status, " ")

	if s[0] != "OK" {
		return errors.New("NOT OK" + status)
	}

	return nil

}

func (m MPDconn) Close() {
	m.conn.Close()
}

func (m MPDconn) Request(req string) (map[string]string, error) {

	err := m.EstablishConn()

	if err != nil {
		return nil, err
	}

	defer m.Close()

	req = strings.TrimSuffix(req, "\n") // remove \n so we have no duplicates
	_, err = m.conn.Write([]byte(req + "\n"))

	if err != nil {
		return nil, err
	}

	resp := make(map[string]string)

	for {
		dtype, value, err := m.readResponse()
		if err != nil {
			return nil, err
		}

		switch dtype {
		case "OK":
			return resp, nil
		case "ACK":
			return resp, errors.New(value)
		case "binary":
			bsize, err := strconv.Atoi(value)
			if err != nil {
				return resp, err
			}

			fbuf := make([]byte, bsize)

			_, err = io.ReadFull(m.buf, fbuf)
			if err != nil {
				return resp, err
			}
		}

		resp[dtype] = value

	}

	return resp, nil
}

func (m MPDconn) readResponse() (string, string, error) {

	data, err := m.buf.ReadString('\n')
	if err != nil {
		return data, data, err
	}

	data = strings.TrimSuffix(data, "\n")
	_data := strings.SplitN(data, " ", 2)

	value := ""

	if len(_data) >= 2 {
		value = _data[1]
	}
	dtype := strings.TrimSuffix(_data[0], ":")

	return dtype, value, nil

}

func (m MPDconn) DownloadCover(name string, file string) error {

	err := m.EstablishConn()
	if err != nil {
		return err
	}
	defer m.Close()

	offset, size, bsize := 0, 1, 0

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for offset < size {

		//Make coverart request
		fmt.Fprintf(m.conn, "albumart \"%s\" %d\n", name, offset)

		for {

			dtype, value, err := m.readResponse()
			if err != nil {
				return err
			}

			if dtype == "OK" {
				break // cannot break inside switch...
			}

			switch dtype {

			case "ACK":
				return errors.New(value)

			case "size":
				size, err = strconv.Atoi(value)
				if err != nil {
					return err
				}

			case "binary":
				bsize, err = strconv.Atoi(value)
				if err != nil {
					return err
				}

				fbuf := make([]byte, bsize)

				n, err := io.ReadFull(m.buf, fbuf)
				if err != nil {
					return err
				}

				_, err = w.Write(fbuf)
				if err != nil {
					return err
				}

				offset = offset + n
			}
		}

	}

	return nil

}
