// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package passwd

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

// open opens /etc/password for reading.
// abstracted so that it can be mocked for tests.
var open = func() (io.ReadCloser, error) {
	return os.Open("/etc/passwd")
}

// PasswdEntry represents an entry in /etc/passwd.
type Entry struct {
	Username string
	UID      int
	GID      int
	Home     string
	Shell    string
}

// Walk iterates over all entries in /etc/passwd and calls f for each one
// f returns whether the walk should continue - i.e. returning false from f will
// exit the walk without processing more entries.
func Walk(f func(Entry) bool) error {
	file, err := open()
	if err != nil {
		return err
	}
	defer file.Close()
	s := bufio.NewScanner(file)
	for s.Scan() {
		entry, err := parse(s.Text())
		if err != nil {
			return err
		}
		if !f(entry) {
			return nil
		}
	}
	return nil
}

func parse(s string) (Entry, error) {
	parts := strings.Split(s, ":")
	name := parts[0]
	// parts[1] - ignore password info
	uid, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		return Entry{}, err
	}
	gid, err := strconv.ParseInt(parts[3], 10, 32)
	if err != nil {
		return Entry{}, nil
	}
	// parts[4] - ignore GECOS
	home := parts[5]
	shell := parts[6]
	return Entry{
		Username: name,
		UID:      int(uid),
		GID:      int(gid),
		Home:     home,
		Shell:    shell,
	}, nil
}
