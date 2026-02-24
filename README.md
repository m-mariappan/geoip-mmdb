# GeoIP MMDB

This application scans a directory for `aggregated.json` files from the [`ipverse/country-ip-blocks`](https://github.com/ipverse/country-ip-blocks) or [`ipverse/as-ip-blocks`](https://github.com/ipverse/as-ip-blocks) repositories and generates MaxMind-compatible `GeoIP2-Country.mmdb` and `GeoIP2-ASN.mmdb` files.

## Data Sources

- Country data: [ipverse/country-ip-blocks](https://github.com/ipverse/country-ip-blocks)
- ASN data: [ipverse/as-ip-blocks](https://github.com/ipverse/as-ip-blocks)

## Prerequisites

- Go 1.18 or later
- The `ipverse/country-ip-blocks` and/or `ipverse/as-ip-blocks` repositories cloned or downloaded.

## How to Run

1. **Place this directory (`geoip-mmdb`) next to the `country-ip-blocks` and/or `as-ip-blocks` directories**, or adjust the path in the `--country-dir` and `--asn-dir` flags. Your directory structure should look like this:
    ```
    .
    ├── geoip-mmdb/
    │   ├── go.mod
    │   ├── main.go
    │   └── README.md
    ├── country-ip-blocks/
    │   └── country/
    │       ├── ad/
    │       │   └── aggregated.json
    │       ├── ae/
    │       │   └── aggregated.json
    │       └── ...
    └── as-ip-blocks/
        └── as/
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
    # For Country database (uses default: country-ip-blocks/country)
    go run main.go
    # Or specify custom directory
    go run main.go --country-dir ../country-ip-blocks/country

    # For ASN database (uses default: as-ip-blocks/as)
    go run main.go --asn-dir ../as-ip-blocks/as

    # Generate both databases
    go run main.go --country-dir ../country-ip-blocks/country --asn-dir ../as-ip-blocks/as
    ```
    Or build a binary and run it:
    ```sh
    go build -o geoip-mmdb .
    # For Country database
    ./geoip-mmdb --country-dir ../country-ip-blocks/country
    # For ASN database
    ./geoip-mmdb --asn-dir ../as-ip-blocks/as
    # For both databases
    ./geoip-mmdb
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