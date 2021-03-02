package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

const (
	// DefaultConfigFile is the default path to the OVM exporter config
	DefaultConfigFile = "/etc/coriolis-ovm-exporter/config.toml"

	// DefaultDBFile is the default location for the DB file.
	DefaultDBFile = "/etc/coriolis-ovm-exporter/exporter.db"

	// DefaultListenPort is the default HTTPS listen port
	DefaultListenPort = 5544

	// DefaultJWTTTL is the default duration in seconds a JWT token
	// will be valid. Default 7 days.
	DefaultJWTTTL time.Duration = 168 * time.Hour
)

// ParseConfig parses the file passed in as cfgFile and returns
// a *Config object.
func ParseConfig(cfgFile string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(cfgFile, &config); err != nil {
		return nil, errors.Wrap(err, "decoding toml")
	}

	if config.DBFile == "" {
		config.DBFile = DefaultDBFile
	}

	if config.JWTAuth.TimeToLive.Duration == 0 {
		config.JWTAuth.TimeToLive.Duration = DefaultJWTTTL
	}

	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "validating config")
	}
	return &config, nil
}

// Config is the coriolis-ovm-exporter config
type Config struct {
	// DBFile is the path on disk to the database location
	DBFile string `toml:"db_file"`
	// OVMEndpoint is the API endpoint of the OVM manager.
	// We use this to authenticate client requests to the exporter.
	OVMEndpoint string `toml:"ovm_endpoint"`

	// APIServer is the api server configuration.
	APIServer APIServer `toml:"api"`

	// JWTAuth is the jwt config.
	JWTAuth JWTAuth `toml:"jwt"`
}

// Validate validates the config options
func (c *Config) Validate() error {
	return nil
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	if err != nil {
		return errors.Wrap(err, "parsing time_to_live")
	}
	return nil
}

// JWTAuth holds the jwt config.
type JWTAuth struct {
	Secret     string   `toml:"secret"`
	TimeToLive duration `toml:"time_to_live"`
}

// Validate validates the JWT config.
func (j *JWTAuth) Validate() error {
	if j.Secret == "" {
		return fmt.Errorf("missing jwt secret")
	}
	return nil
}

// APIServer holds configuration for the API server
// worker
type APIServer struct {
	Bind      string    `toml:"bind"`
	Port      int       `toml:"port"`
	TLSConfig TLSConfig `toml:"tls"`
}

// Validate validates the API server config
func (a *APIServer) Validate() error {
	if a.Port > 65535 || a.Port < 1 {
		return fmt.Errorf("invalid port nr %q", a.Port)
	}

	ip := net.ParseIP(a.Bind)
	if ip == nil {
		// No need for deeper validation here, as any invalid
		// IP address specified in this setting will raise an error
		// when we try to bind to it.
		return fmt.Errorf("invalid IP address")
	}
	if err := a.TLSConfig.Validate(); err != nil {
		return errors.Wrap(err, "validating TLS config")
	}
	return nil
}

// TLSConfig is the API server TLS config
type TLSConfig struct {
	Cert   string `toml:"certificate"`
	Key    string `toml:"key"`
	CACert string `toml:"ca_certificate"`
}

// Validate validates the TLS config
func (t *TLSConfig) Validate() error {
	if _, err := t.TLSConfig(); err != nil {
		return err
	}
	return nil
}

// TLSConfig returns a *tls.Config for the ovm exporter server
func (t *TLSConfig) TLSConfig() (*tls.Config, error) {
	caCertPEM, err := ioutil.ReadFile(t.CACert)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		return nil, fmt.Errorf("failed to parse CA cert")
	}

	cert, err := tls.LoadX509KeyPair(t.Cert, t.Key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    roots,
	}, nil
}

// Dump dumps the config to a file
func (c *Config) Dump(destination string) error {
	fd, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE, 00700)
	if err != nil {
		return err
	}

	enc := toml.NewEncoder(fd)
	if err := enc.Encode(c); err != nil {
		return err
	}
	return nil
}

// // Repo holds information about a single repository
// type Repo struct {
// 	// Name is the name of the repo. This must match the name
// 	// the repo has in OVM.
// 	Name string `toml:"name"`
// 	// FStype is the filesystem type of the repo. Only OCFS2 is
// 	// supported for now. NFS v4.2 should support reflink as well,
// 	// but it needs a newer kernel and userspace binaries to work,
// 	// as well as a backing filesystem that supports reflinks.
// 	// It seems that CIFS also supports reflinks via the
// 	// FSCTL_DUPLICATE_EXTENTS_TO_FILE ioctl, but we need to
// 	// investigate further.
// 	FStype string `toml:"filesystem"`
// 	// Location is the mount point of the repository
// 	Location string `toml:"location"`
// 	// SnapshotDir is the path relative to Location, where we save
// 	// the VM disk snapshots. If location is /mounts/repo1 then a
// 	// SnapshotDir scratch/snapshots will result in
// 	// /mounts/repo1/scratch/snapshots
// 	SnapshotDir string `toml:"snapshot_dir"`
// }
