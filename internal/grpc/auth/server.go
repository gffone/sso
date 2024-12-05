package auth

import ssov1 "github.com/gffone/protos/gen/go/sso"

type serverAPI struct {
	ssov1.UnimplementedAuthServer
}
