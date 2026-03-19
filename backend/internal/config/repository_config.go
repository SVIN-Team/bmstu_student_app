package config

type RepositoryConfig struct {
	PostgresConfig `yaml:",inline"`
	RedisConfig    `yaml:",inline"`
}

type PostgresConfig struct {
	PostgresConnectionString string `yaml:"postgres_connection_string"`
	PerformOrmMigration bool `yaml:"perform_orm_migration"`
}

type RedisConfig struct {
	RedisConnectionString string `yaml:"redis_connection_string"`
}
