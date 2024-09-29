# 💸 Money Transfer Service

This project implements a RESTful API for handling bulk transfers in a financial system. It provides endpoints for initiating and managing bulk transfers, with a focus on performance, scalability, and reliability.

## ✨ Features

- 🌐 RESTful API for bulk transfers
- 🔄 Retry logic for better error handling
- ⚙️ Configurable concurrency limits
- 🐳 Docker support for easy deployment
- 📊 Structured logging for better observability

## 📁 Project Structure

The project follows a clean architecture pattern, separating concerns into different layers: 

🏗️ Inspiration from [DDD Hamburger architecture](https://medium.com/@remast/the-ddd-hamburger-for-go-61dba99c4aaf) 🍔

- `cmd`: Contains the main application entry points. Note that we can implement CLI tool for the service easily. 
- `internal/api/rest`: Implements the REST API handlers and router, This is presentation layer where other servers can be implemented(eg. gRPC or graphQL)
- `internal/service`: Contains the business logic for transfer processing
- `internal/transfer and internal/account`: Domains defined using repository patter
- `config`: Handles application configuration
- `Makefile`: Provides convenient commands for building and running the application

## 🚀 Getting Started

### Prerequisites

- Go 1.22 or later
- Docker and Docker Compose

### Running the Application

1. Run tests
   ```
   make test
   ```
Note: If the test is failing due to testcontainer, please rerun again. [Known issue](https://github.com/testcontainers/testcontainers-go/issues/2172)

2. Build and run the application using Docker Compose:
   ```
   docker compose up --build (docker-compose if older version)
   ```

   This will start the API server and any required dependencies.

3. The API will be available at [http://localhost:8080/api/v1/swagger/index.html](http://localhost:8080/api/v1/swagger/index.html#/)

4. Call `/health` endpoint to check the application health.

5. You can also use [air](https://github.com/air-verse/air) for live reload during development. Just have the PostgreSQL database running and use the following command to start the application:
   ```
   air rest
   ```

## 🧠 Design Decisions and Logic

1. **🔄 Asynchronous Processing**: The bulk transfer requests are processed asynchronously to improve responsiveness and handle large volumes of transfers efficiently. This decision allows the API to quickly acknowledge receipt of the request while processing transfers in the background.

2. **🔢 Concurrency Control**: The application uses a configurable concurrency limit to prevent overwhelming downstream systems or databases. This ensures optimal performance and resource utilization.

3. **🏗️ Clean Architecture**: The project structure follows clean architecture principles, separating concerns into different layers (API, Service, Repository). This improves maintainability, testability, and allows for easier future extensions.

4. **📊 Structured Logging**: The application uses structured logging to improve observability and make it easier to track and debug issues in production environments.

5. **⚙️ Configuration Management**: A separate configuration package is used to manage application settings, allowing for easy configuration changes without modifying the code.

6. **🐳 Docker Support**: Docker and Docker Compose are used for containerization, ensuring consistent environments across development, testing, and production.

7. **🛠️ Makefile**: A Makefile is provided to simplify common development tasks and standardize build and test processes.

## 🔗 API Endpoints

- `POST /api/v1/transfers`: Initiate a new bulk transfer
- `GET /api/v1/health`: Health check endpoint

For detailed API documentation, please refer to the API specification document.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
