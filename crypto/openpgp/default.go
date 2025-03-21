package openpgp

import (
	"errors"
	"github.com/ProtonMail/gopenpgp/v3/constants"
	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/crypto/std"
	"time"
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
func GenerateSecretKey(size int) (sessionKey interface{}, err error) {
	if size <= 0 {
		sessionKey, err = crypto.GenerateSessionKeyAlgo(constants.AES256)
	} else {
		sessionKey, err = crypto.RandomToken(size)
	}

	return sessionKey, err
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
	if err == nil && privateKey != nil {
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
func GenerateKeyPair(keyType string, passphrase []byte, name string, email string) interface{} {
	var keypair interface{}
	var err error

	if name == "" {
		name, _ = config.GetString("server.name")
	}
	if email == "" {
		email, _ = config.GetString("server.email")
	}

	pgpDefault := crypto.PGPWithProfile(profile.Default())
	pgp4880 := crypto.PGPWithProfile(profile.RFC4880())
	pgpCryptoRefresh := crypto.PGPWithProfile(profile.RFC9580())
	keyGenHandle := pgp4880.KeyGeneration().AddUserId(name, email).New()

	if keyType == "RSA" {

		// RSA, Key struct
		keypair, err = keyGenHandle.GenerateKeyWithSecurity(constants.HighSecurity)
	} else if keyType == "Ed25519" || keyType == "x25519" {
		keyGenHandle = pgpDefault.KeyGeneration().AddUserId(name, email).New()

		keypair, err = keyGenHandle.GenerateKey()
	} else {
		keyGenHandle = pgpCryptoRefresh.KeyGeneration().AddUserId(name, email).New()

		keypair, err = keyGenHandle.GenerateKeyWithSecurity(constants.HighSecurity)
	}
	if err != nil {
		panic(err)
	}
	return keypair
}

func GetPrivateKey(keyPair *crypto.Key, password []byte) *crypto.Key {
	var err error
	if keyPair.IsPrivate() {
		if !keyPair.IsExpired(time.Now().Unix()) {
			locked, err := keyPair.IsLocked()
			if err != nil {
				panic(err)
			}
			if locked {
				_, err := keyPair.Unlock(password)
				if err != nil {
					return nil
				}
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

func BytePublicKey(key *crypto.Key) []byte {
	bs, err := key.GetPublicKey()
	if err != nil {
		panic(err)
	}

	return bs
}

func BytePrivateKey(key *crypto.Key, password []byte) []byte {
	pgp := crypto.PGP()
	if key.IsPrivate() {
		unlocked, err := key.IsUnlocked()
		if err != nil {

		} else if unlocked {
			key, err = pgp.LockKey(key, password)
			if err != nil {
				panic(err)
			}
		}
	}

	bs, err := key.Serialize()
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
func Encrypt(key *crypto.Key, plaintext []byte) (ciphertext []byte) {
	var err error
	pgp := crypto.PGP()
	encHandle, err := pgp.Encryption().Recipient(key).New()
	if err != nil {
		panic(err)
	}
	pgpMessage, err := encHandle.Encrypt(plaintext)
	ciphertext, err = pgpMessage.ArmorBytes()
	if err != nil {
		panic(err)
	}

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
func Decrypt(key *crypto.Key, ciphertext []byte) []byte {
	var err error
	pgp := crypto.PGP()
	decHandle, err := pgp.Decryption().DecryptionKey(key).New()
	if err != nil {
		panic(err)
	}
	decrypted, err := decHandle.Decrypt(ciphertext, crypto.Armor)
	if err != nil {
		panic(err)
	}

	decHandle.ClearPrivateParams()

	return decrypted.Bytes()
}

/**
 * 对称加密
 *
 * @param key  加密的密钥
 * @param data 待加密的明文数据
 * @return 加密后的数据
 * @throws EncryptException
 */
func EncryptSymmetrical(password []byte, plaintext []byte) []byte {
	var err error
	pgp := crypto.PGP()
	// Encrypt data with a password
	encHandle, err := pgp.Encryption().Password(password).New()
	if err != nil {
		panic(err)
	}
	pgpMessage, err := encHandle.Encrypt(plaintext)
	if err != nil {
		panic(err)
	}
	armored, err := pgpMessage.ArmorBytes()
	if err != nil {
		panic(err)
	}
	return armored
}

/**
 * 非对称解密
 *
 * @param key 解密的密钥
 * @param raw 已经加密的数据
 * @return 解密后的明文
 * @throws EncryptException
 */
func DecryptSymmetrical(password []byte, ciphertext []byte) []byte {
	var err error
	pgp := crypto.PGP()
	decHandle, err := pgp.Decryption().Password(password).New()
	if err != nil {
		return nil
	}
	decrypted, err := decHandle.Decrypt(ciphertext, crypto.Auto)
	if err != nil {
		panic(err)
	}
	return decrypted.Bytes()
}

func ValidateKey(keyPair *crypto.Key, password []byte) bool {
	unlockedKey := GetPrivateKey(keyPair, password)

	isVerified, _ := unlockedKey.Check()
	if !isVerified {
		return false
	}

	return true
}

func Sign(key *crypto.Key, plaintext []byte) (ciphertext []byte, err error) {
	pgp := crypto.PGP()
	signer, err := pgp.Sign().SigningKey(key).New()
	ciphertext, err = signer.Sign(plaintext, crypto.Auto)
	signer.ClearPrivateParams()
	return
}

func Verify(publicKey *crypto.Key, ciphertext []byte) (success bool, err error) {
	pgp := crypto.PGP()
	verifier, err := pgp.Verify().VerificationKey(publicKey).New()
	verifyResult, err := verifier.VerifyInline(ciphertext, crypto.Auto)
	if sigErr := verifyResult.SignatureError(); sigErr == nil {
		return true, nil
	}

	return false, err
}

func WritePublicKey(publicKey *crypto.Key) string {
	b := BytePublicKey(publicKey)

	return std.EncodeBase64(b)
}

func WritePrivateKey(privateKey *crypto.Key, password []byte) string {
	b := BytePrivateKey(privateKey, password)

	return std.EncodeBase64(b)
}
