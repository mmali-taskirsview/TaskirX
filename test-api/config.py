from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    database_host: str = "postgres"
    database_port: int = 5432
    database_name: str = "taskir_adx"
    database_user: str = "taskir"
    database_password: str
    redis_host: str = "redis"
    redis_port: int = 6379
    redis_password: str
    port: int = 8000

    class Config:
        env_file = ".env"

settings = Settings()
