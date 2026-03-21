module shopkuber/catalog

go 1.22

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/cors v1.2.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	shopkuber/shared v0.0.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace shopkuber/shared => ../../shared
