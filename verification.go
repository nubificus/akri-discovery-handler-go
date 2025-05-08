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

func fetchAndConvertToPEM(sourceURL string) (string, error) {
        resp, err := http.Get(sourceURL)
        if err != nil {
                return "", fmt.Errorf("HTTP GET failed: %w", err)
        }
        defer resp.Body.Close()

        rawData, err := io.ReadAll(resp.Body)
        if err != nil {
                return "", fmt.Errorf("reading body failed: %w", err)
        }

        // Find first byte 0x30 (start of ASN.1 SEQUENCE)
        idx := bytes.IndexByte(rawData, 0x30)
        if idx == -1 {
                return "", fmt.Errorf("could not locate DER start")
        }
        derData := rawData[idx:]

        cert, err := x509.ParseCertificate(derData)
        if err != nil {
                return "", fmt.Errorf("failed to parse certificate: %w", err)
        }

        pemBytes := pem.EncodeToMemory(&pem.Block{
                Type:  "CERTIFICATE",
                Bytes: cert.Raw,
        })

        return strings.TrimSpace(string(pemBytes)), nil
}

func VerifyCertificate2(pemData string, destURL string) error {
        req, err := http.NewRequest("POST", destURL, bytes.NewBufferString(pemData))
        if err != nil {
                return fmt.Errorf("failed to create request: %w", err)
        }
        req.Header.Set("Content-Type", "text/plain")

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
                return fmt.Errorf("failed to POST PEM data: %w", err)
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                return fmt.Errorf("failed to read response body: %w", err)
        }

        fmt.Printf("Server response (%s):\n%s\n", resp.Status, string(body))

        if resp.StatusCode != http.StatusOK {
                return fmt.Errorf("non-OK HTTP status: %s", resp.Status)
        }

        return nil
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
	endpoint := fmt.Sprintf("http://%s/onboard", ip)
	/* derData, err := GetDerData(endpoint)
	if err != nil {
		return false
	}
	pemData, err := DerToPem(derData)
	if err != nil {
		return false
	}
	*/
	attestationServer := GetAttestationServer()
	if attestationServer == "" {
		return false
	}
	pemData, err := fetchAndConvertToPEM(endpoint)
        if err != nil {
                fmt.Println("Error fetching PEM:", err)
                return false
        }
	err = VerifyCertificate2(pemData, attestationServer)
	if err != nil {
		fmt.Printf("Error verifying certificate: %v\n", err)
		return false
	}
	return true
}
