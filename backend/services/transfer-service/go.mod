module github.com/zippyra/backend/services/transfer-service

go 1.25.0

replace github.com/zippyra/backend => ../..

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.12.3
	github.com/mattn/go-sqlite3 v1.14.49
	github.com/zippyra/backend v0.0.0-00010101000000-000000000000
)

require github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
