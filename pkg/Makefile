proto:
	rm -rf pb/*.go
	protoc --go_out=plugins=grpc+retag:. proto/*.proto

.PHONY: proto
