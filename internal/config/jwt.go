package config

type JWTTokenConfig struct {
	CookieName  string `mapstructure:"cookie_name"`
	JwtSecret   string `mapstructure:"jwt_secret"`
	ExpTime     uint64 `mapstructure:"exp_time"` //expire time in min
	ValidateJwt bool   `mapstructure:"validate_jwt"`
}
