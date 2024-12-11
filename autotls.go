package ants

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/ipshipyard/p2p-forge/client"

	pebbleCA "github.com/letsencrypt/pebble/v2/ca"
	pebbleDB "github.com/letsencrypt/pebble/v2/db"
	pebbleVA "github.com/letsencrypt/pebble/v2/va"
	pebbleWFE "github.com/letsencrypt/pebble/v2/wfe"
)

// copied many things from https://github.com/ipshipyard/p2p-forge/blob/a588b1966f8043c3812dc0909464b6d0af185413/e2e_test.go

const (
	forge             = "libp2p.direct"
	forgeRegistration = "registration.libp2p.direct"
	authEnvVar        = client.ForgeAuthEnv
	authToken         = "testToken"
	authForgeHeader   = client.ForgeAuthHeader
)

type CertManager struct {
	CertMgr          *client.P2PForgeCertMgr
	CertLoaded       chan bool
	tmpDir           string
	instance         *caddy.Instance
	acmeHTTPListener net.Listener
}

func NewCertManager(port int) (*CertManager, error) {
	tmpDir, err := os.MkdirTemp("", "p2p-forge")
	if err != nil {
		return nil, err
	}

	if err := os.Setenv(authEnvVar, authToken); err != nil {
		return nil, err
	}

	// defer os.RemoveAll(tmpDir)

	tmpListener, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		return nil, err
	}
	httpPort := tmpListener.Addr().(*net.TCPAddr).Port
	if err := tmpListener.Close(); err != nil {
		return nil, err
	}

	dnsserver.Directives = []string{
		"log",
		"whoami",
		"startup",
		"shutdown",
		"ipparser",
		"acme",
	}

	corefile := fmt.Sprintf(`.:0 {
		log
		ipparser %s
		acme %s {
			registration-domain %s listen-address=:%d external-tls=true
			database-type badger %s
        }
	}`, forge, forge, forgeRegistration, httpPort, tmpDir)

	instance, err := caddy.Start(newInput(corefile))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	dnsServerAddress := instance.Servers()[0].LocalAddr().String()
	db := pebbleDB.NewMemoryStore()
	logger := log.New(os.Stdout, "", 0)
	ca := pebbleCA.New(logger, db, "", 0, 1, 0)
	va := pebbleVA.New(logger, 0, 0, false, dnsServerAddress, db)

	wfeImpl := pebbleWFE.New(logger, db, va, ca, false, false, 3, 5)
	muxHandler := wfeImpl.Handler()
	acmeHTTPListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	// defer acmeHTTPListener.Close()

	// Generate the self-signed certificate and private key
	certPEM, privPEM, err := generateSelfSignedCert("127.0.0.1")
	if err != nil {
		log.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	// Load the certificate and key into tls.Certificate
	cert, err := tls.X509KeyPair(certPEM, privPEM)
	if err != nil {
		log.Fatalf("Failed to load key pair: %v", err)
	}

	// Create a TLS configuration with the certificate
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Wrap the listener with TLS
	acmeHTTPListener = tls.NewListener(acmeHTTPListener, tlsConfig)

	go func() {
		http.Serve(acmeHTTPListener, muxHandler)
	}()
	cas := x509.NewCertPool()
	cas.AppendCertsFromPEM(certPEM)

	acmeEndpoint := fmt.Sprintf("https://%s%s", acmeHTTPListener.Addr(), pebbleWFE.DirectoryPath)
	certLoaded := make(chan bool, 1)
	certMgr, err := client.NewP2PForgeCertMgr(
		client.WithForgeDomain(forge), client.WithForgeRegistrationEndpoint(fmt.Sprintf("http://127.0.0.1:%d", httpPort)), client.WithCAEndpoint(acmeEndpoint), client.WithTrustedRoots(cas),
		client.WithModifiedForgeRequest(func(req *http.Request) error {
			req.Host = forgeRegistration
			req.Header.Set(authForgeHeader, authToken)
			return nil
		}),
		client.WithAllowPrivateForgeAddrs(),
		client.WithOnCertLoaded(func() {
			certLoaded <- true
		}))
	if err != nil {
		return nil, err
	}
	certMgr.Start()

	return &CertManager{
		CertMgr:          certMgr,
		CertLoaded:       certLoaded,
		tmpDir:           tmpDir,
		instance:         instance,
		acmeHTTPListener: acmeHTTPListener,
	}, nil
}

func (mgr *CertManager) Stop() error {
	defer os.RemoveAll(mgr.tmpDir)
	defer mgr.acmeHTTPListener.Close()

	errs := mgr.instance.ShutdownCallbacks()
	err := errors.Join(errs...)
	if err != nil {
		return err
	}

	if err := mgr.instance.Stop(); err != nil {
		return err
	}
	mgr.instance.Wait()
	return nil
}

func generateSelfSignedCert(ipAddr string) ([]byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"My Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // Valid for 10 years
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP(ipAddr)},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	privDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})

	return certPEM, privPEM, nil
}

// Input implements the caddy.Input interface and acts as an easy way to use a string as a Corefile.
type input struct {
	corefile []byte
}

// NewInput returns a pointer to Input, containing the corefile string as input.
func newInput(corefile string) *input {
	return &input{corefile: []byte(corefile)}
}

// Body implements the Input interface.
func (i *input) Body() []byte { return i.corefile }

// Path implements the Input interface.
func (i *input) Path() string { return "Corefile" }

// ServerType implements the Input interface.
func (i *input) ServerType() string { return "dns" }
