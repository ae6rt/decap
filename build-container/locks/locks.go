package locks

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var client *http.Client = &http.Client{}

var Log *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

func Unlock(baseURL, buildID, buildLockKey string) error {
	url := fmt.Sprintf("%s/v2/keys/buildlocks/%?prevValue=%s", baseURL, buildLockKey, buildID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode != 201 {
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			Log.Printf("Error reading non-201 response body: %v\n", err)
			return err
		} else {
			Log.Printf("%s\n", string(data))
			return nil
		}
	}
	return nil
}
