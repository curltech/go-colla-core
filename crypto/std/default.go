package std

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/security"
	ecies "github.com/ethereum/go-ethereum/crypto/ecies"
	"golang.org/x/crypto/ed25519"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
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
func GenerateSecretKey(length int) string {
	return security.RandString(length)
}

/**
对对称密钥用对方公钥加密
*/
func WriteSecretKey(key []byte, publicKey *ecdsa.PublicKey) []byte {
	return EncryptKey(key, publicKey)
}

/**
 * 生成指定字符串的对称密码
 */
func BuildSecretKey(keyValue string) string {
	return keyValue
}

func ReadSecretKey(keyValue []byte, privateKey interface{}) []byte {
	return DecryptKey(keyValue, privateKey)
}

/**
 * 从公钥字符串表示中还原公钥
 *
 *
 * @param keyValue
 * @return
 * @throws NoSuchAlgorithmException
 * @throws InvalidKeySpecException
 */
func LoadPublicKey(keyValue []byte, byteType string) interface{} {
	var err error
	block, _ := pem.Decode(keyValue)
	public, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	publicKey := public.(*ecdsa.PublicKey)

	return publicKey
}

/**
 * 从证书中的私钥二进制字符串表示中还原私钥
 *
 *
 * @param keyValue
 * @return
 */
func LoadPrivateKey(keyValue []byte, password string, byteType string) interface{} {
	var privateKey interface{}
	var err error
	block, _ := pem.Decode(keyValue)
	if byteType == "sec1" {
		private, err := x509.ParseECPrivateKey(block.Bytes)
		if err == nil {
			return private
		}
	} else if byteType == "pkcs8" {
		bytePrivate, err := x509.DecryptPEMBlock(block, []byte(password))
		private, err := x509.ParsePKCS8PrivateKey(bytePrivate)
		if err == nil {
			privateKey, ok := private.(*ecdsa.PrivateKey)
			if ok {
				return privateKey
			} else {
				privateKey, ok := private.(*ed25519.PrivateKey)
				if ok {
					return privateKey
				}
			}
		}
	}
	if err != nil {
		panic(err)
	}
	return privateKey
}

/**
 * 生成非对称的密钥对
 *
 * @return
 */
func GenerateKeyPair(keyType string) interface{} {
	//生成密钥对
	var curveType elliptic.Curve
	switch keyType {
	case "p256":
		curveType = elliptic.P256()
	case "p384":
		curveType = elliptic.P384()
	case "p521":
		curveType = elliptic.P521()
	case "secp256k1":
		//curveType = elliptic.S256()
	}
	if curveType != nil {
		keypair, err := ecdsa.GenerateKey(curveType, rand.Reader)
		if err != nil {
			panic(err)
		}

		return keypair
	} else if keyType == "ed25519" {
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err == nil {
			return privateKey
		}
	}

	return nil
}

func GetPrivateKey(keyPair interface{}, password string, pos int) interface{} {

	return keyPair
}

func GetPublicKey(keyPair interface{}) crypto.PublicKey {
	privateKey, ok := keyPair.(*ecdsa.PrivateKey)
	if ok {
		return privateKey.Public()
	}
	privateKey_, ok := keyPair.(*ed25519.PrivateKey)
	if ok {
		return privateKey_.Public()
	}
	return nil
}

func BytePublicKey(pubkey interface{}) []byte {
	bytePublic, _ := x509.MarshalPKIXPublicKey(pubkey)
	block := &pem.Block{}
	block.Headers = make(map[string]string)
	block.Type = "ECC PUBLIC KEY"
	block.Bytes = bytePublic
	buf := new(bytes.Buffer)
	pem.Encode(buf, block)

	return buf.Bytes()
}

func BytePrivateKey(privateKey interface{}, password []byte, byteType string) []byte {
	var bytePrivate []byte
	var err error
	if byteType == "sec1" {
		priv, ok := privateKey.(*ecdsa.PrivateKey)
		if ok {
			bytePrivate, err = x509.MarshalECPrivateKey(priv)
			if err != nil {
				panic(err)
			}
		}
	} else if byteType == "pkcs8" {
		bytePrivate, err = x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			panic(err)
		}
	}

	block, err := x509.EncryptPEMBlock(rand.Reader, "ECC PRIVATE KEY", bytePrivate, password, x509.PEMCipherAES256)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	pem.Encode(buf, block)

	return buf.Bytes()
}

/**
 * 根据用户名和密钥对生成 X509 证书
 *
 *
 * @param user
 * @param keyPair
 * @return
 */
func GenerateCert(parent *x509.Certificate, notBefore time.Time, notAfter time.Time, subject *pkix.Name, password string, keyPair *ecdsa.PrivateKey) *pem.Block {
	max := new(big.Int).Lsh(big.NewInt(1), 128)   //把 1 左移 128 位，返回给 big.Int
	serialNumber, _ := rand.Int(rand.Reader, max) //返回在 [0, max) 区间均匀随机分布的一个随机值
	//subject = pkix.Name{   //Name代表一个X.509识别名。只包含识别名的公共属性，额外的属性被忽略。
	//	Organization:       []string{"Manning Publications Co."},
	//	OrganizationalUnit: []string{"Books"},
	//	CommonName:         "Go Web Programming",
	//}
	template := x509.Certificate{
		SerialNumber: serialNumber, // SerialNumber 是 CA 颁布的唯一序列号，在此使用一个大随机数来代表它
		Subject:      *subject,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		//KeyUsage 与 ExtKeyUsage 用来表明该证书是用来做服务器认证的
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, // 密钥扩展用途的序列
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	//CreateCertificate基于模板创建一个新的证书
	//第二个第三个参数相同，则证书是自签名的
	//返回的切片是DER编码的证书
	byteCert, err := x509.CreateCertificate(rand.Reader, &template, parent, keyPair.PublicKey, keyPair) //DER 格式
	if err != nil {
		panic(err)
	}
	block, err := x509.EncryptPEMBlock(rand.Reader, "CERTIFICAET", byteCert, []byte(password), x509.PEMCipherAES256)
	if err != nil {
		panic(err)
	}

	return block
}

func LoadCert(buf []byte, password string) *x509.Certificate {
	block, _ := pem.Decode(buf)
	byteCert, err := x509.DecryptPEMBlock(block, []byte(password))
	if err != nil {
		panic(err)
	}
	cert, err := x509.ParseCertificate(byteCert)
	if err != nil {
		panic(err)
	}

	return cert
}

func WritePEMFile(fileName string, block *pem.Block) {
	out, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	pem.Encode(out, block)
	out.Close()
}

func ReadPEMFile(fileName string) *pem.Block {
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	//pem解码
	block, _ := pem.Decode(buf)

	return block
}

/**
 * 非对称加密
 *
 * @param key  加密的密钥
 * @param data 待加密的明文数据
 * @return 加密后的数据
 * @throws EncryptException
 */
func Encrypt(publicKey interface{}, plaintext []byte) []byte {
	pub, ok := publicKey.(*ecdsa.PublicKey)
	if ok {
		public := ecies.ImportECDSAPublic(pub)
		ct, err := ecies.Encrypt(rand.Reader, public, plaintext, nil, nil)
		if err != nil {
			panic(err)
		}
		return ct
	}
	pub_, ok := publicKey.(*ed25519.PublicKey)
	if ok {
		logger.Errorf("NotSupport:%v", pub_)
	}

	return nil
}

/**
 * 非对称解密
 *
 * @param privkey 解密的密钥
 * @param ciphertext 已经加密的数据
 * @return 解密后的明文
 * @throws EncryptException
 */
func Decrypt(privateKey interface{}, ciphertext []byte) []byte {
	priv, ok := privateKey.(*ecdsa.PrivateKey)
	if ok {
		private := ecies.ImportECDSA(priv)
		plaintext, err := private.Decrypt(ciphertext, nil, nil)
		if err != nil {
			panic(err)
		}
		return plaintext
	}
	priv_, ok := privateKey.(*ed25519.PrivateKey)
	if ok {
		logger.Errorf("NotSupport:%v", priv_)
	}

	return nil
}

/**
 * 对称加密
 *
 * @param key  加密的密钥
	The key argument should be the AES key, either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256
 * @param data 待加密的明文数据
 * @return 加密后的数据
 * @throws EncryptException
*/
func EncryptSymmetrical(key []byte, plaintext []byte, mode string) []byte {
	if mode == "CBC" {
		return encryptCBC(plaintext, key)
	} else if mode == "CFB" {
		return encryptCFB(plaintext, key)
	} else if mode == "GCM" {
		return encryptGCM(plaintext, key)
	}
	panic("NotSupportMode")
}

/**
 * 对称解密
 *
 * @param key 解密的密钥
 * @param raw 已经加密的数据
 * @return 解密后的明文
 * @throws EncryptException
 */
func DecryptSymmetrical(key []byte, ciphertext []byte, mode string) []byte {
	if mode == "CBC" {
		return decryptCBC(ciphertext, key)
	} else if mode == "CFB" {
		return decryptCFB(ciphertext, key)
	} else if mode == "GCM" {
		return decryptGCM(ciphertext, key)
	}
	panic("NotSupportMode")
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}
func pkcs5UnPadding(plaintext []byte) []byte {
	length := len(plaintext)
	unpadding := int(plaintext[length-1])

	return plaintext[:(length - unpadding)]
}

// PKCS7加填充/和PKCS5填充一样,只是填充字段多少的区别
func pkcs7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

// PKCS7解填充/和PKCS5填充一样,只是填充字段多少的区别
func pkcs7UnPadding(encrypt []byte) []byte {
	length := len(encrypt)
	unPadding := int(encrypt[length-1])
	return encrypt[:(length - unPadding)]
}

// =================== CBC ======================
func encryptCBC(plaintext []byte, key []byte) []byte {
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	plaintext = pkcs5Padding(plaintext, blockSize)              // 补全码
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) // 加密模式
	encrypted := make([]byte, len(plaintext))                   // 创建数组
	blockMode.CryptBlocks(encrypted, plaintext)                 // 加密

	return encrypted
}
func decryptCBC(ciphertext []byte, key []byte) []byte {
	block, err := aes.NewCipher(key) // 分组秘钥
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) // 加密模式
	decrypted := make([]byte, len(ciphertext))                  // 创建数组
	blockMode.CryptBlocks(decrypted, ciphertext)                // 解密
	decrypted = pkcs5UnPadding(decrypted)                       // 去除补全码

	return decrypted
}

var gcm_nonce, _ = hex.DecodeString("48656c6c6f20476f7068657221")

func encryptGCM(plaintext []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	ciphertext := aesgcm.Seal(nil, gcm_nonce, plaintext, nil)

	return ciphertext
}

func decryptGCM(ciphertext []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	plaintext, err := aesgcm.Open(nil, gcm_nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}

	return plaintext
}

// =================== CFB ======================
func encryptCFB(plaintext []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	encrypted := make([]byte, aes.BlockSize+len(plaintext))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], plaintext)

	return encrypted
}

func decryptCFB(ciphertext []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	encrypted := ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encrypted, encrypted)

	return encrypted
}

func ValidateKey(keypair *ecdsa.PrivateKey) bool {
	secretKey := GenerateSecretKey(32)
	message := GenerateSecretKey(64)
	rtext, stext := Sign(keypair, message)
	success := Verify(&keypair.PublicKey, message, rtext, stext)
	log.Printf("Verify :b", success)

	ciphertext := Encrypt(&keypair.PublicKey, []byte(message))
	plaintext := Decrypt(keypair, ciphertext)
	success = success && message == string(plaintext)
	if message == string(plaintext) {
		log.Printf("Encrypt :b", true)
	}

	ciphertext = EncryptSymmetrical([]byte(secretKey), []byte(message), "GCM")
	plaintext = DecryptSymmetrical([]byte(secretKey), ciphertext, "GCM")
	success = success && message == string(plaintext)
	if message == string(plaintext) {
		log.Printf("GCM Encrypt :b", true)
	}

	return success
}

func Sign(privateKey *ecdsa.PrivateKey, plaintext string) ([]byte, []byte) {
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, []byte(plaintext))
	if err != nil {
		panic(err)
	}
	rtext, _ := r.MarshalText()
	stext, _ := s.MarshalText()

	return rtext, stext
}

func Verify(publicKey *ecdsa.PublicKey, plaintext string, rtext, stext []byte) bool {
	//验证数字签名
	var r, s big.Int
	r.UnmarshalText(rtext)
	s.UnmarshalText(stext)
	verify := ecdsa.Verify(publicKey, []byte(plaintext), &r, &s)

	return verify
}

/**
对对称密钥用对方公钥加密
*/
func EncryptKey(key []byte, publicKey interface{}) []byte {
	return Encrypt(publicKey, key)
}

func DecryptKey(keyValue []byte, privateKey interface{}) []byte {
	return Decrypt(privateKey, keyValue)
}

func WritePublicKey(publicKey interface{}) string {
	b := BytePublicKey(publicKey)

	return EncodeBase64(b)
}

func WritePrivateKey(privateKey interface{}, password []byte, byteType string) string {
	b := BytePrivateKey(privateKey, password, byteType)

	return EncodeBase64(b)
}

func EncodeBase64(raw []byte) string {
	return base64.StdEncoding.EncodeToString(raw)
}

func DecodeBase64(cipherText string) []byte {
	bs, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		panic(err)
	}

	return bs
}

// 16进制转换字符串
func EncodeHex(data []byte) string {
	return hex.EncodeToString(data)
}

// 16进制字符串转换成byte
func DecodeHexUpper(str string) []byte {
	bs, err := hex.DecodeString(strings.ToLower(str))
	if err != nil {
		panic(err)
	}

	return bs
}

// 获取hash加密字符串
func Hash(plaintext string, mode string) []byte {
	var hash hash.Hash
	if mode == "md5" {
		hash = md5.New()
	}
	if mode == "sha256" {
		hash = sha256.New()
	}
	if mode == "sha512" {
		hash = crypto.SHA512.New()
	}
	if mode == "sha3_256" {
		hash = crypto.SHA3_256.New()
	}
	if mode == "sha3_512" {
		hash = crypto.SHA3_512.New()
	}

	hash.Reset()
	hash.Write([]byte(plaintext))

	return hash.Sum(nil)
}

// Hmac-sha256加密
func EncodeHmac(data, key []byte, mode string) []byte {
	var hash func() hash.Hash
	if mode == "md5" {
		hash = md5.New
	}
	if mode == "sha256" {
		hash = sha256.New
	}
	if mode == "sha512" {
		hash = crypto.SHA512.New
	}
	if mode == "sha3_256" {
		hash = crypto.SHA3_256.New
	}
	if mode == "sha3_512" {
		hash = crypto.SHA3_512.New
	}
	h := hmac.New(hash, key)
	h.Write(data)

	return h.Sum(nil)
}
