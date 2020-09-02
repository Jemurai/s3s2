package main_test

import (
	"testing"
	"bytes"
	"os"
	"golang.org/x/crypto/openpgp/packet"
	log "github.com/sirupsen/logrus"
	encrypt "github.com/tempuslabs/s3s2/encrypt"
	"github.com/stretchr/testify/assert"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/tempuslabs/s3s2/options"
	"io/ioutil"
)


// helper functions to prepare tests
func get_options() options.Options {
    return options.Options{
            Directory       : "",
            AwsKey          : "",
            Bucket          : "",
            Region          : "",
            Org             : "",
            Prefix          : "",
            PubKey          : "resources/testkey.pubkey",
            PrivKey         : "resources/testkey.privkey",
            SSMPubKey       : "",
            ScratchDirectory: "",
            ArchiveDirectory: "",
            AwsProfile      : "",
            Parallelism     : 10,
            BatchSize       : 10,
            LambdaTrigger   : false,
        }
}

func cryptFileExists(filename string) bool {
    _, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return true
}

func read_pub_key() *packet.PublicKey {
    log.Info(os.Getwd())
    sess, _ := session.NewSession()
    pub_key := encrypt.GetPubKey(sess, get_options())
    return pub_key
}

func read_priv_key() *packet.PrivateKey {
    log.Info(os.Getwd())
    sess, _ := session.NewSession()
    priv_key := encrypt.GetPrivKey(sess, get_options())
    return priv_key
}


// test that our encryption creates a new file
func TestEncryptFile(t * testing.T) {

    fn_in := "resources/test_data_file.txt"
    fn_out := "resources/test_data_file.txt.gpg"

    // ensure the fn_out is created as a part of this test
    os.RemoveAll(fn_out)

    assert := assert.New(t)

    pub_key := read_pub_key()
    encrypt.EncryptFile(pub_key, fn_in, fn_out, get_options())

    assert.True(cryptFileExists(fn_out))

}

// test that our decryption results in the same text as the input file
func TestDecryptFile(t * testing.T) {
    fn_original := "resources/test_data_file.txt"
    fn_in := "resources/test_data_file.txt.gpg"
    fn_out := "resources/test_data_file_result.txt"

    // ensure the fn_out is created as a part of this test
    os.RemoveAll(fn_out)

    pub_key := read_pub_key()
    priv_key := read_priv_key()

    encrypt.DecryptFile(pub_key, priv_key, fn_in, fn_out, get_options())

    assert := assert.New(t)

    assert.True(cryptFileExists(fn_out))

    expected, err1 := ioutil.ReadFile(fn_original)
    actual, err2 := ioutil.ReadFile(fn_out)

    if err1 != nil {
        log.Fatal(err1)
        panic(err1)
    }
    if err2 != nil {
        log.Fatal(err2)
        panic(err2)
    }

    assert.True(bytes.Equal(expected, actual))

    // cleanup
    os.Remove(fn_in)
    os.Remove(fn_out)
}