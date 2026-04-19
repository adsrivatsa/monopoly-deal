gen:
	mkdir -p ./src/frontend/src/generated
	protoc --plugin=./src/frontend/node_modules/.bin/protoc-gen-ts_proto --ts_proto_out=./src/frontend/src/generated --proto_path=./src/schema ./src/schema/*.proto
	rm -f ./src/backend/internal/schema/*.pb.go
	protoc -I ./src/schema ./src/schema/*.proto --go_out=./src/backend --go_opt=module=fun-kames
