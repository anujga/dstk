package core

import (
	"encoding/json"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func ParseYaml(filename string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(filename)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}

func Obj2Proto(o interface{}, m proto.Message) error {
	bs, err := json.Marshal(o)
	if err != nil {
		return err
	}

	err = protojson.Unmarshal(bs, m)
	if err != nil {
		return err
	}

	return nil
}
