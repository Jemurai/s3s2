
package encrypt

import (
	"compress/gzip"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	options "github.com/tempuslabs/s3s2/options"
	aws_helpers "github.com/tempuslabs/s3s2/aws_helpers"

	// For the signature algorithm.
	_ "golang.org/x/crypto/ripemd160"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// Possible references...
// There are some challenges here.
//
// Specifically, we want to stream the data.
//
// https://gist.github.com/ayubmalik/a83ee23c7c700cdce2f8c5bf5f2e9f20
// https://gist.github.com/stuart-warren/93750a142d3de4e8fdd2
// https://play.golang.org/p/vk58yYArMh
// https://github.com/jchavannes/go-pgp/blob/master/pgp/encrypt.go
// https://gist.github.com/eliquious/9e96017f47d9bd43cdf9
// https://medium.com/@raul_11817/golang-cryptography-rsa-asymmetric-algorithm-e91363a2f7b3
// https://github.com/jamesruan/sodium
// https://github.com/keybase/client/blob/master/go/engine/crypto.go
// https://github.com/keybase/saltpack
// https://github.com/hashicorp/vault/blob/master/command/pgp_test.go

// Decrypt a file with a provided key.
func Decrypt(filename string, opts options.Options) {
	decryptFile(filename, opts)
}

// Encrypt a file
func Encrypt(filename string, opts options.Options) {
	encryptFile(filename, opts)
}

func getPubKey(opts options.Options) *armor.Block {
    var in io.Reader
    var err error
    // if provided SSM Pub Key, then fetch from SSM
    if opts.SSMPubKey != "" {
		in = strings.NewReader(aws_helpers.GetParameterValue(opts.SSMPubKey, opts))

    // if provided original filepath value, then use instead
    } else if opts.PubKey != ""{
		in, err = os.Open(opts.PubKey)
        if err != nil {
            log.Error(err)
        }
    } else {
        panic("You must provide a public key argument!")
    }

	block, err := armor.Decode(in)
	if err != nil {
		log.Error(err)
	}
    return block
}

func getPrivKey(opts options.Options) *armor.Block {
    var in io.Reader
    var err error
    // if provided SSM Pub Key, then fetch from SSM
    if opts.SSMPrivKey != "" {
		in = strings.NewReader(aws_helpers.GetParameterValue(opts.SSMPrivKey, opts))

    // if provided original filepath value, then use instead
    } else if opts.PrivKey != ""{
		in, err = os.Open(opts.PrivKey)
        if err != nil {
            log.Error(err)
        }
    } else {
        panic("You must provide a public key argument!")
    }

	block, err := armor.Decode(in)
	if err != nil {
		log.Error(err)
	}
    return block
}

func decodePrivateKey(opts options.Options) *packet.PrivateKey {

	block := getPrivKey(opts)

	if block.Type != openpgp.PrivateKeyType {
		log.Error("Invalid private key file")
	}

	reader := packet.NewReader(block.Body)
	pkt, err := reader.Next()
	if err != nil {
		log.Error(err)
	}

	key, ok := pkt.(*packet.PrivateKey)
	if !ok {
		log.Error("Invalid private key")
	}
	return key
}

func decodePublicKey(opts options.Options) *packet.PublicKey {

    block := getPubKey(opts)

	if block.Type != openpgp.PublicKeyType {
		log.Error("Invalid public key file")
	}

	reader := packet.NewReader(block.Body)
	pkt, err := reader.Next()
	if err != nil {
		log.Error(err)
	}

	key, ok := pkt.(*packet.PublicKey)
	if !ok {
		log.Error("Invalid public key")
	}
	return key
}

func getEncryptionConfig() packet.Config {
	config := packet.Config{
		DefaultHash:            crypto.SHA256,
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionNone, // We already zstd'd it.
		//		DefaultCompressionAlgo: packet.CompressionZLIB,
		CompressionConfig: &packet.CompressionConfig{
			Level: 9,
		},
		RSABits: 4096,
	}
	return config
}

func createEntityFromKeys(pubKey *packet.PublicKey, privKey *packet.PrivateKey) *openpgp.Entity {
	config := getEncryptionConfig()
	currentTime := config.Now()
	uid := packet.NewUserId("", "", "")

	e := openpgp.Entity{
		PrimaryKey: pubKey,
		PrivateKey: privKey,
		Identities: make(map[string]*openpgp.Identity),
	}
	isPrimaryID := false

	e.Identities[uid.Id] = &openpgp.Identity{
		Name:   uid.Name,
		UserId: uid,
		SelfSignature: &packet.Signature{
			CreationTime: currentTime,
			SigType:      packet.SigTypePositiveCert,
			PubKeyAlgo:   packet.PubKeyAlgoRSA,
			Hash:         config.Hash(),
			IsPrimaryId:  &isPrimaryID,
			FlagsValid:   true,
			FlagSign:     true,
			FlagCertify:  true,
			IssuerKeyId:  &e.PrimaryKey.KeyId,
		},
	}

	keyLifetimeSecs := uint32(86400 * 365)

	e.Subkeys = make([]openpgp.Subkey, 1)
	e.Subkeys[0] = openpgp.Subkey{
		PublicKey:  pubKey,
		PrivateKey: privKey,
		Sig: &packet.Signature{
			CreationTime:              currentTime,
			SigType:                   packet.SigTypeSubkeyBinding,
			PubKeyAlgo:                packet.PubKeyAlgoRSA,
			Hash:                      config.Hash(),
			PreferredHash:             []uint8{8}, // SHA-256
			FlagsValid:                true,
			FlagEncryptStorage:        true,
			FlagEncryptCommunications: true,
			IssuerKeyId:               &e.PrimaryKey.KeyId,
			KeyLifetimeSecs:           &keyLifetimeSecs,
		},
	}
	return &e
}

func encryptFile(file string, opts options.Options) {
    log.Debugf("Encrypting file: '%s'...", file)

	pubKey := decodePublicKey(opts)

	config := getEncryptionConfig()
	//	privKey := decodePrivateKey(privateKey, options)
	to := createEntityFromKeys(pubKey, nil) // We shouldn't have the receiver's private key!!!

	ofile, err := os.Create(file + ".gpg")
	if err != nil {
		log.Error(err)
	}
	defer ofile.Close()

	w, err := armor.Encode(ofile, "Message", make(map[string]string))
	if err != nil {
		log.Error(err)
	}
	defer w.Close()

	// Here the signer should be the sender
	plain, err := openpgp.Encrypt(w, []*openpgp.Entity{to}, nil, &openpgp.FileHints{IsBinary: true}, &config)
	if err != nil {
		log.Error(err)
	}
	defer plain.Close()

	compressed, err := gzip.NewWriterLevel(plain, gzip.BestCompression) //BestCompression)
	if err != nil {
		log.Error(err)
	}

	infile, err := os.Open(file)
	if err != nil {
		log.Error(err)
	}
	defer infile.Close()

	n, err := io.Copy(compressed, infile)
	if err != nil {
		log.Errorf("Error writing encrypted file %d", n)
	}

	compressed.Close()

	log.Infof("Encrypted file: '%s'", file)
}

func decryptFile(file string, opts options.Options) {
	pubKey := decodePublicKey(opts)
	privKey := decodePrivateKey(opts)

	entity := createEntityFromKeys(pubKey, privKey)

	in, err := os.Open(file)
	if err != nil {
		log.Error(err)
	}
	defer in.Close()

	block, err := armor.Decode(in)
	if err != nil {
		log.Error(err)
	}

	if block.Type != "Message" {
		log.Error("Invalid message type")
	}

	var entityList openpgp.EntityList
	entityList = append(entityList, entity)

	config := getEncryptionConfig()
	md, err := openpgp.ReadMessage(block.Body, entityList, nil, &config)
	if err != nil {
		log.Error(err)
	}

	compressed, err := gzip.NewReader(md.UnverifiedBody)
	if err != nil {
		log.Error(err)
	}
	defer compressed.Close()
	if err != nil {
		log.Error(err)
	}
	dfn := strings.TrimSuffix(file, ".gpg")
	dfile, err := os.Create(dfn)
	if err != nil {
		log.Error(err)
	}
	defer dfile.Close()

	n, err := io.Copy(dfile, compressed)
	if err != nil {
		log.Error(err, "Error reading encrypted file")
		log.Errorf("Decrypted %d bytes", n)
	}

}

func encodePrivateKey(out io.Writer, key *rsa.PrivateKey) {
	w, err := armor.Encode(out, openpgp.PrivateKeyType, make(map[string]string))
	if err != nil {
		log.Error(err)
	}

	pgpKey := packet.NewRSAPrivateKey(time.Now(), key)
	err = pgpKey.Serialize(w)
	if err != nil {
		log.Error(err)
	}
	err = w.Close()
	if err != nil {
		log.Error(err)
	}
}

func encodePublicKey(out io.Writer, key *rsa.PrivateKey) {
	w, err := armor.Encode(out, openpgp.PublicKeyType, make(map[string]string))
	if err != nil {
		log.Error(err)
	}

	pgpKey := packet.NewRSAPublicKey(time.Now(), &key.PublicKey)
	err = pgpKey.Serialize(w)
	if err != nil {
		log.Error(err)
	}
	err = w.Close()
	if err != nil {
		log.Error(err)
	}
}

// GenerateKeys PGP Keys
func GenerateKeys(directory string, keyname string, bits int) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Error(err)
	}

	priv, err := os.Create(filepath.Join(directory, keyname+".privkey"))
	if err != nil {
		log.Error(err)
	}
	defer priv.Close()

	pub, err := os.Create(filepath.Join(directory, keyname+".pubkey"))
	if err != nil {
		log.Error(err)
	}
	defer pub.Close()

	encodePrivateKey(priv, key)
	encodePublicKey(pub, key)
}