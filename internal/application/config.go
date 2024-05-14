package application

type Config struct {
	AwsSesConfig AwsSesConfig `json:"aws_ses_config"`
}

type AwsSesConfig struct {
	AwsConfig
	From string `json:"from"`
}

type AwsConfig struct {
	Region string `json:"region"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type PgConfig struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}
