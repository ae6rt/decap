package deferrals

import (
	"os"
)

func awsCoordinates() (string, string, string) {
	key := os.Getenv("AWS_ACCESS_KEY_ID")
	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_DEFAULT_REGION")

	if key == "" || secret == "" || region == "" {
		panic("AWS key, secret, and region expected in the environment at AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_DEFAULT_REGION, respectively.")
	}

	return key, secret, region
}
