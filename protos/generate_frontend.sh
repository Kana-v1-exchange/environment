#!bash/sh

curl -sSL https://github.com/grpc/grpc-web/releases/download/1.3.1/protoc-gen-grpc-web-1.3.1-linux-x86_64 \
-o /usr/local/bin/protoc-gen-grpc-web

chmod +x /usr/local/bin/protoc-gen-grpc-web

protoc -I . enviroment/protos/*.proto \
--js_out=import_style=commonjs,binary:frontend/src/grpc \
--grpc-web_out=import_style=commonjs,mode=grpcwebtext:fr/src/grpc