package options

// Options is the information we need about a particular sharing activity.
type Options struct {
	// For both encrypt/decrypt
	Region string `json:"region"`
	Bucket string `json:"bucket"`
	AwsProfile string `json:"awsprofile"`
	Directory string `json:"directory"`
	Org       string `json:"org"`

	// Encrypt only
	PubKey    string `json:"pubkey"`
	SSMPubKey string `json:"ssmpubkey"`
	AwsKey    string `json:"awskey"`
	Prefix    string `json:"prefix"`
	Hash      bool   `json:"hash"`
	ArchiveDirectory string `json:"archive-directory"`

	// Decrypt only
	File        string `json:"file"`
	PrivKey     string `json:"privkey"`
	SSMPrivKey string `json:"ssmprivkey"`
}
