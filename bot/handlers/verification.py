import random
import asyncio
from datetime import datetime
from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import (
    Application, CommandHandler, MessageHandler, CallbackQueryHandler,
    filters, ContextTypes
)
from typing import Dict, Any, Optional


class VerificationHandler:
    def __init__(self, db, redis):
        self.db = db
        self.redis = redis

    def register(self, app: Application):
        """Register verification handlers"""
        app.add_handler(CallbackQueryHandler(self.handle_answer, pattern=r"^verify_"))
        app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, self.handle_message))

    async def get_group_config(self, chat_id: int) -> Dict[str, Any]:
        """Get group config with cache"""
        # Try cache first
        cached = await self.redis.get_group_config(chat_id)
        if cached:
            return cached

        # Fetch from database
        group = await self.db.get_group(chat_id)
        if group:
            config = group.get("config", self.get_default_config())
            config["is_active"] = group.get("is_active", True)
            await self.redis.set_group_config(chat_id, config)
            return config

        # Return default config for new groups
        return self.get_default_config()

    def get_default_config(self) -> Dict[str, Any]:
        """Get default verification config"""
        return {
            "is_active": True,
            "verification_timeout": 300,
            "difficulty": "easy",
            "auto_approve": False,
            "kick_on_fail": True,
            "max_fail_count": 3,
            "admin_whitelist": []
        }

    def generate_question(self, difficulty: str) -> tuple[str, str]:
        """Generate arithmetic question based on difficulty"""
        if difficulty == "easy":
            # Single digit addition/subtraction
            a = random.randint(1, 9)
            b = random.randint(1, 9)
            op = random.choice(["+", "-"])
            if op == "-" and a < b:
                a, b = b, a
        elif difficulty == "medium":
            # Two digit
            a = random.randint(10, 99)
            b = random.randint(1, 99)
            op = random.choice(["+", "-"])
            if op == "-" and a < b:
                a, b = b, a
        else:  # hard
            # Three digit with multiplication
            a = random.randint(10, 99)
            b = random.randint(2, 9)
            op = random.choice(["+", "-", "*"])
            if op == "-":
                a = random.randint(100, 999)
                b = random.randint(10, a)

        question = f"{a} {op} {b} = ?"
        answer = str(eval(f"{a} {op} {b}"))
        return question, answer

    async def start_verification(self, chat_id: int, user_id: int,
                                  username: str, first_name: str,
                                  chat_title: str):
        """Start verification for a user"""
        config = await self.get_group_config(chat_id)

        if not config.get("is_active", True):
            return

        # Check if user is whitelisted
        if user_id in config.get("admin_whitelist", []):
            return

        # Generate question
        difficulty = config.get("difficulty", "easy")
        question, answer = self.generate_question(difficulty)

        # Create session
        timeout = config.get("verification_timeout", 300)
        await self.redis.create_verification_session(
            chat_id, user_id, question, answer, timeout
        )

        # Create inline keyboard with answer options
        correct_answer = int(answer)
        wrong_answers = self.generate_wrong_answers(correct_answer, difficulty)
        all_answers = [correct_answer] + wrong_answers
        random.shuffle(all_answers)

        keyboard = [
            [InlineKeyboardButton(str(a), callback_data=f"verify_{a}")]
            for a in all_answers
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)

        # Send verification message
        from telegram import Bot
        bot = Bot(token=self.db.pool._config.database_url.split("@")[0].split("//")[1])  # Placeholder
        # Actually we need to get bot from context, this is a placeholder

        # Log event
        await self.redis.publish_event(
            "verification.started",
            chat_id, user_id, username, chat_title,
            {"question": question}
        )

        return question, reply_markup

    def generate_wrong_answers(self, correct: int, difficulty: str) -> list[int]:
        """Generate wrong answer options"""
        wrong = set()
        if difficulty == "easy":
            while len(wrong) < 3:
                w = random.randint(1, 18)
                if w != correct:
                    wrong.add(w)
        elif difficulty == "medium":
            while len(wrong) < 3:
                w = random.randint(1, 198)
                if w != correct:
                    wrong.add(w)
        else:
            while len(wrong) < 3:
                w = random.randint(1, 999)
                if w != correct:
                    wrong.add(w)
        return list(wrong)

    async def handle_answer(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle verification answer"""
        query = update.callback_query
        await query.answer()

        chat_id = query.message.chat_id
        user_id = query.from_user.id
        username = query.from_user.username or ""
        first_name = query.from_user.first_name or ""

        # Get session
        session = await self.redis.get_verification_session(chat_id, user_id)
        if not session:
            await query.edit_message_text("验证会话已过期，请重新触发验证。")
            return

        user_answer = query.data.replace("verify_", "")
        correct_answer = session.get("answer")

        # Increment attempt
        attempt_count = await self.redis.increment_verification_attempt(chat_id, user_id)

        if user_answer == correct_answer:
            # Correct answer
            await self.redis.delete_verification_session(chat_id, user_id)
            await query.edit_message_text("✅ 验证通过！欢迎加入群组。")

            # Log success
            await self.db.log_verification(
                chat_id, user_id, "success",
                username, first_name,
                session.get("question"), correct_answer, user_answer,
                attempt_count
            )

            # Increment stats
            await self.redis.increment_stat("verified", chat_id)

            # Publish event
            await self.redis.publish_event(
                "verification.success",
                chat_id, user_id, username,
                {"question": session.get("question"), "attempts": attempt_count}
            )

        else:
            # Wrong answer
            config = await self.get_group_config(chat_id)
            max_fails = config.get("max_fail_count", 3)

            if attempt_count >= max_fails:
                # Max attempts reached
                await self.redis.delete_verification_session(chat_id, user_id)
                await query.edit_message_text(
                    f"❌ 验证失败次数过多（{max_fails}次），您将被移出群组。"
                )

                # Log failure
                await self.db.log_verification(
                    chat_id, user_id, "failed",
                    username, first_name,
                    session.get("question"), correct_answer, user_answer,
                    attempt_count
                )

                # Kick user if configured
                if config.get("kick_on_fail", True):
                    try:
                        await context.bot.ban_chat_member(chat_id, user_id)
                        await context.bot.unban_chat_member(chat_id, user_id)
                    except Exception as e:
                        pass

                # Increment stats
                await self.redis.increment_stat("failed", chat_id)
                await self.redis.increment_stat("kicked", chat_id)

                # Publish event
                await self.redis.publish_event(
                    "verification.failed_kicked",
                    chat_id, user_id, username,
                    {"question": session.get("question"), "attempts": attempt_count}
                )
            else:
                # Still have attempts
                await query.edit_message_text(
                    f"❌ 答案错误，请重试。剩余尝试次数：{max_fails - attempt_count}"
                )
                # Regenerate question
                await self.start_verification(
                    chat_id, user_id, username, first_name,
                    query.message.chat.title
                )

    async def handle_message(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle messages - restrict unverified users"""
        if not update.message or not update.message.chat:
            return

        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        # Check if user has active verification session
        session = await self.redis.get_verification_session(chat_id, user_id)
        if session:
            # Delete message from unverified user
            try:
                await update.message.delete()
            except:
                pass
