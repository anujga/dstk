package io

import (
	"gopkg.in/errgo.v2/fmt/errors"
	"regexp"
	"strconv"
)

func ScaleSetOrdinal(name string) (int, error) {
	r, err := regexp.Compile(".*-([0-9]+)$")
	if err != nil {
		return 0, err
	}

	num := r.FindStringSubmatch(name)
	if len(num) == 0 {
		return 0, errors.Newf("Count not parse oridinal from %s", name)
	}

	return strconv.Atoi(num[1])
}
