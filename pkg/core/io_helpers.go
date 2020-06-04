package core

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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

type YamlRefresher interface {
	RefreshFile(*viper.Viper) error
}

func YamlParser(filename string, watch bool, p YamlRefresher) error {
	v, err := ParseYaml(filename)
	if err != nil {
		return err
	}

	err = p.RefreshFile(v)
	if err != nil {
		return err
	}

	if watch {
		v.WatchConfig()
		v.OnConfigChange(func(in fsnotify.Event) {
			filename := v.ConfigFileUsed()
			zap.S().Infow("config file modified", "filename", filename)
			if (in.Op & (fsnotify.Write | fsnotify.Create)) == 0 {
				return
			}
			zap.S().Infow("config parsing file", "filename", filename)

			err := p.RefreshFile(v)
			if err != nil {
				zap.S().Errorw("failed to refresh config",
					"filename", filename,
					"error", err)
			}
		})
	}

	return nil
}
