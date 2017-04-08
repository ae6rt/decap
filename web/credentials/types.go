package credentials

// AWSCredential encapsulates the set of Decap AWS credentials
type AWSCredential struct {
	AccessKey    string
	AccessSecret string
	Region       string
}
