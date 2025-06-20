# Project Structure

## Directory Layout

```
GOLANG-ORDER-MATCHING-SYSTEM/
├── .env
├── README.md
├── test_cases.md
├── generate_struct.py
├── go.mod
├── go.sum
├── main.go
├── order_matching_system.zip
├── utils/
│   └── response.go
├── models/
│   ├── order.go
│   └── trade.go
├── scripts/
│   ├── schema.sql
│   └── setup_project.sh
├── db/
│   ├── connection.go
│   ├── order_queries.go
│   ├── trade_queries.go
│   └── utils.go
├── engine/
│   └── matcher.go
└── api/
    └── api_handler.go
```

## File Descriptions

.env
Purpose: Contains environment variables for database connection and server configuration.


README.md
Purpose: Provides project overview, setup instructions, and documentation.


test_cases.md
Purpose: Documents test cases for validating the system's functionality.


generate_struct.py
Purpose: Python script to generate the project structure in Markdown format.


go.mod
Purpose: Defines the Go module and its dependencies.


go.sum
Purpose: Ensures dependency integrity with checksums.


main.go
Purpose: Entry point of the application, initializes the server and API routes.


order_matching_system.zip
Purpose: Archived version of the project for submission.


utils/response.go
Purpose: Utility functions for handling API responses.


models/order.go
Purpose: Defines the Order struct and related methods.


models/trade.go
Purpose: Defines the Trade struct and related methods.


scripts/schema.sql
Purpose: SQL script to create the database schema and user.


scripts/setup_project.sh
Purpose: Bash script to automate project setup and server startup.


db/connection.go
Purpose: Manages database connection and initialization.


db/order_queries.go
Purpose: Contains SQL queries for order operations.


db/trade_queries.go
Purpose: Contains SQL queries for trade operations.


db/utils.go
Purpose: Utility functions for database interactions.


api/api_handler.go
Purpose: Implements API handlers for order creation, matching, and cancellation.



Notes

The structure is generated automatically. Update the 'File Descriptions' section with specific purposes for each file.
Use this file alongside README.md and test_cases.md for project documentation.

