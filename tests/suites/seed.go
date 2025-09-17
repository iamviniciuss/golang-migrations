package main

import (
	"fmt"

	// database "github.com/iamviniciuss/golang-migrations/src/database"
	pkg "github.com/iamviniciuss/golang-migrations/src/pkg"
	repository "github.com/iamviniciuss/golang-migrations/src/repository"
	seeds "github.com/iamviniciuss/golang-migrations/tests/seed"
)

func main() {
	fmt.Println("Up seeds...")

	database, err := pkg.MongoConnect("mongodb://localhost:27017", "test-migration01")

	if err != nil {
		panic(err)
	}

	// mysqlconn, err := repository.NewConnection()

	// if err != nil {
	// 	panic(err)
	// }

	// defer mysqlconn.Close()

	// migrationRepo := repository.NewMigrationRepositoryMySQL(mysqlconn)
	migrationRepo := repository.NewMigrationRepositoryMongo(database)
	onlineReviewRepo := seeds.NewOnlineRepositoryMongo(database)

	migrationManager := pkg.NewMigrate(migrationRepo)
	migrationManager.Register(seeds.NewAddMyIndexUser(onlineReviewRepo))
	migrationManager.Register(seeds.NewAddMyIndex(database))

	if err := migrationManager.Up(pkg.AllAvailable); err != nil {
		panic(err)
	}
}
