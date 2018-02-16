package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func fillEntry(e *Entry) *Entry {
	// Generate SHA1 for S3 filename storing
	h := sha1.New()
	io.WriteString(h, e.id)
	io.WriteString(h, e.key)
	e.sha1 = fmt.Sprintf("%x", h.Sum(nil))

	// Checking default TTL value
	if e.ttl == 0 {
		e.ttl = 60
	}

	return e
}

func parsePayload(pl []string) []Entry {
	var default_id string
	var resp []Entry
	for _, v := range pl {
		var data [4]string
		copy(data[:], strings.Split(v, ":"))

		// Default ID value for next params. We can use construction like "get 600:700 :800 :900"
		if data[0] != "" {
			default_id = data[0]
		} else if data[0] == "" {
			data[0] = default_id
		}

		// Convert string into integer for TTL
		ttl, _ := strconv.Atoi(data[3]) // TODO: Exception Handling for err

		entry := Entry{id: data[0], key: data[1], value: data[2], ttl: ttl}
		fillEntry(&entry)
		resp = append(resp, entry)
	}
	return resp
}

func parseArgs(c *Config) *Config {

	// Detection purpose type and argument position in command line
	var pos int = 9
DetectType:
	for i, v := range os.Args {
		switch v {
		case "var":
			c.purpose, pos = false, i
			break DetectType
		case "file":
			c.purpose, pos = true, i
			break DetectType
		}
	}

	if pos == 9 {
		os.Stderr.WriteString("Only file and env!")
		os.Exit(1)
	}

	// Select action, put or get
	switch os.Args[pos+1] {
	case "get":
		c.action = 0
	case "set":
		c.action = 1
	case "rm":
		c.action = 2
	default:
		os.Stderr.WriteString("Only put/set/rm params are available!")
		os.Exit(1)
	}

	// Be sure that after purpose we have enough args.
	if len(os.Args[pos+1:]) < 2 { // TODO TEST
		os.Stderr.WriteString("Not enoung arguments!")
		os.Exit(1)
	}

	c.entries = parsePayload(os.Args[pos+2:])
	return c
}
