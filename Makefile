gen:
	mkdir -p ./src/frontend/src/generated
	protoc --plugin=./src/frontend/node_modules/.bin/protoc-gen-ts_proto --ts_proto_out=./src/frontend/src/generated --proto_path=./src/schema ./src/schema/*.proto
	mkdir -p ./src/backend/internal/schema
	protoc -I ./src/schema ./src/schema/*.proto --go_out=./src/backend/internal/schema --go_opt=paths=source_relative
