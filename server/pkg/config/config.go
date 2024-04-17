package config

type Config struct {
	DB_URL string
}

// LoadConfig reads configuration from environment variables or configuration files.
func LoadConfig() (*Config, error) {
	return &Config{
		//DB_URL: os.Getenv("DATABASE_URL"),
		DB_URL: "postgresql://lcduels_owner:4G1VQtyAHCcT@ep-patient-sun-a57j3woq.us-east-2.aws.neon.tech/lcduels?sslmode=require",
		//pooled -> "postgresql://lcduels_owner:************@ep-patient-sun-a57j3woq-pooler.us-east-2.aws.neon.tech/lcduels?sslmode=require"
	}, nil
}
