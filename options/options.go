package options

// Options is the information we need about a particular sharing activity.
type Options struct {
	Directory string `json:"directory"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	PubKey    string `json:"pubkey"`
	AwsKey    string `json:"awskey"`
	Org       string `json:"org"`
	Prefix    string `json:"prefix"`
}
