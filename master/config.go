package master

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	API   *apiConfig   `json:"api"`
	ETCD  *etcdConfig  `json:"etcd"`
	Web   *webConfig   `json:"web"`
	Mongo *mongoConfig `json:"mongo"`
	Log   *logConfig   `json:"log"`
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

type webConfig struct {
	Root string `json:"root"`
}

type mongoConfig struct {
	URI            string `json:"uri"`
	ConnectTimeout int    `json:"connect_timeout"`
}

type logConfig struct {
	BatchSize     int `json:"batch_size"`
	CommitTimeout int `json:"commit_timeout"`
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
