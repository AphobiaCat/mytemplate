package common

type GaGenSecretRequest struct {
	GaType     int64 `json:"gaType" validate:"required"`
	RelationID int64 `json:"relationId" validate:"required"`
}

type GaGenSecretResponse struct {
	OtpAuthURL string `json:"otpAuthUrl"`
}

type GaVerifyRequest struct {
	GaType     int    `json:"gaType" validate:"required"`
	RelationID int64  `json:"relationId" validate:"required"`
	GaCode     string `json:"gaCode" validate:"required"`
}

type GaVerifyResponse struct {
	IsVerify bool `json:"isVerify"`
}

type GaBindRequest struct {
	GaType     int    `json:"gaType" validate:"required"`
	RelationID int64  `json:"relationId" validate:"required"`
	GaCode     string `json:"gaCode" validate:"required"`
	BindType   int    `json:"bindType" validate:"oneof=0 1"`
}

type GaBindResponse struct {
	IsGaBind bool `json:"isGaBind"`
}

type DappInfoRequest struct {
	Website string `json:"website" validate:"required"`
}
