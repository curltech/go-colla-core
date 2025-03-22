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
	if _, ok := keyValue.([]byte); ok {
		bytePublicKey = keyValue.([]byte)
		publicKey, err = crypto.NewKey(bytePublicKey)
	}
	if err == nil && publicKey != nil {
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
					return nil, err
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
func GenerateKeyPair(keyType string, passphrase []byte, name string, email string) (keypair *crypto.Key, err error) {
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
		keypair, err = keyGenHandle.GenerateKeyWithSecurity(constants.HighSecurity)
	} else if keyType == "Ed25519" || keyType == "x25519" {
		keyGenHandle = pgpDefault.KeyGeneration().AddUserId(name, email).New()
		keypair, err = keyGenHandle.GenerateKey()
	} else {
		keyGenHandle = pgpCryptoRefresh.KeyGeneration().AddUserId(name, email).New()
		keypair, err = keyGenHandle.GenerateKeyWithSecurity(constants.HighSecurity)
	}
	pgp := crypto.PGP()
	keypair, err = pgp.LockKey(keypair, passphrase)

	return keypair, err
}

func GetPrivateKey(keyPair *crypto.Key, password []byte) (key *crypto.Key, err error) {
	if keyPair.IsPrivate() {
		if !keyPair.IsExpired(time.Now().Unix()) {
			locked, err := keyPair.IsLocked()
			if err != nil {
				return nil, err
			}
			if locked {
				_, err := keyPair.Unlock(password)
				if err != nil {
					return nil, err
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
		return nil, err
	}

	return keyPair, nil
}

func GetPublicKey(keyPair *crypto.Key) (key *crypto.Key, err error) {
	b, err := keyPair.GetPublicKey()
	if err != nil {
		return nil, err
	}
	key, err = crypto.NewKey(b)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func BytePublicKey(key *crypto.Key) (bs []byte, err error) {
	return key.GetPublicKey()
}

func BytePrivateKey(key *crypto.Key, password []byte) (bs []byte, err error) {
	pgp := crypto.PGP()
	if key.IsPrivate() {
		unlocked, err := key.IsUnlocked()
		if err != nil {
			return nil, err
		} else if unlocked {
			key, err = pgp.LockKey(key, password)
			if err != nil {
				return nil, err
			}
		}
	}

	return key.Serialize()
}

/**
 * 非对称加密
 *
 * @param key  加密的密钥
 * @param data 待加密的明文数据
 * @return 加密后的数据
 * @throws EncryptException
 */
func Encrypt(key *crypto.Key, plaintext []byte) (ciphertext []byte, err error) {
	pgp := crypto.PGP()
	encHandle, err := pgp.Encryption().Recipient(key).New()
	if err != nil {
		return nil, err
	}
	pgpMessage, err := encHandle.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	return pgpMessage.ArmorBytes()
}

/**
 * 非对称解密
 *
 * @param privkey 解密的密钥
 * @param ciphertext 已经加密的数据
 * @return 解密后的明文
 * @throws EncryptException
 */
func Decrypt(key *crypto.Key, ciphertext []byte) (decrypted []byte, err error) {
	pgp := crypto.PGP()
	decHandle, err := pgp.Decryption().DecryptionKey(key).New()
	if err != nil {
		return nil, err
	}
	result, err := decHandle.Decrypt(ciphertext, crypto.Armor)
	if err != nil {
		return nil, err
	}

	decHandle.ClearPrivateParams()

	return result.Bytes(), nil
}

/**
 * 对称加密
 *
 * @param key  加密的密钥
 * @param data 待加密的明文数据
 * @return 加密后的数据
 * @throws EncryptException
 */
func EncryptSymmetrical(password []byte, plaintext []byte) (armored []byte, err error) {
	pgp := crypto.PGP()
	// Encrypt data with a password
	encHandle, err := pgp.Encryption().Password(password).New()
	if err != nil {
		return nil, err
	}
	pgpMessage, err := encHandle.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	return pgpMessage.ArmorBytes()
}

/**
 * 非对称解密
 *
 * @param key 解密的密钥
 * @param raw 已经加密的数据
 * @return 解密后的明文
 * @throws EncryptException
 */
func DecryptSymmetrical(password []byte, ciphertext []byte) (decrypted []byte, err error) {
	pgp := crypto.PGP()
	decHandle, err := pgp.Decryption().Password(password).New()
	if err != nil {
		return nil, err
	}
	result, err := decHandle.Decrypt(ciphertext, crypto.Auto)
	if err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

func ValidateKey(keyPair *crypto.Key, password []byte) (result bool, err error) {
	_, err = GetPrivateKey(keyPair, password)
	if err != nil {
		return false, err
	}

	return true, nil
}

func Sign(key *crypto.Key, plaintext []byte) (signature []byte, err error) {
	pgp := crypto.PGP()
	signer, err := pgp.Sign().SigningKey(key).Detached().New()
	if err != nil {
		return nil, err
	}
	signature, err = signer.Sign(plaintext, crypto.Armor)
	signer.ClearPrivateParams()

	return
}

func Verify(publicKey *crypto.Key, data []byte, signature []byte) (success bool, err error) {
	pgp := crypto.PGP()
	verifier, err := pgp.Verify().VerificationKey(publicKey).New()
	if err != nil {
		return false, err
	}
	verifyResult, err := verifier.VerifyDetached(data, signature, crypto.Armor)
	if err != nil {
		return false, err
	}
	err = verifyResult.SignatureError()
	if err != nil {
		return false, err
	}
	return true, nil
}

func EncryptKey(key []byte, publicKey *crypto.Key) (ciphertext []byte, err error) {
	return Encrypt(publicKey, key)
}

func DecryptKey(keyValue []byte, privateKey *crypto.Key) (decrypted []byte, err error) {
	return Decrypt(privateKey, keyValue)
}

func WritePublicKey(publicKey *crypto.Key) (base64 string, err error) {
	bs, err := BytePublicKey(publicKey)
	if err != nil {
		return "", err
	}
	return std.EncodeBase64(bs), nil
}
