package io

import "io/ioutil"

func Cat(file string) (string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	text := string(content)
	return text, nil
}
