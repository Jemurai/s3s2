
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

	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
	options "github.com/tempuslabs/s3s2/options"
	aws_helpers "github.com/tempuslabs/s3s2/aws_helpers"
	utils "github.com/tempuslabs/s3s2/utils"

	// For the signature algorithm.
	_ "golang.org/x/crypto/ripemd160"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// Decrypt a file with a provided key.
func Decrypt(_pubkey *packet.PublicKey, _privkey *packet.PrivateKey, filename string, opts options.Options) {
	decryptFile(_pubkey, _privkey, filename, opts)
}

// Encrypt a file
func Encrypt(pubkey *packet.PublicKey, filename string, opts options.Options) {
	encryptFile(pubkey, filename, opts)
}

func GetPubKey(opts options.Options) *packet.PublicKey {
    var in io.Reader
    var err error

    sess := utils.GetAwsSession(opts)

    // if provided SSM Pub Key, then fetch from SSM
    if opts.SSMPubKey != "" {
        ssm_service := ssm.New(sess)
		in = strings.NewReader(aws_helpers.GetParameterValue(ssm_service, opts.SSMPubKey))

    // if provided original filepath value, then use instead
    } else if opts.PubKey != "" {
		in, err = os.Open(opts.PubKey)
        if err != nil {
            log.Error(err)
        }
    } else {
        panic("You must provide a public key argument!")
    }

	pub_key_block, err := armor.Decode(in)
	if err != nil {
		log.Error(err)
	}

	if pub_key_block.Type != openpgp.PublicKeyType {
		log.Error("Invalid public key file")
	}

	reader := packet.NewReader(pub_key_block.Body)
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


func GetPrivKey(opts options.Options) *packet.PrivateKey {
    var in io.Reader
    var err error

    sess := utils.GetAwsSession(opts)

    // if provided SSM Pub Key, then fetch from SSM
    if opts.SSMPrivKey != "" {
        ssm_service := ssm.New(sess)
		in = strings.NewReader(aws_helpers.GetParameterValue(ssm_service, opts.SSMPrivKey))

    // if provided original filepath value, then use instead
    } else if opts.PrivKey != "" {
		in, err = os.Open(opts.PrivKey)
        if err != nil {
            log.Error(err)
        }

    } else {
        panic("You must provide a public key argument!")
    }

	priv_key_block, err := armor.Decode(in)
	if err != nil {
		log.Error(err)
	}

	if priv_key_block.Type != openpgp.PrivateKeyType {
		log.Error("Invalid private key file")
	}

	reader := packet.NewReader(priv_key_block.Body)

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

func getEncryptionConfig() packet.Config {
	config := packet.Config{
		DefaultHash:            crypto.SHA256,
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionNone,
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

func encryptFile(pubKey *packet.PublicKey, file string, opts options.Options) {
    log.Debugf("Encrypting file: '%s'...", file)

	to := createEntityFromKeys(pubKey, nil) // We shouldn't have the receiver's private key!

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

    config := getEncryptionConfig()
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

func decryptFile(_pubkey *packet.PublicKey, _privkey *packet.PrivateKey, file string, opts options.Options) {

	entity := createEntityFromKeys(_pubkey, _privkey)

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