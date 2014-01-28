package tls

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"vlog"
)

var (
	CertPath string
	Insecure bool
)

func init() {
	flag.StringVar(&CertPath, "tls-certs", "", "path containing anons's TLS certs and keys. If empty, $BIN/../certs and $PWD/../../../../certs are searched")
	flag.BoolVar(&Insecure, "tls-insecure", false, "ignore TLS cert verification errors")
}

var packagedCertDir string

// locate the path of the packaged PEM store which is the directory named
// "certs". functions that take a (name string) parameter look for files named
// ${name}-key.pem and/or ${name}-cert.pem in this directory.
func LocatePackagedPEMDir() (dir string, err error) {
	if packagedCertDir != "" {
		dir = packagedCertDir
		return
	} else if CertPath != "" {
		packagedCertDir = CertPath
		dir = packagedCertDir
		return
	}

	var binDir string
	binDir, err = ExecutableDir()
	if err != nil {
		return
	}

	cwd, _ := os.Getwd()

	searchList := []string{
		binDir + "../certs",           // git and deb: certs is one up from bin
		cwd + "/../../../../../certs", // tests: certs is 4 up from _test files
	}
	for _, l := range searchList {
		d := filepath.Clean(l)
		var info os.FileInfo
		if info, err = os.Stat(d); err == nil && info.IsDir() {
			packagedCertDir = d
			dir = packagedCertDir
			return
		}
	}
	err = fmt.Errorf("couldn't locate packaged PEMs in any of %v", searchList)
	return
}

// load a single PEM file (with -cert or -key suffix) from the package store
func LocatePackagedPEMFile(name string) (file string, err error) {
	if strings.IndexRune(name, '/') < 0 {
		var dir string
		if dir, err = LocatePackagedPEMDir(); err != nil {
			return
		}
		f := filepath.Join(dir, name+".pem")
		if _, err = os.Stat(f); err != nil {
			return
		}
		file = f
	} else {
		// explicit path
		if _, err = os.Stat(name); err != nil {
			return "", fmt.Errorf("Cert(s) at path %q couldn't be loaded: %s", name, err)
		}
		file = name
	}
	return
}

// load a cert/key pair from the package store
func LoadPackagedKeypair(name string) (cert tls.Certificate, certFile, keyFile string, err error) {
	certFile, err = LocatePackagedPEMFile(name + "-cert")
	if err != nil {
		return
	}
	keyFile, err = LocatePackagedPEMFile(name + "-key")
	if err != nil {
		return
	}
	cert, err = tls.LoadX509KeyPair(certFile, keyFile)
	return
}

func GenerateConfig(client bool, keyPairName string, caCertNames []string) (config *tls.Config, err error) {
	config = &tls.Config{
		InsecureSkipVerify: Insecure,
	}

	label := "server"
	if client {
		label = "client"
	}

	var keyPair tls.Certificate
	var cFile, kFile string
	if keyPairName != "" {
		keyPair, cFile, kFile, err = LoadPackagedKeypair(keyPairName)
		if err != nil {
			return
		}

		VLogf("Loaded packaged %s keypair from %s and %s", label, cFile, kFile)
		config.Certificates = []tls.Certificate{keyPair}
	}

	if len(caCertNames) > 0 {
		caPool := x509.NewCertPool()
		if client {
			config.RootCAs = caPool
		} else {
			config.ClientCAs = caPool
			if Insecure {
				config.ClientAuth = tls.RequestClientCert
			} else {
				config.ClientAuth = tls.RequireAndVerifyClientCert
			}
		}

		for _, name := range caCertNames {
			n := name
			if strings.Index(name, "/") < 0 {
				n = name + "-cert"
			}
			var file string
			if file, err = LocatePackagedPEMFile(n); err != nil {
				return nil, fmt.Errorf("Failed to find cert named %q: %s", name, err)
			}
			var b []byte
			if b, err = ioutil.ReadFile(file); err != nil {
				return nil, fmt.Errorf("Can't read cert from %s: %s", file, err)
			}

			VLogf("Allowing %s CA cert from %s", label, file)
			if ok := caPool.AppendCertsFromPEM(b); !ok {
				return nil, fmt.Errorf("No cert could be found in %s", file)
			}
		}
	}
	return
}

// returns a TLS server configuration that presents serverKeyPairName to
// clients. if clientCertNames is non-empty the server will request a client
// certificate and require that it be provided and signed by one of the named
// certs.
func ConfigureServer(serverKeyPairName string, clientCertNames ...string) (config *tls.Config, err error) {
	return GenerateConfig(false, serverKeyPairName, clientCertNames)
}

// returns a TLS client configuration that presents clientKeyPair to the remote
// server. if serverCertNames is non-empty, server certificates must be signed
// by one of the named certs; otherwise the default system CA list will be
// used.
func ConfigureClient(clientKeyPairName string, serverCertNames ...string) (config *tls.Config, err error) {
	return GenerateConfig(true, clientKeyPairName, serverCertNames)
}
