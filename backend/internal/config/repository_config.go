package config

type RepositoryConfig struct {
	PostgresConfig
	RedisConfig
}

type PostgresConfig struct {
	PostgresConnectionString string `yaml:"postgres_connection_string"`
}

type RedisConfig struct {
	RedisServer    string `yaml:"redis_connection_string"`
	RedisPassword  string `yaml:"redis_connection_string"`
	DatabaseNumber int
}