package signer

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/circl/sign/mldsa/mldsa44"
)

// KeyStore is the interface for persisting keys and tracking the active key.
type KeyStore interface {
	Save(kp *HybridKeyPair) error
	Load(kid string) (*HybridKeyPair, error)
	LoadAll() ([]*HybridKeyPair, error)
	SetActiveKid(kid string) error
	GetActiveKid() (string, error)
}

// SerializedKeyPair is the JSON-serialisable form of a HybridKeyPair.
type SerializedKeyPair struct {
	Kid            string `json:"kid"`
	Ed25519Public  []byte `json:"ed25519_public"`
	Ed25519Private []byte `json:"ed25519_private"`
	MLDSAPublic    []byte `json:"mldsa44_public"`
	MLDSAPrivate   []byte `json:"mldsa44_private"`
}

// MarshalKeyPair serialises a HybridKeyPair to JSON bytes.
func MarshalKeyPair(kp *HybridKeyPair) ([]byte, error) {
	s := SerializedKeyPair{
		Kid:            kp.Kid,
		Ed25519Public:  []byte(kp.Ed25519Public),
		Ed25519Private: []byte(kp.Ed25519Private),
		MLDSAPublic:    kp.MLDSAPublic.Bytes(),
		MLDSAPrivate:   kp.MLDSAPrivate.Bytes(),
	}
	return json.Marshal(s)
}

// UnmarshalKeyPair deserialises a HybridKeyPair from JSON bytes.
func UnmarshalKeyPair(data []byte) (*HybridKeyPair, error) {
	var s SerializedKeyPair
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("signer: unmarshalling key pair: %w", err)
	}

	mlPub := new(mldsa44.PublicKey)
	if err := mlPub.UnmarshalBinary(s.MLDSAPublic); err != nil {
		return nil, fmt.Errorf("signer: unmarshalling ML-DSA-44 public key: %w", err)
	}

	mlPriv := new(mldsa44.PrivateKey)
	if err := mlPriv.UnmarshalBinary(s.MLDSAPrivate); err != nil {
		return nil, fmt.Errorf("signer: unmarshalling ML-DSA-44 private key: %w", err)
	}

	return &HybridKeyPair{
		Kid:            s.Kid,
		Ed25519Public:  ed25519.PublicKey(s.Ed25519Public),
		Ed25519Private: ed25519.PrivateKey(s.Ed25519Private),
		MLDSAPublic:    mlPub,
		MLDSAPrivate:   mlPriv,
	}, nil
}
