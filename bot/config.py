import os
from dotenv import load_dotenv

load_dotenv()


class Config:
    def __init__(self):
        self.bot_token = os.getenv("BOT_TOKEN")
        self.database_url = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/tgbot")
        self.redis_url = os.getenv("REDIS_URL", "redis://localhost:6379")
        self.admin_user_ids = [int(x) for x in os.getenv("ADMIN_USER_IDS", "").split(",") if x]

    @property
    def redis_host(self) -> str:
        """Parse Redis URL to get host"""
        url = self.redis_url.replace("redis://", "")
        if "@" in url:
            # Handle auth: redis://:password@host:port
            parts = url.split("@")
            return parts[1].split(":")[0]
        return url.split(":")[0]

    @property
    def redis_port(self) -> int:
        """Parse Redis URL to get port"""
        url = self.redis_url.replace("redis://", "")
        if "@" in url:
            parts = url.split("@")[1].split(":")
        else:
            parts = url.split(":")
        return int(parts[1]) if len(parts) > 1 else 6379

    @property
    def redis_password(self) -> str:
        """Parse Redis URL to get password"""
        if "@" in self.redis_url:
            auth_part = self.redis_url.split("@")[0].replace("redis://", "")
            if auth_part.startswith(":"):
                return auth_part[1:]
        return None
