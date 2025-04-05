package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetDerData(endpoint string) ([]byte, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status code: %d", resp.StatusCode)
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func DerToPem(derBytes []byte) ([]byte, error) {
	// Parse DER bytes
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, err
	}

	// Create PEM block
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	// Encode to PEM format
	return pem.EncodeToMemory(pemBlock), nil
}

func VerifyCertificate(data []byte, endpoint string) error {
	postResp, err := http.Post(
		endpoint,
		"text/plain",
		bytes.NewReader(data),
	)

	if err != nil {
		return err
	}
	defer postResp.Body.Close()

	_, _ = io.ReadAll(postResp.Body)
	if postResp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", postResp.StatusCode)
	}
	return nil
}

func GetAttestationServer() string {
	tmp := os.Getenv("DICE_AUTH_SERVICE_PORT")
	return strings.ReplaceAll(tmp, "tcp", "http")
}

func VerifyDevice(ip string) bool {
	endpoint := fmt.Sprintf("http://%s/oboard", ip)
	derData, err := GetDerData(endpoint)
	if err != nil {
		return false
	}
	pemData, err := DerToPem(derData)
	if err != nil {
		return false
	}
	attestationServer := GetAttestationServer()
	if attestationServer == "" {
		return false
	}
	err = VerifyCertificate(pemData, attestationServer)
	return err == nil
}
