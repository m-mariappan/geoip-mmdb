# GeoIP MMDB

This application scans a directory for `aggregated.json` files from the [`ipverse/rir-ip`](https://github.com/ipverse/rir-ip) (for country) or [`ipverse/asn-ip`](https://github.com/ipverse/asn-ip) repositories and generates MaxMind-compatible `GeoIP2-Country.mmdb` and `GeoIP2-ASN.mmdb` files.

## Data Sources

- Country data: [ipverse/rir-ip](https://github.com/ipverse/rir-ip)
- ASN data: [ipverse/asn-ip](https://github.com/ipverse/asn-ip)

## Prerequisites

- Go 1.18 or later
- The `ipverse/rir-ip` and/or `ipverse/asn-ip` repositories cloned or downloaded.

## How to Run

1. **Place this directory (`geoip-mmdb`) next to the `rir-ip` and/or `asn-ip` directories**, or adjust the path in the `-dir` flag. Your directory structure should look like this:
    ```
    .
    ├── geoip-mmdb/
    │   ├── go.mod
    │   ├── main.go
    │   └── README.md
    ├── rir-ip/
    │   └── country/
    │       ├── ad/
    │       │   └── aggregated.json
    │       ├── ae/
    │       │   └── aggregated.json
    │       └── ...
    └── asn-ip/
        └── asn/
            ├── 13335/
            │   └── aggregated.json
            ├── 15169/
            │   └── aggregated.json
            └── ...
    ```

2. **Navigate into the `geoip-mmdb` directory and tidy dependencies:**
    ```sh
    cd geoip-mmdb
    go mod tidy
    ```

3. **Run the builder:**
    ```sh
    # For Country database
    go run main.go -dir ../rir-ip/country -output GeoIP2-Country.mmdb
    # For ASN database
    go run main.go -dir ../asn-ip/asn -output GeoIP2-ASN.mmdb
    ```
    Or build a binary and run it:
    ```sh
    go build -o geoip-mmdb .
    # For Country database
    ./geoip-mmdb -dir ../rir-ip/country -output GeoIP2-Country.mmdb
    # For ASN database
    ./geoip-mmdb -dir ../asn-ip/asn -output GeoIP2-ASN.mmdb
    ```

4. **Verify the output:**
    Files named `GeoIP2-Country.mmdb` and/or `GeoIP2-ASN.mmdb` will be created. You can use tools like `mmdblookup` to inspect their contents.
    ```sh
    # Example using mmdblookup (install with: `brew install libmaxminddb`)
    mmdblookup -f GeoIP2-Country.mmdb -i 85.94.160.1
    # Expected output for Andorra

    mmdblookup -f GeoIP2-ASN.mmdb -i 8.8.8.8
    # Expected output for Google ASN
    ```