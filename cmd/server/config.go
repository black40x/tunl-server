package server

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Server struct {
	HTTPAddr    string `yaml:"http_addr"`
	HTTPPort    string `yaml:"http_port"`
	KeyFile     string `yaml:"key_file"`
	CertFile    string `yaml:"cert_file"`
	IdleTimeout int    `yaml:"idle_timeout"`
	ReadTimeout int    `yaml:"read_timeout"`
	ThreadCount int    `yaml:"thread_count"`
	KeepAlive   bool   `yaml:"keep_alive"`
	SSL         bool   `yaml:"user_ssl"`
	Debug       bool   `yaml:"debug"`
}

type Tunl struct {
	TunlAddr         string `yaml:"tunl_addr"`
	TunlPort         string `yaml:"tunl_port"`
	MaxClients       int    `yaml:"max_clients"`
	MaxTimeout       int    `yaml:"max_timeout"`
	ClientSubDomain  string `yaml:"client_subdomain"`
	ClientPublicAddr string `yaml:"client_public_addr"`
	MaxPostSize      int    `yaml:"max_post_size"`
	ClientExpireAt   int    `yaml:"client_expire_at"`
	UriPrefixSize    int    `yaml:"uri_prefix_size"`
	ServerPrivate    bool   `yaml:"server_private"`
	ServerPassword   string `yaml:"server_password"`
	AllowNames       bool   `yaml:"allow_names"`
}

type Log struct {
	Enabled  bool   `yaml:"enabled"`
	LogDir   string `yaml:"log_dir"`
	LogDaily bool   `yaml:"log_daily"`
}

type Config struct {
	Server *Server `yaml:"server"`
	Tunl   *Tunl   `yaml:"tunl"`
	Log    *Log    `yaml:"log"`
}

var configFiles = []string{
	"prod.yaml",
	"dev.yaml",
	"default.yaml",
}

func getConfigYamlFile() (string, error) {
	for _, fn := range configFiles {
		fname := fmt.Sprintf("./conf/%s", fn)
		if _, err := os.Stat(fname); err == nil {
			return fname, nil
		}
	}

	return "", errors.New("not found config files")
}

func LoadConfig() (*Config, error) {
	fn, err := getConfigYamlFile()
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	err = yaml.Unmarshal(bytes, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
