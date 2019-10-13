// Package mpdconn provides a simple wrapper around a TCP connection to an MPD daemon
// with very basic functionality.
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

// MpdConn MPD connection object
type MpdConn struct {
	conn net.Conn
	buf  *bufio.Reader
	url  string
}

// NewMpdConn creates a new MpdConn object to the MPD server at URL
// and checks that the connection can be established
func NewMpdConn(URL string) (*MpdConn, error) {
	m := &MpdConn{nil, nil, URL}

	err := m.establishConn()
	if err != nil {
		return nil, err
	}
	defer m.close()

	return m, nil
}

// establishConn establishes a connection. It fails if MPD does not answer OK
func (m *MpdConn) establishConn() error {
	var err error

	m.conn, err = net.Dial("tcp", m.url)
	if err != nil {
		return err
	}

	m.buf = bufio.NewReader(m.conn)
	status, err := m.buf.ReadString('\n')

	if strings.Split(status, " ")[0] != "OK" {
		return errors.New("NOT OK: " + status)
	}

	return nil
}

// close closes de underlying MPD connection
func (m MpdConn) close() {
	m.conn.Close()
}

// Request sends a request to the MPD daemon and returns the answer as a map
func (m MpdConn) Request(req string) (map[string]string, error) {
	err := m.establishConn()
	if err != nil {
		return nil, err
	}
	defer m.close()

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
}

// readResponse reads MPD response and parses it as type and value
func (m MpdConn) readResponse() (string, string, error) {
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

// DownloadCover downloads the cover for the song specified into the given file
func (m MpdConn) DownloadCover(name string, file *os.File) error {
	err := m.establishConn()
	if err != nil {
		return err
	}
	defer m.close()

	offset, size, bsize := 0, 1, 0

	if err = file.Truncate(0); err != nil {
		return err
	}
	if _, err = file.Seek(0, 0); err != nil {
		return err
	}

	w := bufio.NewWriter(file)
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
