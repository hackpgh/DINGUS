
# (Work-In-Progress) Deployable Integrated Network-Generalized User Selector - DINGUS for HackPGH

## Overview

In a makerspace like HackPGH where diverse paths cross, each member carries a unique story. The DINGUS project resonates with this notion, a silent observer at the crossroads of the maker spirit and collaboration.

Hesse writes, "The river is everywhere at the same time, at the source and at the mouth... in the ocean and in the mountains." Similarly, each tag's journey through this access control system is unique, yet part of a greater flow.

This project, therefore, is more than a functional necessity; it is a symbol of curiosity and, importantly, trust. In the simple act of swiping a tag, members not only access a physical space but also reaffirm their place in a collaborative history. This microcosm of an event cystallizes the trust we place in our fellow makers to create with passion, care, and mutual respect.

In this confluence of technology and humanity, this server becomes a 'river running everywhere' connecting our stories, echoing the interconnectedness Hesse so eloquently depicted, and reminding us that while our journeys are our own, they are made richer for our intersections with the paths of others.
```
      |\      _,,,---,,_     "Uh-huh..."
ZZZzz /,`.-'`'    -.  ;-;;,_
     |,4-  ) )-,_. ,\ (  `'-'
    '---''(_/--'  `-'\_)  
```
### TL:DR
This project is an RFID access control system's backend server written in Golang for the HackPGH makerspace. It uses Wild Apricot API as its source of truth for member data. 

## Features

-   **Wild Apricot Integration**: Synchronizes member [contact data](https://app.swaggerhub.com/apis-docs/WildApricot/wild-apricot_api_for_non_administrative_access/7.15.0#/Contacts/get_accounts__accountId__contacts) from the [Wild Apricot API](https://gethelp.wildapricot.com/en/articles/182-using-wildapricot-s-api).
-   **Distributed RFID Access Control**: Synchronizes authorization data caches for Wiegand26 RFID tag readers.
-   **SSO OAuth2 Authentication**: Implements Wild Apricot [SSO OAuth2](https://gethelp.wildapricot.com/en/articles/200-single-sign-on-service-sso#overview) for secure access to web-based interfaces.
-   **SQLite Database**: Maintains persistent data, including Wild Apricot Contact IDs, RFID tags and safety training records.
-   **Automated Data Sync**: Regular updates from the Wild Apricot API as well as real-time Contact and Membership webhook support.
-   **Secure Web UI**: Web interface for configuration and device management, secured via HTTPS.

## Web UI Screens

1.  **Configuration Screen**: Modify server settings, effective upon reboot.
2.  **Device Management**: Monitor and manage RFID devices.
3.  **User Authentication**: Secured with Wild Apricot SSO OAuth2, restricting access to authorized users.

## Project Structure

-   `auth`: Authentication logic, including OAuth2 SSO.
-   `config`: Configuration file parsing and loading.
-   `db`: Database initialization and schema management.
-   `handlers`: HTTP server endpoint handlers.
-   `models`: Database and API response structures.
-   `services`: Business logic for API and database interactions.
-   `setup`: Server and component initialization.
-   `utils`: General utility functions.
-   `webhooks`: Wild Apricot webhook handling.
-   `web-ui`: Frontend assets.

## Getting Started

### Prerequisites

-   [Go](https://go.dev/doc/install) (latest stable version)
-   Access to [Wild Apricot API](https://gethelp.wildapricot.com/en/articles/182-using-wildapricot-s-api)
-   SSL certificate and key
-   GCC for SQLite Go package compilation (requires cgo)

### Setting CGO_ENABLED

To successfully build and run this project, `CGO_ENABLED` must be set to `1`. This allows for the compilation of C code, a requirement for the SQLite package used in the project.

-   **Bash**: `export CGO_ENABLED=1`
-   **PowerShell**: `set CGO_ENABLED=1`

## Generating SSL Certificates for HTTPS

### Prerequisites

-   OpenSSL installed on your system. For Windows, you can download it from [here](https://indy.fulgan.com/SSL/). Most Linux distributions and MacOS have it pre-installed.

### Instructions

#### For Bash (Linux/MacOS)

1.  Open a terminal.
2.  Run the following command to generate a private key and a certificate:

    `openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365` 
    
3.  You will be prompted to enter details for the certificate and a passphrase for the private key.
4.  Remove the passphrase from the private key (optional):
    
    `openssl rsa -in key.pem -out key.unencrypted.pem`
    
    `mv key.unencrypted.pem key.pem` 
    

#### For Windows

1.  [Download](https://indy.fulgan.com/SSL/) and install OpenSSL from the provided link.
2.  Open OpenSSL via the command prompt (you might need to navigate to the OpenSSL `bin` directory).
3.  Run the following command:

    `openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365` 
    
4.  Enter the required information when prompted, and set a passphrase for the private key.
5.  Remove the passphrase from the private key (optional):

    `openssl rsa -in key.pem -out key.unencrypted.pem`
    
    `move key.unencrypted.pem key.pem` 
    

After generating the `key.pem` and `cert.pem` files, place them in the appropriate directory as specified in your `config.yaml` file for the server to use them for HTTPS.

### Configuring the Server

Update the `config.yaml` file to point to the location of your newly generated `cert.pem` and `key.pem` files under the SSL certificate and key file paths, respectively.

### Running the Server

You can start the server with `go run main.go`

### Accessing Swagger Documentation

Swagger documentation can be accessed by navigating to `https://localhost/swagger/index.html` once the server is running. This provides an interactive UI to explore and test the available API endpoints.

### Running Unit Tests

Run `go test` to execute the unit tests.

## Endpoints

-   `/`: Update Configuration web UI. Server reboot required for changes to take effect.
-   `/webhooks`: Wild Apricot webhooks endpoint.
-   `/registerDevice`: DEPRECATED - Process registration requests from ESP controllers on the network.

## Contributing

Contributions to improve the DINGUS project are welcome. Please follow the [standard pull request process](CONTRIBUTING.md) for your contributions.

## License

[MIT License](LICENSE)
