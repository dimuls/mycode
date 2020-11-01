//go:generate protoc --go_out=. runner.proto
//go:generate protoc --go_out=. --twirp_out=. api.proto
//go:generate sed -i "s/__mycode/mycode/g" api.twirp.go
package mycode
