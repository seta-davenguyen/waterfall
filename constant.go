package main

const (
	NO_ERROR      = 0
	INVALID_PARAM = 1
	NO_DOC        = 2
	ER_PARSOR     = 3
	TOKEN_EXPIRED = 4
	TOKEN_INVALID = 5
	ER_INTERNAL   = 6
	NO_TOKEN      = 7
	LIMIT_MAX     = 200
)

const (
	privKeyPath = "keys/priv.pem" //openssl genrsa -out priv.pem 2048
	pubKeyPath  = "keys/pub.pem"  //openssl rsa -in priv.pem -pubout > pub.pem
)
