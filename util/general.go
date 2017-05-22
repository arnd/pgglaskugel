// Package util - general module
// Copyright © 2017 Alexander Sosna <alexander@xxor.de>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package util

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// AnswerConfirmation is a wrapper to parse possible ser answers
func AnswerConfirmation(msg string) (confirmed bool, err error) {
	var input string
	log.Warn(msg)

	_, err = fmt.Scanln(&input)
	if err != nil {
		return false, err
	}
	positive := []string{"j", "ja", "y", "yes", "do it", "let's rock"}
	negative := []string{"nein", "n", "no", "hell no", "fuck off"}

	input = strings.ToLower(input)

	for _, element := range positive {
		if element == input {
			return true, nil
		}
	}

	for _, element := range negative {
		if element == input {
			return false, nil
		}
	}
	doesNotParse := errors.New("Answer can not be parsed: " + input)
	return false, doesNotParse
}

// MustAnswerConfirmation to evaluate the answers from a user, if necessary
func MustAnswerConfirmation(msg string) (confirmed bool) {
	if confirmed, err := AnswerConfirmation(msg); err == nil {
		return confirmed
	}
	return MustAnswerConfirmation(msg)
}

// WatchOutput logs debug output
func WatchOutput(input io.Reader, outputFunc func(args ...interface{}), done chan struct{}) {
	log.Debug("watchOutput started")
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		outputFunc(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		outputFunc("error reading input:", err)
	}
	log.Debug("watchOutput end")
	done <- struct{}{}
}

// Exists returns whether the given file or directory exists or not
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// IsEmpty returns true if the given dir is empty
func IsEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Readdirnames returns at most n names, or io.EOF if not enough are available
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// StreamToByte takes an io.Reader and returns []byte
func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}
