
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

    session "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
	options "github.com/tempuslabs/s3s2_new/options"
	aws_helpers "github.com/tempuslabs/s3s2_new/aws_helpers"
	utils "github.com/tempuslabs/s3s2_new/utils"

	// For the signature algorithm.
	_ "golang.org/x/crypto/ripemd160"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)


// Logic to fetch the public encryption key
// depending on arguments provided, will get from SSM or provided file
func GetPubKey(sess *session.Session, opts options.Options) *packet.PublicKey {
    var in io.Reader
    var err error

    // if provided SSM Pub Key, then fetch from SSM
    if opts.SSMPubKey != "" {
        ssm_service := ssm.New(sess)
		in = strings.NewReader(aws_helpers.GetParameterValue(ssm_service, opts.SSMPubKey))

    // if provided original filepath value, then use instead
    } else if opts.PubKey != "" {
		in, err = os.Open(opts.PubKey)
		utils.LogIfError("Unable to open public key file - ", err)
    } else {
        panic("You must provide a public key argument!")
    }

	pub_key_block, err := armor.Decode(in)
	utils.LogIfError("Unable to decode public key block - ", err)

	if pub_key_block.Type != openpgp.PublicKeyType {
		log.Error("Invalid public key file")
	}

	reader := packet.NewReader(pub_key_block.Body)
	pkt, err := reader.Next()
	utils.LogIfError("Error creating reader from public key block - ", err)

	key, ok := pkt.(*packet.PublicKey)
	if !ok {
		log.Error("Invalid public key")
	}

	return key
}


func GetPrivKey(sess *session.Session, opts options.Options) *packet.PrivateKey {
    var in io.Reader
    var err error

    // if provided SSM Priv Key, then fetch from SSM
    if opts.SSMPrivKey != "" {
        ssm_service := ssm.New(sess)
		in = strings.NewReader(aws_helpers.GetParameterValue(ssm_service, opts.SSMPrivKey))

    // if provided original filepath value, then use instead
    } else if opts.PrivKey != "" {
		in, err = os.Open(opts.PrivKey)
		utils.LogIfError("Unable to open private key block - ", err)

    } else {
        panic("You must provide a private key argument!")
    }

	priv_key_block, err := armor.Decode(in)
    utils.LogIfError("Unable to decode private key block - ", err)

	if priv_key_block.Type != openpgp.PrivateKeyType {
		log.Error("Invalid private key file")
	}

	reader := packet.NewReader(priv_key_block.Body)

	pkt, err := reader.Next()
	utils.LogIfError("Unable to read private key packet - ", err)

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

func EncryptFile(pubKey *packet.PublicKey, InputFn string, OutputFn string, Opts options.Options) string {
    log.Infof("Encrypting file '%s' to '%s'...", InputFn, OutputFn)

	to := createEntityFromKeys(pubKey, nil) // We shouldn't have the receiver's private key!

	ofile, err := os.Create(OutputFn)
    utils.LogIfError("Unable to create encrypted file - ", err)
	defer ofile.Close()

	w, err := armor.Encode(ofile, "Message", make(map[string]string))
	utils.LogIfError("Unable to encode encrypted file location - ", err)
	defer w.Close()

    config := getEncryptionConfig()
	// Here the signer should be the sender
	plain, err := openpgp.Encrypt(w, []*openpgp.Entity{to}, nil, &openpgp.FileHints{IsBinary: true}, &config)
	utils.LogIfError("Unable to perform encryption - ", err)
	defer plain.Close()

	compressed, err := gzip.NewWriterLevel(plain, gzip.BestCompression)
   	utils.LogIfError("Unable to perform compression - ", err)

	infile, err := os.Open(InputFn)
	utils.LogIfError("Unable to open encrypted file location - ", err)
	defer infile.Close()

	_, err = io.Copy(compressed, infile)
	utils.LogIfError("Error writing encrypted file - ", err)
	log.Debugf("Encrypted file: '%s'", infile.Name())
	compressed.Close()

	return OutputFn
}

func DecryptFile(_pubkey *packet.PublicKey, _privkey *packet.PrivateKey, InputFn string, OutputFn string, opts options.Options) {
    log.Infof("Decrypting file '%s' to '%s'", InputFn, OutputFn)

	in, err := os.Open(InputFn)
	if err != nil {
	    log.Errorf("Unable to open decrypted file location - '%s'", InputFn)
	}
	defer in.Close()

	block, err := armor.Decode(in)
	if block.Type != "Message" {
		log.Errorf("Invalid message type")
	}

	entity := createEntityFromKeys(_pubkey, _privkey)

	var entityList openpgp.EntityList

	entityList = append(entityList, entity)

	config := getEncryptionConfig()
	md, err := openpgp.ReadMessage(block.Body, entityList, nil, &config)
	if err != nil {
		log.Errorf("Unable to read encryption - '%s'", InputFn)
    }

	compressed, err := gzip.NewReader(md.UnverifiedBody)
	if err != nil {
	    log.Errorf("Unable to open compressed encryption information location - '%s'", InputFn)
    }
	defer compressed.Close()

	dfile, err := os.Create(OutputFn)
	if err != nil {
	    log.Errorf("Unable to create encrypted file location - '%s'", OutputFn)
    }
	defer dfile.Close()

	_, err = io.Copy(dfile, compressed)
	if err != nil {
	    log.Errorf("Unable to open encrypted file location - '%s'", OutputFn)
	}

}

func encodePrivateKey(out io.Writer, key *rsa.PrivateKey) {
	w, err := armor.Encode(out, openpgp.PrivateKeyType, make(map[string]string))
	utils.LogIfError("Error executing armor.Encode for private key", err)

	pgpKey := packet.NewRSAPrivateKey(time.Now(), key)
	err = pgpKey.Serialize(w)
    utils.LogIfError("Error serializing GPG key for private key", err)

	err = w.Close()
	utils.LogIfError("Error closing public key file", err)
}

func encodePublicKey(out io.Writer, key *rsa.PrivateKey) {
	w, err := armor.Encode(out, openpgp.PublicKeyType, make(map[string]string))
	utils.LogIfError("Error executing armor.Encode for public key", err)

	pgpKey := packet.NewRSAPublicKey(time.Now(), &key.PublicKey)
	err = pgpKey.Serialize(w)
    utils.LogIfError("Error serializing GPG key for public key", err)

	err = w.Close()
	utils.LogIfError("Error closing public key file", err)
}

// GenerateKeys PGP Keys
func GenerateKeys(directory string, keyname string, bits int) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	utils.LogIfError("Error generating encryption key", err)

	priv, err := os.Create(filepath.Join(directory, keyname+".privkey"))
	utils.LogIfError("Error creating private encryption key", err)
	defer priv.Close()

	pub, err := os.Create(filepath.Join(directory, keyname+".pubkey"))
	utils.LogIfError("Error creating public encryption key", err)
	defer pub.Close()

	encodePrivateKey(priv, key)
	encodePublicKey(pub, key)
}