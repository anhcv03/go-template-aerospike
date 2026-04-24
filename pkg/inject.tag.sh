for file in ../pb/*.pb.go; do
  protoc-go-inject-tag -input="$file"
done