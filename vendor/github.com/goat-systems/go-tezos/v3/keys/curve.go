package keys

import (
	"fmt"
)

type ECKind string

const (
	Ed25519   ECKind = "Ed25519"
	Secp256k1 ECKind = "Secp256k1"
	NistP256  ECKind = "NistP256"
)

type iCurve interface {
	addressPrefix() []byte
	publicKeyPrefix() []byte
	privateKeyPrefix() []byte
	signaturePrefix() []byte
	getECKind() ECKind
	getPrivateKey(v []byte) []byte
	getPublicKey(privateKey []byte) ([]byte, error)
	sign(msg []byte, privateKey []byte) (Signature, error)
	verify(v []byte, signature []byte, pubKey []byte) bool
}

func getCurve(kind ECKind) iCurve {
	if kind == Secp256k1 {
		return &secp256k1Curve{}
	} else if kind == NistP256 {
		return &nistP256Curve{}
	}

	return &ed25519Curve{}
}

func getCurveByPrefix(prefix string) (iCurve, error) {
	if prefix == "edpk" || prefix == "edsk" || prefix == "tz1" || prefix == "edesk" || prefix == "edsig" {
		return &ed25519Curve{}, nil
	}

	if prefix == "sppk" || prefix == "spsk" || prefix == "tz2" || prefix == "spesk" || prefix == "spsig" {
		return &secp256k1Curve{}, nil
	}

	if prefix == "p2pk" || prefix == "p2sk" || prefix == "tz3" || prefix == "p2esk" || prefix == "p2sig" {
		return &nistP256Curve{}, nil
	}

	return nil, fmt.Errorf("failed to find curve with prefix '%s'", prefix)
}
