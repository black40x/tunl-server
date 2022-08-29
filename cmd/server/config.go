package server

import (
	"gopkg.in/ini.v1"
)

type BaseConfig struct {
	HTTPAddr    string
	HTTPPort    string
	KeyFile     string
	CertFile    string
	IdleTimeout int
	ReadTimeout int
	ThreadCount int
	KeepAlive   bool
	SSL         bool
	Debug       bool
}

type TunlConfig struct {
	TunlAddr         string
	TunlPort         string
	MaxClients       int
	MaxTimeout       int
	ClientSubDomain  string
	ClientPublicAddr string
	MaxPostSize      int
	ClientExpireAt   int
	UriPrefixSize    int
}

type Config struct {
	Base *BaseConfig
	Tunl *TunlConfig
}

func LoadConfig() (*Config, error) {
	cfg, err := ini.LooseLoad("./conf/default.ini", "./conf/dev.ini", "./conf/prod.ini")
	if err != nil {
		return nil, err
	}

	c := &Config{}
	sect := cfg.Section("server")
	c.Base = &BaseConfig{}
	c.Base.HTTPAddr = sect.Key("http_addr").String()
	c.Base.HTTPPort = sect.Key("http_port").String()
	c.Base.CertFile = sect.Key("cert_file").String()
	c.Base.KeyFile = sect.Key("key_file").String()
	c.Base.IdleTimeout, _ = sect.Key("idle_timeout").Int()
	c.Base.ThreadCount, _ = sect.Key("thread_count").Int()
	c.Base.ReadTimeout, _ = sect.Key("read_timeout").Int()
	c.Base.KeepAlive, _ = sect.Key("keep_alive").Bool()
	c.Base.SSL, _ = sect.Key("use_ssl").Bool()
	c.Base.Debug, _ = sect.Key("debug").Bool()

	sect = cfg.Section("tunl")
	c.Tunl = &TunlConfig{}
	c.Tunl.TunlAddr = sect.Key("tunl_addr").String()
	c.Tunl.TunlPort = sect.Key("tunl_port").String()
	c.Tunl.MaxClients, _ = sect.Key("max_clients").Int()
	c.Tunl.MaxTimeout, _ = sect.Key("max_timeout").Int()
	c.Tunl.ClientSubDomain = sect.Key("client_subdomain").String()
	c.Tunl.ClientPublicAddr = sect.Key("client_public_addr").String()
	c.Tunl.MaxPostSize, _ = sect.Key("max_post_size").Int()
	c.Tunl.ClientExpireAt, _ = sect.Key("client_expire_at").Int()
	c.Tunl.UriPrefixSize, _ = sect.Key("uri_prefix_size").Int()

	return c, nil
}
