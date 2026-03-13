import asyncio
import logging
import os
import signal
import sys
from datetime import datetime

import psutil
from telegram.ext import ApplicationBuilder
from config import Config
from database import Database
from redis_client import RedisClient
from handlers import setup_handlers

# Configure logging
logging.basicConfig(
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    level=logging.INFO,
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler('bot.log')
    ]
)
logger = logging.getLogger(__name__)


class BotEngine:
    def __init__(self):
        self.config = Config()
        self.db = None
        self.redis = None
        self.application = None
        self.start_time = datetime.now()
        self.running = False

    async def initialize(self):
        """Initialize database, Redis, and Telegram application"""
        # Initialize database
        self.db = Database(self.config.database_url)
        await self.db.connect()
        logger.info("Database connected")

        # Initialize Redis
        self.redis = RedisClient(self.config.redis_url)
        await self.redis.connect()
        logger.info("Redis connected")

        # Initialize Telegram application
        self.application = ApplicationBuilder().token(self.config.bot_token).build()

        # Setup handlers
        setup_handlers(self.application, self.db, self.redis)
        logger.info("Handlers registered")

        # Setup Redis command listener
        asyncio.create_task(self.listen_for_commands())

        # Start metrics reporter
        asyncio.create_task(self.report_metrics())

    async def listen_for_commands(self):
        """Listen for commands from API via Redis"""
        pubsub = self.redis.pubsub()
        await pubsub.subscribe("bot:command")

        async for message in pubsub.listen():
            if message["type"] == "message":
                command = message["data"].decode()
                logger.info(f"Received command: {command}")
                await self.handle_command(command)

    async def handle_command(self, command: str):
        """Handle commands from API"""
        if command.startswith("reload_plugin:"):
            plugin_id = command.split(":")[1]
            # Plugin reload logic would go here
            logger.info(f"Reload plugin requested: {plugin_id}")

        elif command.startswith("sync_group:"):
            chat_id = int(command.split(":")[1])
            # Sync group info from Telegram
            logger.info(f"Sync group requested: {chat_id}")

    async def report_metrics(self):
        """Periodically report bot metrics to Redis"""
        while self.running:
            try:
                process = psutil.Process()
                memory_mb = process.memory_info().rss / 1024 / 1024
                cpu_percent = process.cpu_percent()

                await self.redis.hset("bot:status", mapping={
                    "online": "1",
                    "pid": str(os.getpid()),
                    "started_at": str(int(self.start_time.timestamp()))
                })

                await self.redis.hset("bot:metrics", mapping={
                    "memory_mb": str(int(memory_mb)),
                    "cpu_percent": str(int(cpu_percent))
                })

                await self.redis.expire("bot:status", 30)
                await self.redis.expire("bot:metrics", 30)

            except Exception as e:
                logger.error(f"Error reporting metrics: {e}")

            await asyncio.sleep(5)

    async def start(self):
        """Start the bot"""
        self.running = True
        await self.initialize()

        # Set bot online status
        await self.redis.hset("bot:status", "online", "1")

        logger.info("Bot starting...")
        await self.application.initialize()
        await self.application.start()
        await self.application.updater.start_polling()

        # Keep running
        while self.running:
            await asyncio.sleep(1)

    async def stop(self):
        """Stop the bot gracefully"""
        self.running = False
        logger.info("Bot stopping...")

        # Set offline status
        await self.redis.hset("bot:status", "online", "0")

        if self.application:
            await self.application.updater.stop()
            await self.application.stop()
            await self.application.shutdown()

        if self.db:
            await self.db.disconnect()

        if self.redis:
            await self.redis.disconnect()

        logger.info("Bot stopped")


async def main():
    bot = BotEngine()

    # Handle shutdown signals
    loop = asyncio.get_event_loop()

    def signal_handler():
        asyncio.create_task(bot.stop())

    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, signal_handler)

    try:
        await bot.start()
    except Exception as e:
        logger.error(f"Bot error: {e}")
        await bot.stop()


if __name__ == "__main__":
    asyncio.run(main())
