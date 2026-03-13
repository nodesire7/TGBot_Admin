import asyncpg
import json
from typing import Optional, Dict, Any, List
from datetime import datetime


class Database:
    def __init__(self, database_url: str):
        self.database_url = database_url
        self.pool: Optional[asyncpg.Pool] = None

    async def connect(self):
        """Create connection pool"""
        self.pool = await asyncpg.create_pool(
            self.database_url,
            min_size=5,
            max_size=20,
            command_timeout=60
        )

    async def disconnect(self):
        """Close connection pool"""
        if self.pool:
            await self.pool.close()

    # ==================== Groups ====================

    async def get_group(self, chat_id: int) -> Optional[Dict[str, Any]]:
        """Get group configuration by chat_id"""
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(
                "SELECT * FROM groups WHERE chat_id = $1",
                chat_id
            )
            return dict(row) if row else None

    async def upsert_group(self, chat_id: int, title: str, username: str = None,
                           member_count: int = 0) -> Dict[str, Any]:
        """Create or update group"""
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(
                """
                INSERT INTO groups (chat_id, title, username, member_count, last_active_at)
                VALUES ($1, $2, $3, $4, NOW())
                ON CONFLICT (chat_id) DO UPDATE SET
                    title = EXCLUDED.title,
                    username = EXCLUDED.username,
                    member_count = EXCLUDED.member_count,
                    last_active_at = NOW()
                RETURNING *
                """,
                chat_id, title, username, member_count
            )
            return dict(row)

    async def update_group_config(self, chat_id: int, config: Dict[str, Any]) -> bool:
        """Update group configuration"""
        async with self.pool.acquire() as conn:
            result = await conn.execute(
                "UPDATE groups SET config = $1, updated_at = NOW() WHERE chat_id = $2",
                json.dumps(config), chat_id
            )
            return result == "UPDATE 1"

    # ==================== Blacklist ====================

    async def is_blacklisted(self, chat_id: int, user_id: int) -> bool:
        """Check if user is blacklisted in group"""
        async with self.pool.acquire() as conn:
            exists = await conn.fetchval(
                "SELECT EXISTS(SELECT 1 FROM blacklist WHERE chat_id = $1 AND user_id = $2)",
                chat_id, user_id
            )
            return exists

    async def add_to_blacklist(self, chat_id: int, user_id: int, username: str = None,
                                first_name: str = None, reason: str = None,
                                banned_by: int = None) -> bool:
        """Add user to blacklist"""
        async with self.pool.acquire() as conn:
            await conn.execute(
                """
                INSERT INTO blacklist (chat_id, user_id, username, first_name, reason, banned_by)
                VALUES ($1, $2, $3, $4, $5, $6)
                ON CONFLICT (chat_id, user_id) DO UPDATE SET reason = $5
                """,
                chat_id, user_id, username, first_name, reason, banned_by
            )
            return True

    async def remove_from_blacklist(self, chat_id: int, user_id: int) -> bool:
        """Remove user from blacklist"""
        async with self.pool.acquire() as conn:
            result = await conn.execute(
                "DELETE FROM blacklist WHERE chat_id = $1 AND user_id = $2",
                chat_id, user_id
            )
            return result == "DELETE 1"

    # ==================== Verification Logs ====================

    async def log_verification(self, chat_id: int, user_id: int, status: str,
                               username: str = None, first_name: str = None,
                               question: str = None, answer: str = None,
                               user_answer: str = None, attempt_count: int = 1,
                               duration_seconds: int = None) -> int:
        """Log verification attempt"""
        async with self.pool.acquire() as conn:
            return await conn.fetchval(
                """
                INSERT INTO verification_logs
                (chat_id, user_id, status, username, first_name, question, answer,
                 user_answer, attempt_count, duration_seconds)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
                RETURNING id
                """,
                chat_id, user_id, status, username, first_name, question, answer,
                user_answer, attempt_count, duration_seconds
            )

    # ==================== Action Logs ====================

    async def log_action(self, action_type: str, chat_id: int = None,
                         user_id: int = None, action_data: Dict = None,
                         operator_id: int = None) -> int:
        """Log bot action"""
        async with self.pool.acquire() as conn:
            return await conn.fetchval(
                """
                INSERT INTO action_logs (action_type, chat_id, user_id, action_data, operator_id)
                VALUES ($1, $2, $3, $4, $5)
                RETURNING id
                """,
                action_type, chat_id, user_id, json.dumps(action_data or {}), operator_id
            )

    # ==================== Stats ====================

    async def get_today_stats(self, chat_id: int = None) -> Dict[str, int]:
        """Get today's verification statistics"""
        async with self.pool.acquire() as conn:
            query = """
                SELECT
                    COUNT(*) FILTER (WHERE status = 'success') as success,
                    COUNT(*) FILTER (WHERE status = 'failed') as failed,
                    COUNT(*) FILTER (WHERE status = 'timeout') as timeout
                FROM verification_logs
                WHERE DATE(created_at) = CURRENT_DATE
            """
            params = []
            if chat_id:
                query += " AND chat_id = $1"
                params.append(chat_id)

            row = await conn.fetchrow(query, *params)
            return dict(row) if row else {"success": 0, "failed": 0, "timeout": 0}
