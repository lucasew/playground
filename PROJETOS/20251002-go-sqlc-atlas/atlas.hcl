data "composite_schema" "app" {
  schema "public" {
    url = "file://db/schema.sql"
  }
}
