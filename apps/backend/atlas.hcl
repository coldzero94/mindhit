env "local" {
  src = "ent://ent/schema"
  dev = "docker://postgres/16/dev?search_path=public"
  url = "postgres://postgres:password@localhost:5433/mindhit?sslmode=disable"
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
