package common

type Eip712SignRequest struct {
	Name        string `json:"name" validate:"required"`
	ChainID     int64  `json:"chainId" validate:"required"`
	Version     string `json:"version" validate:"required"`
	From        string `json:"from"`
	GetContract bool   `json:"getContract"`
}

type EIP712SignResult struct {
	Signature         string `json:"signature"`
	R                 string `json:"r"`
	S                 string `json:"s"`
	V                 int64  `json:"v"`
	VerifyingContract string `json:"verifyingContract"`
}

type Eip712Config struct {
	PrivateKey        string `json:"PrivateKey"`
	SecretKey         string `json:"SecretKey"`
	PublicKey         string `json:"PublicKey"`
	VerifyingContract string `json:"verifyingContract"`
}

type CryptoCustodyWalletConfig struct {
	SecretKey string `json:"SecretKey"`
	Nonce     string `json:"Nonce"`
}
