package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

// IPInfo represents the relevant structure of the country aggregated.json files.
type IPInfo struct {
	CountryCode string `json:"country-code"`
	CountryName string `json:"country"`
	Subnets     struct {
		IPv4 []string `json:"ipv4"`
		IPv6 []string `json:"ipv6"`
	} `json:"subnets"`
}

// ASNInfo represents the relevant structure of the ASN aggregated.json files.
type ASNInfo struct {
	ASN         int    `json:"asn"`
	Description string `json:"description"`
	Subnets     struct {
		IPv4 []string `json:"ipv4"`
		IPv6 []string `json:"ipv6"`
	} `json:"subnets"`
}

// CLI holds the command-line arguments, parsed by Kong.
var CLI struct {
	CountryDir string `kong:"name='country-dir',default='rir-ip/country',help='Root directory for country data. If empty, this DB is not generated.'"`
	AsnDir     string `kong:"name='asn-dir',default='asn-ip/as',help='Root directory for ASN data. If empty, this DB is not generated.'"`
}

func main() {
	// kong.Parse will populate the CLI struct and handle --help automatically.
	// It will exit on parsing errors.
	kong.Parse(&CLI)

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	if CLI.CountryDir != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := generateCountryDB(CLI.CountryDir, "GeoIP2-Country.mmdb"); err != nil {
				errChan <- fmt.Errorf("country DB generation failed: %w", err)
			}
		}()
	}

	if CLI.AsnDir != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := generateAsnDB(CLI.AsnDir, "GeoIP2-ASN.mmdb"); err != nil {
				errChan <- fmt.Errorf("ASN DB generation failed: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	hasErrors := false
	for err := range errChan {
		log.Printf("Error: %v", err)
		hasErrors = true
	}

	if hasErrors {
		log.Fatal("Finished with errors.")
	} else {
		log.Println("All requested databases generated successfully.")
	}
}

func generateCountryDB(rootDir, outputFile string) error {
	log.Printf("Starting Country MMDB generation. Source: %s, Output: %s", rootDir, outputFile)

	writer, err := mmdbwriter.New(
		mmdbwriter.Options{
			DatabaseType: "GeoIP2-Country",
			RecordSize:   24, // Standard for GeoIP2-Country
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create mmdb writer: %w", err)
	}

	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "aggregated.json") {
			log.Printf("Processing country file: %s", path)
			if err := processCountryFile(path, writer); err != nil {
				log.Printf("Error processing country file %s: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path %q: %w", rootDir, err)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputFile, err)
	}
	defer file.Close()

	if _, err = writer.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write MMDB data: %w", err)
	}

	log.Printf("Successfully created %s", outputFile)
	return nil
}

// processCountryFile reads a single country aggregated.json, parses it, and inserts its networks into the writer.
func processCountryFile(filePath string, writer *mmdbwriter.Tree) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var ipInfo IPInfo
	if err := json.Unmarshal(data, &ipInfo); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if ipInfo.CountryCode == "" {
		return fmt.Errorf("country-code is empty in %s", filePath)
	}

	// This record structure is compatible with the GeoIP2 Country database format.
	record := mmdbtype.Map{
		"country": mmdbtype.Map{
			"iso_code": mmdbtype.String(ipInfo.CountryCode),
			"names": mmdbtype.Map{
				"en": mmdbtype.String(ipInfo.CountryName),
			},
		},
	}

	allSubnets := append(ipInfo.Subnets.IPv4, ipInfo.Subnets.IPv6...)
	for _, cidr := range allSubnets {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Printf("Skipping invalid CIDR '%s' in %s: %v", cidr, filePath, err)
			continue
		}
		if err := writer.Insert(network, record); err != nil {
			return fmt.Errorf("failed to insert network %s: %w", cidr, err)
		}
	}
	return nil
}

func generateAsnDB(rootDir, outputFile string) error {
	log.Printf("Starting ASN MMDB generation. Source: %s, Output: %s", rootDir, outputFile)

	writer, err := mmdbwriter.New(
		mmdbwriter.Options{
			DatabaseType: "GeoIP2-ASN",
			RecordSize:   24, // Standard for GeoIP2-ASN
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create mmdb writer: %w", err)
	}

	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "aggregated.json") {
			log.Printf("Processing ASN file: %s", path)
			if err := processAsnFile(path, writer); err != nil {
				log.Printf("Error processing ASN file %s: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path %q: %w", rootDir, err)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputFile, err)
	}
	defer file.Close()

	if _, err = writer.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write MMDB data: %w", err)
	}

	log.Printf("Successfully created %s", outputFile)
	return nil
}

// processAsnFile reads a single ASN aggregated.json, parses it, and inserts its networks into the writer.
func processAsnFile(filePath string, writer *mmdbwriter.Tree) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var asnInfo ASNInfo
	if err := json.Unmarshal(data, &asnInfo); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if asnInfo.ASN == 0 {
		return fmt.Errorf("asn is 0 or missing in %s", filePath)
	}

	record := mmdbtype.Map{
		"autonomous_system_number":       mmdbtype.Uint32(asnInfo.ASN),
		"autonomous_system_organization": mmdbtype.String(asnInfo.Description),
	}

	allSubnets := append(asnInfo.Subnets.IPv4, asnInfo.Subnets.IPv6...)
	for _, cidr := range allSubnets {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Printf("Skipping invalid CIDR '%s' in %s: %v", cidr, filePath, err)
			continue
		}
		if err := writer.Insert(network, record); err != nil {
			return fmt.Errorf("failed to insert network %s: %w", cidr, err)
		}
	}
	return nil
}
