# Seedr

[![GoDoc](https://godoc.org/github.com/josephbuchma/seedr?status.svg)](https://godoc.org/github.com/josephbuchma/seedr)

> WORK IN PROGRESS

Seedr, being heavily inspired by [Factory Girl](https://github.com/thoughtbot/factory_girl),
allows to easily declare factories, insert records into database (optional)
and initialize objects (structs) from those newly created records.

## Available DB drivers

  - MySQL

## Basic Usage

Let's assume we have MySQL database `testdb` with `users` table:

```sql
create table users (
    id         int(10) unsigned not null auto_increment,
    name       varchar(250) not null,
    sex        varchar(10) not null,
    age        int(5) unsigned not null auto_increment,
    created_at datetime not null default NOW(),
    primary key (id)
) engine=InnoDB default charset=utf8;
```

Basic seedr usage:

```go
package whatever

import (
  "database/sql"
  "time"

  . "github.com/josephbuchma/seedr" // dot import for convenience.
  "github.com/josephbuchma/seedr/driver/sql/mysql"
)

// Define 'users' factory

var users = Factory{
  FactoryConfig: {
    // Entity is a MySQL table name in this case
    Entity:     "users",
    PrimaryKey: "id",
  },
  // to see relations in action check out tests and docs (links below)
  Relations: {},
  // Traits are different variations of user.
  // Each field of trait must be same as name of field in `users` table.
  Traits: {
    "base": {
      // Auto means that this field will be initialized by driver (DB)
      "id":         Auto(),
      "name":       SequenceString("John-%d"), // produces John-1, John-2, John-3...
      "age":        nil,                       // nil -> NULL
      "sex":        nil,
      "created_at": time.Now(),
    },
    "old": {
      "age": 80,
    },
    "young": {
      "age": 15,
    },
    "male": {
      "sex": "male",
    },
    "female": {
      "sex": "female",
    },

    "User": {
      Include: "base",
    },
    "OldWoman": {
      Include: "base old female", // traits can be combined using include
      "name":  "Ann",             // you can override any field
    },
    "YoungMan": {
      Include: "base young man",
    },
    "User9999": {
      Include: "base",
      "id":    9999,
    },
  },
}

// Create new Seedr (usually you'll do it in separate file,
// so New() will be seedr.New()

var testdb = New("test_seedr",
  seedr.SetCreateDriver(mysql.New(openTestDB())), // configure MySQL driver (openTestDB is at the end)
  // field mapper is used for mapping trait fields to struct fields.
  // In this case seedr will look for `sql:"column_name"` tag, and if it's not
  // provided it'll fall back to SnakeFieldMapper, which converts struct field
  // name to snake case.
  seedr.SetFieldMapper(
    seedr.TagFieldMapper("sql", seedr.SnakeFieldMapper()),
  ),
).Add("users", users) // add users factory

// Create some records
func test() {
  var u User
  var users []User

  // Lines below are doing exactly what you're thinking (creating records in DB and initializing objects)
  // And yes, you can only create traits that starts with
  // capital letter (other traits are "private" to the factory,
  // and can only be included by another traits in this factory)
  testdb.Create("User").Scan(&u)
  testdb.CreateBatch("OldWoman", 10).Scan(&users)
  testdb.CreateBatch("User9999", 10).Scan(&users)
  testdb.CreateCustom("User", Trait{
    "name": "Mike",
    "age":  22,
  }).Scan(&u)
  testdb.CreateCustomBatch("User", 4, Trait{
    "name": "Mike",
    "age":  22,
  }).Scan(&u)
  
  // You can also only "build" the object, without inserting to DB
  testdb.Build("User").Scan(&u) // u.ID == 0
}

// User model
type User struct {
  Name      string    `sql:"name"`
  Age       *int      `sql:"age"`
  Sex       *string   `sql:"sex"`
  CreatedAt time.Time `sql:"created_at"`
}

func openTestDB() *sql.DB {
  db, err := sql.Open("mysql", "root:@/testdb?parseTime=true")
  if err != nil {
    panic(err)
  }
  return db
}

```

To see how to work with relations and other features, check out
[tests](https://github.com/josephbuchma/seedr/tree/master/driver/sql/mysql/internal/tests)
and [docs](https://godoc.org/github.com/josephbuchma/seedr)

