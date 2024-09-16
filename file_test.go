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
	_ "embed"
	"reflect"
	"testing"
)

//go:embed testdata/test.md
var s string

func Test_readUntil(t *testing.T) {
	type args struct {
		filename string
		marker   string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Marker does exist",
			args: args{
				filename: "testdata/test.md",
				marker:   StartMarker,
			},
			want: []byte(`# A sample Markdown file

## A first section

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

## My Cluster Role Documentation

`),
		},
		{
			name: "Marker does not exist",
			args: args{
				filename: "testdata/test.md",
				marker:   "non-existing-marker",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readUntil(tt.args.filename, tt.args.marker); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readUntil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readFrom(t *testing.T) {
	type args struct {
		filename string
		marker   string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Marker does exist",
			args: args{
				filename: "testdata/test.md",
				marker:   EndMarker,
			},
			want: []byte(`

## Another section

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`),
		},
		{
			name: "Marker does not exist",
			args: args{
				filename: "testdata/test.md",
				marker:   "non-existing-marker",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readFrom(tt.args.filename, tt.args.marker); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readFrom() = \n%v\nwant\n%v", string(got), string(tt.want))
			}
		})
	}
}

func Test_updateMarkers(t *testing.T) {
	type args struct {
		filename    string
		startMarker string
		endMarker   string
		content     []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Marker does exist",
			args: args{
				filename:    "testdata/test.md",
				startMarker: StartMarker,
				endMarker:   EndMarker,
				content:     []byte("Hello, World!\nThis the content that should be inserted.\nAnother line.\n"),
			},
			want: []byte(`# A sample Markdown file

## A first section

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

## My Cluster Role Documentation

<!--- START CLUSTER ROLES DOCUMENTATION --->
Hello, World!
This the content that should be inserted.
Another line.
<!--- END CLUSTER ROLES DOCUMENTATION --->

## Another section

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`),
		},
		{
			name: "Marker does not exist",
			args: args{
				filename:    "testdata/test.without.marker.md",
				startMarker: StartMarker,
				endMarker:   EndMarker,
				content:     []byte("Hello, World!\nThis the content that should be inserted.\nAnother line.\n"),
			},
			want: []byte(`# A sample Markdown file

## A first section

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

## My Cluster Role Documentation

<!--- START CLUSTER ROLES DOCUMENTATION --->
Hello, World!
This the content that should be inserted.
Another line.
<!--- END CLUSTER ROLES DOCUMENTATION --->`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateMarkers(tt.args.filename, tt.args.startMarker, tt.args.endMarker, tt.args.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateMarkers() = \n%v\nwant\n%v", string(got), string(tt.want))
				t.Errorf("updateMarkers() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}
