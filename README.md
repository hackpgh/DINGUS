
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
#### TL:DR
This project is an RFID access control system's backend server written in Golang for the HackPGH makerspace. It uses Wild Apricot API as its source of truth for member data. 

## Features

-   **Wild Apricot Support:** Fetches [Contact data](https://app.swaggerhub.com/apis-docs/WildApricot/wild-apricot_api_for_non_administrative_access/7.15.0#/Contacts/get_accounts__accountId__contacts) from the [Wild Apricot API](https://gethelp.wildapricot.com/en/articles/182-using-wildapricot-s-api).
-   **SQLite Database:** Maintains a synchronised database of tag ids and safety training sign-offs.
-   **Automated Synchronization:** Configurable interval for updating the database with the latest WA API data. Supports WA Webhooks.
-   **Secure HTTP Endpoints:** Provides machine and door access data via HTTPS endpoints.
-   **Configuration Web UI:** Change server settings via web interface hosted at `https:/localhost/` (may require reboot).

## Project Structure

-   `/config`: Configuration file loading logic.
-   `/db`: Database initialization and schema management.
-   `/db/schema`: Database schema files.
-   `/db/data`: Default `tagsdb.sqlite` destination
-   `/handlers`: HTTP handlers for server endpoints.
-   `/models`: Data structures for database entities and API responses.
-   `/services`: Business logic including API and database operations.
-   `/utils`: Utility functions and singleton management.
-   `/web-ui`: Web assets for config update UI

## Getting Started

### Prerequisites

-   Go (latest stable version)
-   Access to Wild Apricot API with a valid API key
-   SSL certificate and key for HTTPS
-   GCC (GNU Compiler Collection) - Required for building SQLite Go package which uses cgo.

### Setting up GCC

Before you can build and run this project, make sure you have GCC installed on your system. You can typically install GCC on Linux-based systems using package managers like `apt-get` (for Debian/Ubuntu) or `yum` (for CentOS/RHEL). For macOS, you can use Homebrew. MinGW-w64 is recommended for Windows.

### Setting the `CGO_ENABLED` Environment Variable

To build and run this project successfully, you need to set the `CGO_ENABLED` environment variable to `1`.

You can set `CGO_ENABLED` temporarily in your terminal by running the following command:
#### Bash command

`export CGO_ENABLED=1` 

#### PowerShell command

`set CGO_ENABLED=1`

### Configuration

Modify the `config.yml` file in the `/config` directory to set the following parameters:

-   Database path - Default: `./db/data`
-   Wild Apricot account ID
-   SSL certificate and key file paths

### Running the Server

To start the server, run:

`go run main.go` 

The server will start listening for requests on port 443 and periodically update the database with data from the Wild Apricot API.

### Running the Unit Tests

`go test`

## Endpoints

-   `/`: Update Configuration web UI. Server reboot required for changes to take effect.
-   `/webhooks`: Wild Apricot webhooks endpoint.
-   `/registerDevice`: Process registration requests from ESP controllers on the network.

## Contributing

Contributions to improve the DINGUS project are welcome. Please follow the [standard pull request process](CONTRIBUTING.md) for your contributions.

## License

[MIT License](LICENSE)
