package awss3

import validation "github.com/go-ozzo/ozzo-validation/v4"

type Config struct {
	AccessID  string `json:"S3_ACCESS_ID" env:"S3_ACCESS_ID"`
	SecretKey string `json:"S3_SECRET_KEY" env:"S3_SECRET_KEY"`
	Token     string `json:"S3_TOKEN" env:"S3_TOKEN"`
	Region    string `json:"S3_REGION" env:"S3_REGION"`
	Endpoint  string `json:"S3_ENDPOINT" env:"S3_ENDPOINT"`
}

func (c *Config) Validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.AccessID, validation.Required),
		validation.Field(&c.SecretKey, validation.Required),
		validation.Field(&c.Region, validation.Required),
	)
}
