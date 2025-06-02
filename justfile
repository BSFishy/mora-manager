# list out available recipes
default:
  @just --list

# run the server in development mode
dev:
  @templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."
