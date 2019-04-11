package options

// Options is the information we need about a particular sharing activity.
type Options struct {
	Directory string
	Bucket    string
	Region    string
	PubKey    string
	AwsKey    string
	Org       string
	Prefix    string
}
