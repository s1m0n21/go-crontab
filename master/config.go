package master

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	API  *apiConfig  `json:"api"`
	ETCD *etcdConfig `json:"etcd"`
}

type apiConfig struct {
	ListenPort   int `json:"listen_port"`
	ReadTimeout  int `json:"read_timeout"`
	WriteTimeout int `json:"write_timeout"`
}

type etcdConfig struct {
	Endpoints   []string `json:"endpoints"`
	DialTimeout int      `json:"dial_timeout"`
}

var Config *config

func InitConfig(filename string) (err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	conf := config{}

	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}

	Config = &conf

	return
}
