/*
 * Copyright (c) SAS Institute Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpmutils

import (
	"bytes"
	"crypto"
	"hash"
	"io"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// SignatureOptions describes additional configuration for SignRpm methods
type SignatureOptions struct {
	// Hash algorithm for the signature. If not set, defaults to SHA-256
	Hash crypto.Hash
	// CreationTime for the signature. If not set, defaults to the current time
	CreationTime time.Time
}

func (opts *SignatureOptions) hash() crypto.Hash {
	if opts != nil {
		return opts.Hash
	}
	return crypto.SHA256
}

func (opts *SignatureOptions) creationTime() time.Time {
	if opts != nil {
		return opts.CreationTime
	}
	return time.Now()
}

func makeSignature(h hash.Hash, key *packet.PrivateKey, opts *SignatureOptions) ([]byte, error) {
	hashType := opts.hash()
	sig := &packet.Signature{
		SigType:      packet.SigTypeBinary,
		CreationTime: opts.creationTime(),
		PubKeyAlgo:   key.PubKeyAlgo,
		Hash:         hashType,
		IssuerKeyId:  &key.KeyId,
	}
	err := sig.Sign(h, key, nil)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = sig.Serialize(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func insertSignature(sigHeader *rpmHeader, tag int, value []byte) {
	sigHeader.entries[tag] = entry{
		dataType: RPM_BIN_TYPE,
		count:    int32(len(value)),
		contents: value,
	}
}

func insertSignatures(sigHeader *rpmHeader, sigPgp, sigRsa []byte) {
	insertSignature(sigHeader, SIG_PGP-_SIGHEADER_TAG_BASE, sigPgp)
	insertSignature(sigHeader, SIG_RSA, sigRsa)
	delete(sigHeader.entries, SIG_GPG-_SIGHEADER_TAG_BASE)
	delete(sigHeader.entries, SIG_DSA)
}

func getSha1(sigHeader *rpmHeader) string {
	vals, err := sigHeader.GetStrings(SIG_SHA1)
	if err != nil {
		return ""
	}
	return vals[0]
}

func getSha256(sigHeader *rpmHeader) string {
	vals, err := sigHeader.GetStrings(SIG_SHA256)
	if err != nil {
		return ""
	}
	return vals[0]
}

func getHashAndType(sigHeader *rpmHeader) (string, crypto.Hash) {
	// RPM v4 introduced SHA256 header signatures, prefer them over the
	// previous SHA1
	if h := getSha256(sigHeader); h != "" {
		return h, crypto.SHA256
	}
	return getSha1(sigHeader), crypto.SHA1
}

func digestForSigning(sigHeader, genHeader *rpmHeader, payloadReader io.Reader, opts *SignatureOptions) (genHash, combinedHash hash.Hash, err error) {
	genHash, combinedHash = opts.hash().New(), opts.hash().New()
	// write header
	genHash.Write(genHeader.orig)
	combinedHash.Write(genHeader.orig)
	// write and verify payload
	err = digestPayload(sigHeader, genHeader, payloadReader, []io.Writer{combinedHash})
	return genHash, combinedHash, err
}

// SignRpmStream reads an RPM and signs it, returning the set of headers updated with the new signature.
func SignRpmStream(stream io.Reader, key *packet.PrivateKey, opts *SignatureOptions) (header *RpmHeader, err error) {
	lead, sigHeader, err := readSignatureHeader(stream)
	if err != nil {
		return nil, err
	}
	// parse the general header
	headerDigestValue, headerDigestType := getHashAndType(sigHeader)
	genHeader, err := readHeader(stream, headerDigestValue, headerDigestType, sigHeader.isSource, false)
	if err != nil {
		return nil, err
	}
	// hash and sign header
	genHash, combinedHash, err := digestForSigning(sigHeader, genHeader, stream, opts)
	if err != nil {
		return nil, err
	}
	// sign header and payload
	sigPgp, err := makeSignature(combinedHash, key, opts)
	if err != nil {
		return header, err
	}
	sigRsa, err := makeSignature(genHash, key, opts)
	if err != nil {
		return nil, err
	}
	insertSignatures(sigHeader, sigPgp, sigRsa)
	return &RpmHeader{
		lead:      lead,
		sigHeader: sigHeader,
		genHeader: genHeader,
		isSource:  sigHeader.isSource,
	}, nil
}

func getPayloadDigest(header *rpmHeader) (string, crypto.Hash) {
	digests, err := header.GetStrings(PAYLOADDIGEST)
	if err != nil || len(digests) == 0 {
		// no payload digest
		return "", 0
	}
	digest := digests[0]
	algos, err := header.GetUint32s(PAYLOADDIGESTALGO)
	if err != nil || len(algos) == 0 {
		return "", 0
	}
	switch algos[0] {
	case HASH_MD5:
		return digest, crypto.MD5
	case HASH_SHA1:
		return digest, crypto.SHA1
	case HASH_SHA256:
		return digest, crypto.SHA256
	case HASH_SHA384:
		return digest, crypto.SHA384
	case HASH_SHA512:
		return digest, crypto.SHA512
	case HASH_SHA224:
		return digest, crypto.SHA224
	}
	return "", 0
}
