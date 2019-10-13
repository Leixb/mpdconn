# mpdconn
![GitHub](https://img.shields.io/github/license/Leixb/mpdconn)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/Leixb/mpdconn)
[![Build Status](https://travis-ci.com/Leixb/jutge.svg?branch=master)](https://travis-ci.com/Leixb/mpdconn)
[![Go Report Card](https://goreportcard.com/badge/github.com/Leixb/jutge)](https://goreportcard.com/report/github.com/Leixb/mpdconn)
[![GoDoc](https://godoc.org/github.com/Leixb/jutge?status.svg)](https://godoc.org/github.com/Leixb/mpdconn)

Simple wrapper around a TCP connection to an MPD daemon with very basic functionality. 
This library is used by [Leixb/MPD_goclient](https://github.com/Leixb/MPD_goclient)

It can make requests and return the reponse as a map and download an album cover to a file.

## Example

``` go
package main

import (
	"fmt"
	"github.com/Leixb/mpdconn"
)

func main() {

	MPDCon, err := mpdconn.NewMpdConn("localhost:6600")
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := MPDCon.Request("status")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(resp)

}
```
```
map[audio:44100:16:2 bitrate:1167 consume:0 duration:203.782 elapsed:72.215 mixrampdb:0.000000 nextsong:116 nextsongid:117 playlist:2 playlistlength:120 random:0 repeat:0 single:0 song:115 songid:116 state:play time:72:204]
```
