mysql: 
	docker run --name mysql5.7 -p 3306:3306 -e MYSQL_USER=user -e MYSQL_PASSWORD=password -e MYSQL_ROOT_PASSWORD=password -e MYSQL_DATABASE=guestlist_db -d mysql:5.7

createdb:
	docker exec -i mysql5.7 mysql -ppassword <<< "CREATE DATABASE guestlist_db;"

dropdb:
	docker exec -i mysql5.7 mysql -ppassword <<< "DROP DATABASE guestlist_db;"

migrateup:
	migrate -path db/migration -database "mysql://user:password@tcp(localhost:3306)/guestlist_db?parseTime=true" -verbose up

migratedown:
	migrate -path db/migration -database "mysql://user:password@tcp(localhost:3306)/guestlist_db?parseTime=true" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/ellisp97/BE_Task_Oct20/golang/db/sqlc Store

swagger:
	swag init --parseDependency

.PHONY: mysql createdb dropdb migrateup migratedown sqlc test server mock swagger

