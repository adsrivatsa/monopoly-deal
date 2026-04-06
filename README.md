Generate proto with (from repo root; requires `npm install` in `src/frontend` for the TS files and `go mod tidy` in `src/backend`):

```bash
mkdir -p ./src/frontend/src/generated
protoc --plugin=./src/frontend/node_modules/.bin/protoc-gen-ts_proto --ts_proto_out=./src/frontend/src/generated --proto_path=./src ./src/schema.proto
mkdir -p ./src/backend/internal/schema
protoc -I ./src schema.proto --go_out=./src/backend/internal/schema --go_opt=paths=source_relative
```
