import redis.asyncio as redis
import json
from typing import Optional, Dict, Any
from datetime import datetime


class RedisClient:
    def __init__(self, redis_url: str):
        self.redis_url = redis_url
        self.client: Optional[redis.Redis] = None

    async def connect(self):
        """Create Redis connection"""
        self.client = redis.from_url(
            self.redis_url,
            encoding="utf-8",
            decode_responses=True
        )

    async def disconnect(self):
        """Close Redis connection"""
        if self.client:
            await self.client.close()

    # ==================== Verification Session ====================

    async def create_verification_session(self, chat_id: int, user_id: int,
                                          question: str, answer: str,
                                          timeout: int = 300) -> str:
        """Create verification session"""
        key = f"verification:{chat_id}:{user_id}"
        data = {
            "question": question,
            "answer": answer,
            "attempt_count": 0,
            "created_at": datetime.now().isoformat()
        }
        await self.client.hset(key, mapping={k: json.dumps(v) if isinstance(v, (dict, list)) else str(v)
                                             for k, v in data.items()})
        await self.client.expire(key, timeout)
        return key

    async def get_verification_session(self, chat_id: int, user_id: int) -> Optional[Dict[str, Any]]:
        """Get verification session data"""
        key = f"verification:{chat_id}:{user_id}"
        data = await self.client.hgetall(key)
        if not data:
            return None

        result = {}
        for k, v in data.items():
            try:
                result[k] = json.loads(v)
            except (json.JSONDecodeError, TypeError):
                result[k] = v
        return result

    async def increment_verification_attempt(self, chat_id: int, user_id: int) -> int:
        """Increment verification attempt count"""
        key = f"verification:{chat_id}:{user_id}"
        count = await self.client.hincrby(key, "attempt_count", 1)
        return int(count)

    async def delete_verification_session(self, chat_id: int, user_id: int):
        """Delete verification session"""
        key = f"verification:{chat_id}:{user_id}"
        await self.client.delete(key)

    # ==================== Group Cache ====================

    async def get_group_config(self, chat_id: int) -> Optional[Dict[str, Any]]:
        """Get cached group config"""
        key = f"cache:group:{chat_id}"
        data = await self.client.get(key)
        if data:
            return json.loads(data)
        return None

    async def set_group_config(self, chat_id: int, config: Dict[str, Any], ttl: int = 600):
        """Cache group config"""
        key = f"cache:group:{chat_id}"
        await self.client.setex(key, ttl, json.dumps(config))

    async def invalidate_group_cache(self, chat_id: int):
        """Invalidate group config cache"""
        key = f"cache:group:{chat_id}"
        await self.client.delete(key)

    # ==================== Stats ====================

    async def increment_stat(self, stat_type: str, chat_id: int = None):
        """Increment daily stat counter"""
        today = datetime.now().strftime("%Y-%m-%d")
        key = f"stats:{today}:{stat_type}"
        await self.client.incr(key)
        await self.client.expire(key, 86400 * 7)  # Keep for 7 days

        if chat_id:
            key = f"stats:group:{chat_id}:{today}:{stat_type}"
            await self.client.incr(key)
            await self.client.expire(key, 86400 * 7)

    # ==================== Events ====================

    async def publish_event(self, event_type: str, chat_id: int, user_id: int,
                            username: str = None, chat_title: str = None,
                            data: Dict = None):
        """Publish event to stream and channel"""
        event_data = {
            "type": event_type,
            "chat_id": chat_id,
            "chat_title": chat_title or "",
            "user_id": user_id,
            "username": username or "",
            "timestamp": datetime.now().isoformat()
        }
        if data:
            event_data.update(data)

        # Add to stream
        await self.client.xadd("stream:events", event_data, maxlen=1000)

        # Publish to channel
        await self.client.publish("channel:events", json.dumps(event_data))

    # ==================== Blacklist Cache ====================

    async def is_blacklisted_cached(self, chat_id: int, user_id: int) -> Optional[bool]:
        """Check blacklist from cache"""
        key = f"cache:blacklist:{chat_id}"
        return await self.client.sismember(key, user_id)

    async def cache_blacklist(self, chat_id: int, user_ids: list, ttl: int = 300):
        """Cache blacklist for a group"""
        key = f"cache:blacklist:{chat_id}"
        if user_ids:
            await self.client.sadd(key, *user_ids)
        await self.client.expire(key, ttl)
