from telegram import Update
from telegram.ext import Application, ChatMemberHandler, ContextTypes
from telegram.ext.filters import StatusUpdate
from datetime import datetime


class MemberHandler:
    def __init__(self, db, redis):
        self.db = db
        self.redis = redis

    def register(self, app: Application):
        """Register member event handlers"""
        app.add_handler(ChatMemberHandler(self.handle_chat_member, ChatMemberHandler.CHAT_MEMBER))
        app.add_handler(StatusUpdate.USER_LEFT, self.handle_user_left)

    async def handle_chat_member(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle chat member updates (join/leave)"""
        result = update.chat_member

        if result.new_chat_member.status == result.new_chat_member.LEFT:
            # User left - no action needed
            return

        if result.new_chat_member.status in ["MEMBER", "RESTRICTED"]:
            # New member joined
            chat_id = result.chat.id
            user = result.new_chat_member.user
            user_id = user.id
            username = user.username or ""
            first_name = user.first_name or ""
            chat_title = result.chat.title or ""

            # Check if already verified (no active session)
            session = await self.redis.get_verification_session(chat_id, user_id)
            if session:
                return

            # Check blacklist
            is_blacklisted = await self.db.is_blacklisted(chat_id, user_id)
            if is_blacklisted:
                # Kick blacklisted user
                try:
                    await context.bot.ban_chat_member(chat_id, user_id)
                    await context.bot.unban_chat_member(chat_id, user_id)
                except:
                    pass

                # Log event
                await self.redis.publish_event(
                    "user.blacklisted_kicked",
                    chat_id, user_id, username, chat_title
                )
                return

            # Check group config
            group = await self.db.get_group(chat_id)
            if not group:
                # New group - create record
                await self.db.upsert_group(chat_id, chat_title, None, 0)
                group = await self.db.get_group(chat_id)

            config = group.get("config", {})
            if not group.get("is_active", True):
                return

            # Check admin whitelist
            if user_id in config.get("admin_whitelist", []):
                return

            # Check auto_approve
            if config.get("auto_approve", False):
                # Auto approve - log and skip verification
                await self.db.log_verification(
                    chat_id, user_id, "success",
                    username, first_name, status="auto_approved"
                )
                await self.redis.increment_stat("verified", chat_id)
                return

            # Start verification
            from .verification import VerificationHandler
            verification = VerificationHandler(self.db, self.redis)

            # Restrict user first
            try:
                await context.bot.restrict_chat_member(
                    chat_id, user_id,
                    permissions={
                        "can_send_messages": False,
                        "can_send_media_messages": False,
                        "can_send_polls": False,
                        "can_send_other_messages": False,
                    }
                )
            except Exception as e:
                pass

            # Send verification
            question, reply_markup = await verification.start_verification(
                chat_id, user_id, username, first_name, chat_title
            )

            if question and reply_markup:
                try:
                    await context.bot.send_message(
                        chat_id,
                        f"🔐 @{username or first_name} 请在 5 分钟内完成验证：\n\n"
                        f"计算：{question}",
                        reply_markup=reply_markup
                    )
                except Exception as e:
                    pass

    async def handle_user_left(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle user leaving chat"""
        if not update.message or not update.message.left_chat_member:
            return

        chat_id = update.message.chat_id
        user = update.message.left_chat_member
        user_id = user.id
        username = user.username or ""
        chat_title = update.message.chat.title or ""

        # Clean up verification session if exists
        await self.redis.delete_verification_session(chat_id, user_id)

        # Log event
        await self.redis.publish_event(
            "user.left",
            chat_id, user_id, username, chat_title
        )
