package openpgp

import (
	"errors"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/crypto/std"
)

func init() {

}

/**
 * 生成随机对称秘钥
 *
 * @param algorithm
 * @return
 * @throws NoSuchAlgorithmException
 */
func GenerateSecretKey(size int) (sessionKey interface{}) {
	var err error
	if size <= 0 {
		sessionKey, err = crypto.GenerateSessionKey()
	} else {
		sessionKey, err = crypto.RandomToken(size)
	}
	if err != nil {
		panic(err)
	}

	return sessionKey
}

/**
对对称密钥用对方公钥加密
*/
func WriteSecretKey(sessionKey *crypto.SessionKey, publicKey *crypto.Key, password []byte) []byte {
	var err error
	publicKeyRing, err := crypto.NewKeyRing(publicKey)
	if err != nil {
		panic(err)
	}
	if password == nil {
		keyPacket, err := publicKeyRing.EncryptSessionKey(sessionKey) // Will encrypt to all the keys in the keyring
		if err != nil {
			panic(err)
		}

		return keyPacket
	} else {
		keyPacketSymm, err := crypto.EncryptSessionKeyWithPassword(sessionKey, password)
		if err != nil {
			panic(err)
		}

		return keyPacketSymm
	}
}

/**
 * 生成指定字符串的对称密码
 */
func BuildSecretKey(keyValue string) string {
	return keyValue
}

func ReadSecretKey(keyPacket []byte, privateKey *crypto.Key, password []byte) *crypto.SessionKey {
	privateKeyRing, err := crypto.NewKeyRing(privateKey)
	if err != nil {
		panic(err)
	}
	if password == nil {
		decodedKeyPacket, err := privateKeyRing.DecryptSessionKey(keyPacket) // Will decode with the first valid key found
		if err != nil {
			panic(err)
		}
		return decodedKeyPacket
	} else {
		decodedSymmKeyPacket, err := crypto.DecryptSessionKeyWithPassword(keyPacket, password)
		if err != nil {
			panic(err)
		}
		return decodedSymmKeyPacket
	}
}

/**
 * 从公钥字符串,Key,KeyRing,表示中还原公钥
 *
 *
 * @param keyValue
 * @return
 * @throws NoSuchAlgorithmException
 * @throws InvalidKeySpecException
 */
func LoadPublicKey(keyValue interface{}) (*crypto.Key, error) {
	var publicKey *crypto.Key
	var bytePublicKey []byte
	var err error
	if _, ok := keyValue.(*crypto.KeyRing); ok {
		keyRing := keyValue.(*crypto.KeyRing)
		publicKey, err = keyRing.GetKey(0)
	}
	if _, ok := keyValue.(*crypto.Key); ok {
		publicKey = keyValue.(*crypto.Key)
	}
	//bytePublicKey, err = publicKey.GetPublicKey()
	if _, ok := keyValue.([]byte); ok {
		bytePublicKey = keyValue.([]byte)
		publicKey, err = crypto.NewKey(bytePublicKey)
	}
	if err == nil {
		if !publicKey.IsPrivate() {
			return publicKey, nil
		} else {
			return publicKey, errors.New("NotPublic")
		}
	}

	return nil, err
}

/**
 * 从证书中的私钥二进制字符串表示中还原私钥
 *
 *
 * @param keyValue
 * @return
 */
func LoadPrivateKey(keyValue interface{}, password string) (*crypto.Key, error) {
	var privateKey *crypto.Key
	var bytePrivateKey []byte
	var err error
	if _, ok := keyValue.(*crypto.KeyRing); ok {
		keyRing := keyValue.(*crypto.KeyRing)
		privateKey, err = keyRing.GetKey(0)
	}
	if _, ok := keyValue.(*crypto.Key); ok {
		privateKey = keyValue.(*crypto.Key)
	}
	if _, ok := keyValue.([]byte); ok {
		bytePrivateKey = keyValue.([]byte)
		privateKey, err = crypto.NewKey(bytePrivateKey)
	}
	if err == nil {
		if privateKey.IsPrivate() {
			if locked, _ := privateKey.IsLocked(); locked {
				privateKey, err = privateKey.Unlock([]byte(password))
				if err != nil {
					panic(err)
				}
			}

			return privateKey, nil
		}
	}

	return nil, err
}

/**
 * 生成非对称的密钥对
 *
 * @return
 */
func GenerateKeyPair(keyType string, passphrase []byte, armored bool, name string, email string) interface{} {
	var keypair interface{}
	var err error

	if name == "" {
		name, _ = config.GetString("server.name")
	}
	if email == "" {
		email, _ = config.GetString("server.email")
	}

	if keyType == "RSA" {
		if armored {
			// RSA, string
			keypair, err = helper.GenerateKey(name, email, passphrase, "rsa", 4096)
		} else {
			// RSA, Key struct
			keypair, err = crypto.GenerateKey(name, email, "rsa", 4096)
		}
	} else if keyType == "Ed25519" || keyType == "x25519" {
		if armored {
			// Curve25519, string
			keypair, err = helper.GenerateKey(name, email, passphrase, "x25519", 0)
		} else {
			// Curve25519, Key struct
			keypair, err = crypto.GenerateKey(name, email, "x25519", 0)
		}
	} else {
		if armored {
			// Curve25519, string
			keypair, err = helper.GenerateKey(name, email, passphrase, keyType, 0)
		} else {
			// Curve25519, Key struct
			keypair, err = crypto.GenerateKey(name, email, keyType, 0)
		}
	}
	if err != nil {
		panic(err)
	}
	return keypair
}

func GetPrivateKey(keyPair *crypto.Key, password []byte) *crypto.Key {
	var err error
	if keyPair.IsPrivate() {
		if !keyPair.IsExpired() {
			locked, err := keyPair.IsLocked()
			if err != nil {
				panic(err)
			}
			if locked {
				keyPair.Unlock(password)
				defer keyPair.ClearPrivateParams()
			}
		} else {
			err = errors.New("")
		}
	} else {
		err = errors.New("")
	}
	if err != nil {
		panic(err)
	}

	return keyPair
}

func GetPublicKey(keyPair *crypto.Key) (key *crypto.Key) {
	b, err := keyPair.GetPublicKey()
	if err != nil {
		panic(err)
	}
	key, err = crypto.NewKey(b)
	if err != nil {
		panic(err)
	}

	return key
}

func BytePublicKey(publicKey *crypto.Key) []byte {
	bs, err := publicKey.GetPublicKey()
	if err != nil {
		panic(err)
	}

	return bs
}

func BytePrivateKey(privateKey *crypto.Key, password []byte) []byte {
	if privateKey.IsPrivate() {
		unlocked, err := privateKey.IsUnlocked()
		if err != nil {

		} else if unlocked {
			privateKey, err = privateKey.Lock(password)
			if err != nil {
				panic(err)
			}
		}
	}

	bs, err := privateKey.Serialize()
	if err != nil {
		panic(err)
	}

	return bs
}

/**
 * 非对称加密
 *
 * @param key  加密的密钥
 * @param data 待加密的明文数据
 * @return 加密后的数据
 * @throws EncryptException
 */
func Encrypt(keyValue interface{}, plaintext []byte) (ciphertext []byte) {
	var err error
	if _, ok := keyValue.(*crypto.SessionKey); ok {
		sessionKey := keyValue.(*crypto.SessionKey)
		var message = crypto.NewPlainMessage(plaintext)

		// Encrypt data with session key
		ciphertext, err = sessionKey.Encrypt(message)
		if err != nil {
			panic(err)
		}

		return ciphertext
	}
	if _, ok := keyValue.([]byte); ok {
		password := keyValue.([]byte)
		var pgpMessage *crypto.PGPMessage
		var message = crypto.NewPlainMessage(plaintext)
		if pgpMessage, err = crypto.EncryptMessageWithPassword(message, password); err != nil {
			panic(err)
		}
		ciphertext := pgpMessage.GetBinary()

		return ciphertext
	}
	var publicKeyRing *crypto.KeyRing
	if _, ok := keyValue.(*crypto.Key); ok {
		publicKey := keyValue.(*crypto.Key)
		if publicKeyRing, err = crypto.NewKeyRing(publicKey); err != nil {
			panic(err)
		}
	}
	var message = crypto.NewPlainMessage(plaintext)
	var pgpMessage *crypto.PGPMessage
	if _, ok := keyValue.(*crypto.KeyRing); ok {
		publicKeyRing = keyValue.(*crypto.KeyRing)
	}
	if pgpMessage, err = publicKeyRing.Encrypt(message, nil); err != nil {
		panic(err)
	}
	ciphertext = pgpMessage.GetBinary()

	return ciphertext
}

/**
 * 非对称解密
 *
 * @param privkey 解密的密钥
 * @param ciphertext 已经加密的数据
 * @return 解密后的明文
 * @throws EncryptException
 */
func Decrypt(keyValue interface{}, passphrase []byte, ciphertext []byte) []byte {
	var pgpMessage *crypto.PGPMessage
	var message *crypto.PlainMessage
	var err error

	if _, ok := keyValue.(*crypto.SessionKey); ok {
		sessionKey := keyValue.(*crypto.SessionKey)
		// Decrypt data with session key
		decrypted, err := sessionKey.Decrypt(ciphertext)
		if err != nil {
			panic(err)
		}

		return decrypted.GetBinary()
	}

	if _, ok := keyValue.([]byte); ok {
		password := keyValue.([]byte)
		pgpMessage = crypto.NewPGPMessage(ciphertext)
		message, err = crypto.DecryptMessageWithPassword(pgpMessage, password)
		if err != nil {
			panic(err)
		}

		return message.GetBinary()
	}
	var privateKey, privateKeyUnlocked *crypto.Key
	var privateKeyRing *crypto.KeyRing
	if _, ok := keyValue.(*crypto.Key); ok {
		privateKey = keyValue.(*crypto.Key)
	}
	privateKeyUnlocked = GetPrivateKey(privateKey, passphrase)

	if privateKeyRing, err = crypto.NewKeyRing(privateKeyUnlocked); err != nil {
		panic(err)
	}
	pgpMessage = crypto.NewPGPMessage(ciphertext)
	if message, err = privateKeyRing.Decrypt(pgpMessage, nil, 0); err != nil {
		panic(err)
	}

	return message.GetBinary()
}

/**
 * 对称加密
 *
 * @param key  加密的密钥
 * @param data 待加密的明文数据
 * @return 加密后的数据
 * @throws EncryptException
 */
func EncryptSymmetrical(key []byte, plaintext []byte) []byte {
	return Encrypt(key, plaintext)
}

/**
 * 非对称解密
 *
 * @param key 解密的密钥
 * @param raw 已经加密的数据
 * @return 解密后的明文
 * @throws EncryptException
 */
func DecryptSymmetrical(key []byte, ciphertext []byte) []byte {
	return Decrypt(key, nil, ciphertext)
}

func ValidateKey(keyPair *crypto.Key, password []byte) bool {
	unlockedKey := GetPrivateKey(keyPair, password)

	isVerified, _ := unlockedKey.Check()
	if !isVerified {
		return false
	}

	return true
}

func Sign(privateKey *crypto.Key, passphrase []byte, plaintext []byte) (ciphertext []byte) {
	var message = crypto.NewPlainMessage(plaintext)
	unlockedKey := GetPrivateKey(privateKey, passphrase)
	signingKeyRing, err := crypto.NewKeyRing(unlockedKey)
	pgpSignature, err := signingKeyRing.SignDetached(message)
	if err != nil {
		panic(err)
	}

	return pgpSignature.GetBinary()
}

func Verify(publicKey *crypto.Key, plaintext []byte, ciphertext []byte) (success bool) {
	message := crypto.NewPlainMessage(plaintext)
	pgpSignature := crypto.NewPGPSignature(ciphertext)
	signingKeyRing, err := crypto.NewKeyRing(publicKey)
	err = signingKeyRing.VerifyDetached(message, pgpSignature, crypto.GetUnixTime())

	if err == nil {
		return true
	} else {
		return false
	}
}

func EncryptKey(key []byte, publicKey *crypto.Key) []byte {
	return Encrypt(publicKey, key)
}

func DecryptKey(keyValue []byte, privateKey *crypto.Key) []byte {
	return Decrypt(privateKey, nil, keyValue)
}

func WritePublicKey(publicKey *crypto.Key) string {
	b := BytePublicKey(publicKey)

	return std.EncodeBase64(b)
}

func WritePrivateKey(privateKey *crypto.Key, password []byte) string {
	b := BytePrivateKey(privateKey, password)

	return std.EncodeBase64(b)
}
