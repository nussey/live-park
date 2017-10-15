package hue

import (
	"bytes"
	"fmt"
	"net/http"
)

const (
	userName = "qbxmE92PVSrw05PzGnqfrqey3xztQ0G6czTbjZ3W"
	address  = "http://192.168.2.2"
	light    = "2"
)

func SetRed() error {
	return callHue(254, 0)
}

func SetGreen() error {
	return callHue(254, 25500)
}

func SetBlue() error {
	return callHue(254, 46920)
}

func callHue(sat, hue int) error {
	url := address + "/api/" + userName + "/lights/" + light + "/state"
	var jsonStr = []byte(fmt.Sprintf(`{"on": true, "sat": %d, "hue": %d, "bri": 254}`, sat, hue))
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
