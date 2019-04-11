// Copyright © 2019 Matt Konda <mkonda@jemurai.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package encrypt

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// Possible reference...
// https://gist.github.com/ayubmalik/a83ee23c7c700cdce2f8c5bf5f2e9f20

// Encrypt the file with public key provided.
func Encrypt(filename string, pubkey string) {
	key := getKey(pubkey)
	log.Println("Public key:", key)

	// Read in public key
	recipient, err := readEntity(key)
	if err != nil {
		fmt.Println(err)
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	dst, err := os.Create(filename + ".gpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dst.Close()
	encrypt([]*openpgp.Entity{recipient}, nil, f, dst)
}

func getKey(keypath string) string {
	if isValidURL(keypath) {
		return "url..."
	}
	return keypath
}

func encrypt(recip []*openpgp.Entity, signer *openpgp.Entity, r io.Reader, w io.Writer) error {
	wc, err := openpgp.Encrypt(w, recip, signer, &openpgp.FileHints{IsBinary: true}, nil)
	if err != nil {
		return err
	}
	if _, err := io.Copy(wc, r); err != nil {
		return err
	}
	return wc.Close()
}

func readEntity(name string) (*openpgp.Entity, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	block, err := armor.Decode(f)
	if err != nil {
		return nil, err
	}
	return openpgp.ReadEntity(packet.NewReader(block.Body))
}

func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}
