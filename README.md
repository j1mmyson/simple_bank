# Simple-Bank in golang fiber

## How to run

1. ```
   make postgres
   ```

2. ```
   make createdb
   ```

3. ```
   make migrateup
   ```

4. ```
   make server
   ```

## Migrate 순서

`make migrateup1` - `make migrateup` - `make migrate down1` - `make migratedown` 

