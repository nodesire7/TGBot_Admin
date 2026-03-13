from telegram import Update
from telegram.ext import Application, CommandHandler, ContextTypes, filters


class AdminHandler:
    def __init__(self, db, redis):
        self.db = db
        self.redis = redis

    def register(self, app: Application):
        """Register admin command handlers"""
        app.add_handler(CommandHandler("start", self.handle_start))
        app.add_handler(CommandHandler("help", self.handle_help))
        app.add_handler(CommandHandler("status", self.handle_status, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("enable", self.handle_enable, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("disable", self.handle_disable, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("config", self.handle_config, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("ban", self.handle_ban, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("unban", self.handle_unban, filters.ChatType.GROUPS))

    async def is_admin(self, context: ContextTypes.DEFAULT_TYPE, chat_id: int, user_id: int) -> bool:
        """Check if user is admin in the chat"""
        try:
            member = await context.bot.get_chat_member(chat_id, user_id)
            return member.status in ["administrator", "creator"]
        except:
            return False

    async def handle_start(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /start command"""
        await update.message.reply_text(
            "🤖 TG Admin Bot\n\n"
            "我是群组管理机器人，提供入群验证、防垃圾等功能。\n\n"
            "命令列表：\n"
            "/status - 查看群组状态\n"
            "/enable - 启用验证\n"
            "/disable - 禁用验证\n"
            "/config - 查看配置\n"
            "/ban <user_id> [reason] - 封禁用户\n"
            "/unban <user_id> - 解封用户"
        )

    async def handle_help(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /help command"""
        await self.handle_start(update, context)

    async def handle_status(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /status command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        # Check admin
        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("只有管理员可以执行此命令。")
            return

        # Get group info
        group = await self.db.get_group(chat_id)
        if not group:
            await update.message.reply_text("群组未在数据库中注册，请先让机器人加入群组。")
            return

        # Get stats
        stats = await self.db.get_today_stats(chat_id)

        status_emoji = "✅" if group.get("is_active") else "❌"
        await update.message.reply_text(
            f"📊 群组状态\n\n"
            f"群组：{group.get('title')}\n"
            f"验证状态：{status_emoji}\n"
            f"今日验证成功：{stats.get('success', 0)}\n"
            f"今日验证失败：{stats.get('failed', 0)}\n"
            f"今日超时：{stats.get('timeout', 0)}"
        )

    async def handle_enable(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /enable command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("只有管理员可以执行此命令。")
            return

        group = await self.db.get_group(chat_id)
        if group:
            await self.db.update_group_config(chat_id, {"is_active": True})
            await self.redis.invalidate_group_cache(chat_id)

        await update.message.reply_text("✅ 验证功能已启用")

        # Log action
        await self.db.log_action(
            "verification.enabled",
            chat_id=chat_id,
            operator_id=user_id
        )

    async def handle_disable(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /disable command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("只有管理员可以执行此命令。")
            return

        group = await self.db.get_group(chat_id)
        if group:
            config = group.get("config", {})
            config["is_active"] = False
            await self.db.update_group_config(chat_id, config)
            await self.redis.invalidate_group_cache(chat_id)

        await update.message.reply_text("❌ 验证功能已禁用")

        # Log action
        await self.db.log_action(
            "verification.disabled",
            chat_id=chat_id,
            operator_id=user_id
        )

    async def handle_config(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /config command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("只有管理员可以执行此命令。")
            return

        group = await self.db.get_group(chat_id)
        if not group:
            await update.message.reply_text("群组未注册")
            return

        config = group.get("config", {})
        await update.message.reply_text(
            f"⚙️ 当前配置\n\n"
            f"验证超时：{config.get('verification_timeout', 300)}秒\n"
            f"题目难度：{config.get('difficulty', 'easy')}\n"
            f"自动同意：{'是' if config.get('auto_approve') else '否'}\n"
            f"失败踢出：{'是' if config.get('kick_on_fail') else '否'}\n"
            f"最大失败次数：{config.get('max_fail_count', 3)}\n\n"
            f"请在 Web 管理面板修改详细配置。"
        )

    async def handle_ban(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /ban command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("只有管理员可以执行此命令。")
            return

        args = context.args
        if not args:
            await update.message.reply_text("用法：/ban <user_id> [reason]")
            return

        try:
            target_user_id = int(args[0])
        except ValueError:
            await update.message.reply_text("user_id 必须是数字")
            return

        reason = " ".join(args[1:]) if len(args) > 1 else None

        # Add to blacklist
        await self.db.add_to_blacklist(
            chat_id, target_user_id,
            reason=reason, banned_by=user_id
        )

        await update.message.reply_text(f"✅ 用户 {target_user_id} 已加入黑名单")

        # Log action
        await self.db.log_action(
            "user.banned",
            chat_id=chat_id,
            user_id=target_user_id,
            action_data={"reason": reason},
            operator_id=user_id
        )

        # Publish event
        await self.redis.publish_event(
            "user.banned",
            chat_id, target_user_id, "", "",
            {"reason": reason, "operator_id": user_id}
        )

    async def handle_unban(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /unban command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("只有管理员可以执行此命令。")
            return

        args = context.args
        if not args:
            await update.message.reply_text("用法：/unban <user_id>")
            return

        try:
            target_user_id = int(args[0])
        except ValueError:
            await update.message.reply_text("user_id 必须是数字")
            return

        # Remove from blacklist
        removed = await self.db.remove_from_blacklist(chat_id, target_user_id)

        if removed:
            await update.message.reply_text(f"✅ 用户 {target_user_id} 已从黑名单移除")
        else:
            await update.message.reply_text(f"用户 {target_user_id} 不在黑名单中")

        # Log action
        await self.db.log_action(
            "user.unbanned",
            chat_id=chat_id,
            user_id=target_user_id,
            operator_id=user_id
        )
