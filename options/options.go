package options

// Options is the information we need about a particular sharing activity.
type Options struct {
	// For both encrypt/decrypt
	Region      string `json:"region"`
	Bucket      string `json:"bucket"`
	AwsProfile  string `json:"awsprofile"`
	Directory   string `json:"directory"`
	Org         string `json:"org"`
	Parallelism int    `json:"parallelism"`

	// Encrypt only
	PubKey             string   `json:"pubkey"`
	SSMPubKey          string   `json:"ssmpubkey"`
	IsGCS          	   bool   	`json:"isgcs"`
	AwsKey             string   `json:"awskey"`
	Prefix             string   `json:"prefix"`
	Hash               bool     `json:"hash"`
	ArchiveDirectory   string   `json:"archive-directory"`
	ScratchDirectory   string   `json:"scratch-directory"`
	MetaDataFiles      []string `json:"metadata-files"`
	ChunkSize          int      `json:"chunksize"`
	BatchSize          int      `json:"batchsize"`
	LambdaTrigger      bool     `json:"lambda-trigger"`
	DeleteOnCompletion bool     `json:"delete-on-completion"`

	// Decrypt only
	File        string `json:"file"`
	PrivKey     string `json:"privkey"`
	SSMPrivKey  string `json:"ssmprivkey"`
}
