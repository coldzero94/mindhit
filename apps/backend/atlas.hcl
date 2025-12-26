env "local" {
  src = "ent://ent/schema"
  dev = "postgres://postgres:password@localhost:5433/mindhit_dev?sslmode=disable"
  migration {
    dir = "file://ent/migrate/migrations"
  }
}

env "prod" {
  src = "ent://ent/schema"
  migration {
    dir = "file://ent/migrate/migrations"
  }
}
