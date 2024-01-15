
# (Work-In-Progress) RFID Backend Server for HackPGH

## Overview

The RFID Backend Server is a work in progress designed for HackPGH's magnetic lock and machine access control system. It integrates with the Wild Apricot API to fetch contact data and utilizes a local SQLite database to manage RFID tags and training information.

## Features

-   **Wild Apricot Integration:** Fetches contact data from the Wild Apricot API.
-   **SQLite Database:** Manages RFID tags and training data.
-   **Automated Synchronization:** Regularly updates the database with the latest API data.
-   **Secure HTTP Endpoints:** Provides machine and door access data via HTTPS endpoints.

## Project Structure

-   `/config`: Configuration file loading logic.
-   `/db`: Database initialization and schema management.
-   `/db/schema`: Database schema files.
-   `/db/data`: Default `tagsdb.sqlite` destination
-   `/handlers`: HTTP handlers for server endpoints.
-   `/models`: Data structures for database entities and API responses.
-   `/services`: Business logic including API and database operations.
-   `/utils`: Utility functions and singleton management.

## Getting Started

### Prerequisites

-   Go (latest stable version)
-   Access to Wild Apricot API with a valid API key
-   SSL certificate and key for HTTPS

### Installation

1.  Clone the repository to your local machine:
    
    `git clone https://github.com/your-repository/rfid-backend.git` 
    
2.  Navigate to the project directory:
    
    `cd rfid-backend` 
    
3.  Set up environment variables or a `.env` file with the Wild Apricot API key.

### Configuration

Modify the `config.yml` file in the `/config` directory to set the following parameters:

-   Database path
-   Wild Apricot account ID
-   SSL certificate and key file paths

### Running the Server

To start the server, run:

`go run main.go` 

The server will start listening for requests on port 443 and periodically update the database with data from the Wild Apricot API.

## Endpoints

-   `/`: Update Configuration web UI. Server reboot required for changes to take effect.
-   `/api/machineCache`: Returns RFID tags for a specified machine.
-   `/api/doorCache`: Returns all RFID tags for door access.

## Contributing

Contributions to improve the RFID Backend Server are welcome. Please follow the [standard pull request process](CONTRIBUTING.md) for your contributions.

## License

[MIT License](LICENSE)
