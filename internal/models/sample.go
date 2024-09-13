/*******************************************************************************
 * Copyright 2021 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package models

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	"io/ioutil"
	"math/rand"
	"time"
)

const alphanumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type SampleData struct {
	Description string `json:"description,omitempty"`
	Seed        string `json:"seed,omitempty"`
	Signature   string `json:"signature,omitempty"`
}

//function reads a cryptographic key from a file, generates random data (description and seed), signs the seed using Ed25519, 
//and returns a SampleData object containing the sample random data
func NewSampleData(cfg config.KeyInfo) (SampleData, error) {
	key, err := ioutil.ReadFile(cfg.Path)  //cfg.Path contains information (path) of the key to be used (Private key in this case)
	if err != nil {
		return SampleData{}, err
	}

	x := SampleData{
		Description: factoryRandomFixedLengthString(128, alphanumericCharset),   //random alphanumeric string of length 128 characters, generated using factoryRandomFixedLengthString
		Seed:        factoryRandomFixedLengthString(64, alphanumericCharset),    //random alphanumeric string of length 64 characters, generated using factoryRandomFixedLengthString
	}

	keyDecoded := make([]byte, hex.DecodedLen(len(key)))   //Creates a byte slice keyDecoded that has enough length to hold the decoded key.
	hex.Decode(keyDecoded, key)      n                     //Decodes the hex-encoded key into bytes and stores in keyDecoded
	signed := ed25519.Sign(keyDecoded, []byte(x.Seed))     //Signs the Seed using the decoded private key (keyDecoded) with the Ed25519 signature algorithm.
	x.Signature = fmt.Sprintf("%x", signed)                //signature is formatted as hex string and stored in x Sample data
	return x, nil
}

func factoryRandomFixedLengthString(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
