// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func readUntil(filename, marker string) []byte {
	file, err := os.ReadFile(filename) // just pass the file name
	if err != nil {
		panic(err)
	}
	startIndex := strings.Index(string(file), marker)
	if startIndex < 0 {
		return nil
	}
	return file[:startIndex]
}

func readFrom(filename, marker string) []byte {
	file, err := os.ReadFile(filename) // just pass the file name
	if err != nil {
		panic(err)
	}
	stopIndex := strings.Index(string(file), marker)
	if stopIndex < 0 {
		return nil
	}
	markerLength := len(marker)
	return file[(stopIndex + markerLength):]
}

func updateMarkers(filename, startMarker, endMarker string, content []byte) []byte {
	firstPart := readUntil(filename, startMarker)
	if firstPart == nil {
		// Append at the end of the current file
		file, err := os.ReadFile(filename)
		if err != nil {
			fmt.Print(err)
		}
		var buf bytes.Buffer
		buf.Write(file)
		buf.Write([]byte("\n"))
		buf.Write([]byte(startMarker))
		buf.Write([]byte("\n"))
		buf.Write(content)
		buf.Write([]byte(endMarker))
		return buf.Bytes()
	}

	lastPart := readFrom(filename, endMarker)
	if lastPart == nil {
		panic(fmt.Sprintf("end marker %s not found in %s", endMarker, filename))
	}

	var buf bytes.Buffer
	buf.Write(firstPart)
	buf.Write([]byte(startMarker))
	buf.Write([]byte("\n"))
	buf.Write(content)
	buf.Write([]byte(endMarker))
	buf.Write(lastPart)
	return buf.Bytes()
}
