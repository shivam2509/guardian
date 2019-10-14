package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/asalih/guardian/data"
	"github.com/asalih/guardian/models"

	"golang.org/x/crypto/acme/autocert"
)

/*HTTPServer The http server handler*/
type HTTPServer struct {
	DB          *data.DBHelper
	CertManager *autocert.Manager
}

//var CertManagerHTTPHandler =

/*NewHTTPServer HTTP server initializer*/
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{&data.DBHelper{}, &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("cert-cache"),
		HostPolicy: autocert.HostWhitelist("guardsparker.com", "www.guardsparker.com"),
	}}
}

func (h HTTPServer) ServeHTTP() {

	/*srv80 := &http.Server{
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      2 * time.Minute,
		ReadTimeout:       1 * time.Minute,
		Handler:           CertManager.HTTPHandler(nil),
		Addr:              ":http",
	}*/

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		GetCertificate:     h.CertManager.GetCertificate,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_RC4_128_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
			tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		},
		PreferServerCipherSuites: true,
	}

	srv := &http.Server{
		ReadHeaderTimeout: 40 * time.Second,
		WriteTimeout:      2 * time.Minute,
		ReadTimeout:       2 * time.Minute,
		Handler:           NewGuardianHandler(false),
		Addr:              ":https",
		TLSConfig:         tlsConfig,
	}

	//go srv80.ListenAndServe()
	go http.ListenAndServe(":80", h.CertManager.HTTPHandler(nil))
	srv.ListenAndServeTLS("", "")
}

func (h HTTPServer) certificateManager() func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	var err error

	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if err != nil {
			return nil, err
		}

		fmt.Println("Incoming TLS request:" + clientHello.ServerName)
		target := h.DB.GetTarget(clientHello.ServerName)

		if target == nil {
			fmt.Println("Incoming TLS request: Target nil")
			return nil, err
		}

		if target.AutoCert {
			fmt.Println("AutoCert GetCertificate triggered.")
			leCert, lerr := h.CertManager.GetCertificate(clientHello)

			fmt.Println(leCert)
			fmt.Println(lerr)

			return leCert, lerr
		}

		if !target.CertCrt.Valid && !target.CertKey.Valid {
			return nil, errors.New("Certification is not enabled.")
		}

		cert, errl := h.loadCertificates(target)

		if errl != nil {
			panic(errl)
		}
		return &cert, nil
	}
}

func (h HTTPServer) loadCertificates(target *models.Target) (tls.Certificate, error) {
	return tls.X509KeyPair([]byte(target.CertCrt.String), []byte(target.CertKey.String))
}
