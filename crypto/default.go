package crypto

import (
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	pb "github.com/libp2p/go-libp2p-core/crypto/pb"
)

type SecurityContext struct {
	Protocol               string `json:",omitempty"`
	KeyFactoryAlgorithm    string `json:",omitempty"`
	KeyStoreType           string `json:",omitempty"`
	KeyPairAlgorithm       string `json:",omitempty"`
	KeyPairType            string `json:",omitempty"`
	KeyPairLength          uint   `json:",omitempty"`
	SecretKeyAlgorithm     string `json:",omitempty"`
	SecretKeySize          uint   `json:",omitempty"`
	HashKeySize            uint   `json:",omitempty"`
	AsymmetricalAlgorithm  string `json:",omitempty"`
	SymmetricalAlgorithm   string `json:",omitempty"`
	SignatureAlgorithm     string `json:",omitempty"`
	MessageDigestAlgorithm string `json:",omitempty"`
	KeyGeneratorAlgorithm  string `json:",omitempty"`
	HmacAlgorithm          string `json:",omitempty"`
	Username               string `json:",omitempty"`
	Password               string `json:",omitempty"`
}

const (
	KeyPairType_RSA       string = "RSA"
	KeyPairType_ECDSA     string = "ECDSA"
	KeyPairType_Secp256k1 string = "Secp256k1"
	KeyPairType_Ed25519   string = "Ed25519"
)

const (
	Protocol_openpgp  string = "OpenPGP"
	Protocol_libp2p   string = "libp2p"
	Protocol_standard string = "standard"
)

func GetPbKeyType(keyPairType string) pb.KeyType {
	var pbKeyType = pb.KeyType_Ed25519
	if keyPairType == KeyPairType_Ed25519 {
		pbKeyType = pb.KeyType_Ed25519
	} else if keyPairType == KeyPairType_RSA {
		pbKeyType = pb.KeyType_RSA
	} else if keyPairType == KeyPairType_Secp256k1 {
		pbKeyType = pb.KeyType_Secp256k1
	} else if keyPairType == KeyPairType_ECDSA {
		pbKeyType = pb.KeyType_ECDSA
	}

	return pbKeyType
}

func GetLibp2pKeyType(keyPairType string) int {
	var libp2pKeyType = libp2pcrypto.Ed25519
	if keyPairType == KeyPairType_Ed25519 {
		libp2pKeyType = libp2pcrypto.Ed25519
	} else if keyPairType == KeyPairType_RSA {
		libp2pKeyType = libp2pcrypto.RSA
	} else if keyPairType == KeyPairType_Secp256k1 {
		libp2pKeyType = libp2pcrypto.Secp256k1
	} else if keyPairType == KeyPairType_ECDSA {
		libp2pKeyType = libp2pcrypto.ECDSA
	}

	return libp2pKeyType
}

func GetKeyType(keyPairType int) string {
	var openKeyType = KeyPairType_Ed25519
	if keyPairType == libp2pcrypto.Ed25519 {
		openKeyType = KeyPairType_Ed25519
	} else if keyPairType == libp2pcrypto.RSA {
		openKeyType = KeyPairType_RSA
	} else if keyPairType == libp2pcrypto.Secp256k1 {
		openKeyType = KeyPairType_Secp256k1
	} else if keyPairType == libp2pcrypto.ECDSA {
		openKeyType = KeyPairType_ECDSA
	}

	return openKeyType
}
