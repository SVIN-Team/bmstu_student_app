package config

type RepositoryConfig struct {
	PostgresConnectionString string `yaml:"postgres_connection_string"`
	RedisConnectionString    string `yaml:"redis_connection_string"`
}