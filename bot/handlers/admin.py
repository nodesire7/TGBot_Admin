from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import Application, CommandHandler, CallbackQueryHandler, ContextTypes, filters
import json


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
        app.add_handler(CommandHandler("setconfig", self.handle_setconfig, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("settimeout", self.handle_settimeout, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("setdifficulty", self.handle_setdifficulty, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("setmaxfail", self.handle_setmaxfail, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("kickonfail", self.handle_kickonfail, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("autoapprove", self.handle_autoapprove, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("ban", self.handle_ban, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("unban", self.handle_unban, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("blacklist", self.handle_blacklist, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("webui", self.handle_webui))
        app.add_handler(CommandHandler("stats", self.handle_stats, filters.ChatType.GROUPS))
        app.add_handler(CommandHandler("resetstats", self.handle_resetstats, filters.ChatType.GROUPS))
        app.add_handler(CallbackQueryHandler(self.handle_callback, pattern=r"^admin_"))

    async def is_admin(self, context: ContextTypes.DEFAULT_TYPE, chat_id: int, user_id: int) -> bool:
        """Check if user is admin in the chat"""
        try:
            member = await context.bot.get_chat_member(chat_id, user_id)
            return member.status in ["administrator", "creator"]
        except:
            return False

    async def handle_start(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /start command"""
        is_group = update.message.chat.type in ["group", "supergroup"]

        if is_group:
            await update.message.reply_text(
                "🤖 **TG Admin Bot**\n\n"
                "群组管理命令：\n"
                "/status - 查看群组状态\n"
                "/enable - 启用验证\n"
                "/disable - 禁用验证\n"
                "/config - 查看当前配置\n"
                "/setconfig <key> <value> - 修改配置\n"
                "/settimeout <seconds> - 设置验证超时\n"
                "/setdifficulty <easy|medium|hard> - 设置题目难度\n"
                "/setmaxfail <count> - 设置最大失败次数\n"
                "/kickonfail <on|off> - 验证失败是否踢出\n"
                "/autoapprove <on|off> - 自动通过入群申请\n"
                "/ban <user_id> [reason] - 封禁用户\n"
                "/unban <user_id> - 解封用户\n"
                "/blacklist - 查看黑名单\n"
                "/stats - 查看统计\n"
                "/webui - 获取管理面板链接",
                parse_mode="Markdown"
            )
        else:
            await update.message.reply_text(
                "🤖 **TG Admin Bot 管理面板**\n\n"
                "欢迎使用 Bot 管理系统！\n\n"
                "**私聊命令：**\n"
                "/webui - 获取 Web 管理面板链接\n"
                "/help - 查看帮助\n\n"
                "**群组命令：**\n"
                "将 Bot 添加到群组后，使用 /help 查看可用命令。",
                parse_mode="Markdown"
            )

    async def handle_help(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /help command"""
        await self.handle_start(update, context)

    async def handle_status(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /status command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        group = await self.db.get_group(chat_id)
        if not group:
            # Auto register group
            await self.db.upsert_group(chat_id, update.message.chat.title or "Unknown")
            group = await self.db.get_group(chat_id)

        stats = await self.db.get_today_stats(chat_id)
        config = group.get("config", {})

        status_emoji = "✅" if group.get("is_active") else "❌"

        await update.message.reply_text(
            f"📊 **群组状态**\n\n"
            f"群组：{group.get('title')}\n"
            f"验证状态：{status_emoji}\n"
            f"成员数：{group.get('member_count', 'N/A')}\n\n"
            f"**今日统计**\n"
            f"✅ 验证成功：{stats.get('success', 0)}\n"
            f"❌ 验证失败：{stats.get('failed', 0)}\n"
            f"⏰ 验证超时：{stats.get('timeout', 0)}\n\n"
            f"**当前配置**\n"
            f"验证超时：{config.get('verification_timeout', 300)}秒\n"
            f"题目难度：{config.get('difficulty', 'easy')}\n"
            f"最大失败次数：{config.get('max_fail_count', 3)}\n"
            f"失败踢出：{'是' if config.get('kick_on_fail', True) else '否'}\n"
            f"自动通过：{'是' if config.get('auto_approve', False) else '否'}",
            parse_mode="Markdown"
        )

    async def handle_enable(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /enable command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        group = await self.db.get_group(chat_id)
        if group:
            config = group.get("config", {})
            config["is_active"] = True
            await self.db.update_group_config(chat_id, config)
        else:
            await self.db.upsert_group(chat_id, update.message.chat.title or "Unknown")

        # Update is_active field
        async with self.db.pool.acquire() as conn:
            await conn.execute(
                "UPDATE groups SET is_active = TRUE WHERE chat_id = $1",
                chat_id
            )

        await self.redis.invalidate_group_cache(chat_id)
        await update.message.reply_text("✅ 验证功能已启用")

        await self.db.log_action("verification.enabled", chat_id=chat_id, operator_id=user_id)

    async def handle_disable(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /disable command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        async with self.db.pool.acquire() as conn:
            await conn.execute(
                "UPDATE groups SET is_active = FALSE WHERE chat_id = $1",
                chat_id
            )

        await self.redis.invalidate_group_cache(chat_id)
        await update.message.reply_text("❌ 验证功能已禁用")

        await self.db.log_action("verification.disabled", chat_id=chat_id, operator_id=user_id)

    async def handle_config(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /config command - show all config with buttons"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        group = await self.db.get_group(chat_id)
        if not group:
            await update.message.reply_text("❌ 群组未注册，请先使用 /enable 启用 Bot")
            return

        config = group.get("config", {})

        # Create inline keyboard for quick settings
        keyboard = [
            [
                InlineKeyboardButton(
                    f"验证超时: {config.get('verification_timeout', 300)}秒",
                    callback_data="admin_config_timeout"
                )
            ],
            [
                InlineKeyboardButton(
                    f"难度: {config.get('difficulty', 'easy')}",
                    callback_data="admin_config_difficulty"
                ),
                InlineKeyboardButton(
                    f"最大失败: {config.get('max_fail_count', 3)}次",
                    callback_data="admin_config_maxfail"
                )
            ],
            [
                InlineKeyboardButton(
                    f"失败踢出: {'✅' if config.get('kick_on_fail', True) else '❌'}",
                    callback_data="admin_config_kickonfail"
                ),
                InlineKeyboardButton(
                    f"自动通过: {'✅' if config.get('auto_approve', False) else '❌'}",
                    callback_data="admin_config_autoapprove"
                )
            ],
            [
                InlineKeyboardButton("🔄 重置为默认", callback_data="admin_config_reset")
            ]
        ]

        reply_markup = InlineKeyboardMarkup(keyboard)

        await update.message.reply_text(
            f"⚙️ **群组配置管理**\n\n"
            f"当前配置：\n"
            f"• 验证超时：{config.get('verification_timeout', 300)} 秒\n"
            f"• 题目难度：{config.get('difficulty', 'easy')}\n"
            f"• 最大失败次数：{config.get('max_fail_count', 3)}\n"
            f"• 失败踢出：{'是' if config.get('kick_on_fail', True) else '否'}\n"
            f"• 自动通过：{'是' if config.get('auto_approve', False) else '否'}\n\n"
            f"点击下方按钮快速修改，或使用命令：\n"
            f"/settimeout <秒数>\n"
            f"/setdifficulty <easy|medium|hard>\n"
            f"/setmaxfail <次数>\n"
            f"/kickonfail <on|off>\n"
            f"/autoapprove <on|off>",
            reply_markup=reply_markup,
            parse_mode="Markdown"
        )

    async def handle_callback(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle inline button callbacks"""
        query = update.callback_query
        await query.answer()

        chat_id = query.message.chat_id
        user_id = query.from_user.id
        data = query.data

        if not data.startswith("admin_"):
            return

        if not await self.is_admin(context, chat_id, user_id):
            await query.answer("只有管理员可以修改配置", show_alert=True)
            return

        group = await self.db.get_group(chat_id)
        if not group:
            await query.answer("群组未注册", show_alert=True)
            return

        config = group.get("config", {})
        modified = False

        if data == "admin_config_timeout":
            # Cycle through timeout options
            timeouts = [60, 120, 300, 600, 900]
            current = config.get('verification_timeout', 300)
            idx = timeouts.index(current) if current in timeouts else 2
            config['verification_timeout'] = timeouts[(idx + 1) % len(timeouts)]
            modified = True

        elif data == "admin_config_difficulty":
            # Cycle through difficulties
            difficulties = ['easy', 'medium', 'hard']
            current = config.get('difficulty', 'easy')
            idx = difficulties.index(current) if current in difficulties else 0
            config['difficulty'] = difficulties[(idx + 1) % len(difficulties)]
            modified = True

        elif data == "admin_config_maxfail":
            # Cycle through max fail counts
            counts = [1, 2, 3, 5, 10]
            current = config.get('max_fail_count', 3)
            idx = counts.index(current) if current in counts else 2
            config['max_fail_count'] = counts[(idx + 1) % len(counts)]
            modified = True

        elif data == "admin_config_kickonfail":
            config['kick_on_fail'] = not config.get('kick_on_fail', True)
            modified = True

        elif data == "admin_config_autoapprove":
            config['auto_approve'] = not config.get('auto_approve', False)
            modified = True

        elif data == "admin_config_reset":
            config = {
                'verification_timeout': 300,
                'difficulty': 'easy',
                'auto_approve': False,
                'kick_on_fail': True,
                'max_fail_count': 3,
                'admin_whitelist': config.get('admin_whitelist', [])
            }
            modified = True

        if modified:
            await self.db.update_group_config(chat_id, config)
            await self.redis.invalidate_group_cache(chat_id)
            await self.db.log_action("config.updated", chat_id=chat_id,
                                     action_data={"changes": data}, operator_id=user_id)

            # Update the message
            keyboard = [
                [
                    InlineKeyboardButton(
                        f"验证超时: {config.get('verification_timeout', 300)}秒",
                        callback_data="admin_config_timeout"
                    )
                ],
                [
                    InlineKeyboardButton(
                        f"难度: {config.get('difficulty', 'easy')}",
                        callback_data="admin_config_difficulty"
                    ),
                    InlineKeyboardButton(
                        f"最大失败: {config.get('max_fail_count', 3)}次",
                        callback_data="admin_config_maxfail"
                    )
                ],
                [
                    InlineKeyboardButton(
                        f"失败踢出: {'✅' if config.get('kick_on_fail', True) else '❌'}",
                        callback_data="admin_config_kickonfail"
                    ),
                    InlineKeyboardButton(
                        f"自动通过: {'✅' if config.get('auto_approve', False) else '❌'}",
                        callback_data="admin_config_autoapprove"
                    )
                ],
                [
                    InlineKeyboardButton("🔄 重置为默认", callback_data="admin_config_reset")
                ]
            ]

            await query.edit_message_reply_markup(reply_markup=InlineKeyboardMarkup(keyboard))
            await query.answer("✅ 配置已更新", show_alert=False)

    async def handle_setconfig(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /setconfig command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        args = context.args
        if len(args) < 2:
            await update.message.reply_text(
                "用法：/setconfig <key> <value>\n\n"
                "可配置项：\n"
                "• verification_timeout (验证超时秒数)\n"
                "• difficulty (easy/medium/hard)\n"
                "• max_fail_count (最大失败次数)\n"
                "• kick_on_fail (true/false)\n"
                "• auto_approve (true/false)"
            )
            return

        key = args[0].lower()
        value = " ".join(args[1])

        group = await self.db.get_group(chat_id)
        if not group:
            await update.message.reply_text("❌ 群组未注册")
            return

        config = group.get("config", {})

        # Parse value based on key
        try:
            if key in ["verification_timeout", "max_fail_count"]:
                config[key] = int(value)
            elif key in ["kick_on_fail", "auto_approve"]:
                config[key] = value.lower() in ["true", "yes", "on", "1"]
            elif key == "difficulty":
                if value.lower() not in ["easy", "medium", "hard"]:
                    await update.message.reply_text("❌ 难度必须是 easy, medium 或 hard")
                    return
                config[key] = value.lower()
            else:
                await update.message.reply_text(f"❌ 未知配置项: {key}")
                return
        except ValueError:
            await update.message.reply_text("❌ 值格式错误")
            return

        await self.db.update_group_config(chat_id, config)
        await self.redis.invalidate_group_cache(chat_id)
        await update.message.reply_text(f"✅ 已更新 {key} = {config[key]}")

    async def handle_settimeout(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /settimeout command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        args = context.args
        if not args:
            await update.message.reply_text("用法：/settimeout <秒数>\n示例：/settimeout 300")
            return

        try:
            timeout = int(args[0])
            if timeout < 30 or timeout > 3600:
                await update.message.reply_text("❌ 超时时间必须在 30-3600 秒之间")
                return
        except ValueError:
            await update.message.reply_text("❌ 请输入有效的数字")
            return

        await self._update_config(chat_id, "verification_timeout", timeout)
        await update.message.reply_text(f"✅ 验证超时已设置为 {timeout} 秒")

    async def handle_setdifficulty(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /setdifficulty command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        args = context.args
        if not args or args[0].lower() not in ["easy", "medium", "hard"]:
            await update.message.reply_text(
                "用法：/setdifficulty <easy|medium|hard>\n\n"
                "• easy - 个位数加减法\n"
                "• medium - 两位数加减法\n"
                "• hard - 两位数乘除法"
            )
            return

        difficulty = args[0].lower()
        await self._update_config(chat_id, "difficulty", difficulty)

        difficulty_names = {"easy": "简单", "medium": "中等", "hard": "困难"}
        await update.message.reply_text(f"✅ 题目难度已设置为 {difficulty_names[difficulty]}")

    async def handle_setmaxfail(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /setmaxfail command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        args = context.args
        if not args:
            await update.message.reply_text("用法：/setmaxfail <次数>\n示例：/setmaxfail 3")
            return

        try:
            count = int(args[0])
            if count < 1 or count > 20:
                await update.message.reply_text("❌ 次数必须在 1-20 之间")
                return
        except ValueError:
            await update.message.reply_text("❌ 请输入有效的数字")
            return

        await self._update_config(chat_id, "max_fail_count", count)
        await update.message.reply_text(f"✅ 最大失败次数已设置为 {count}")

    async def handle_kickonfail(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /kickonfail command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        args = context.args
        if not args or args[0].lower() not in ["on", "off", "yes", "no", "true", "false"]:
            await update.message.reply_text("用法：/kickonfail <on|off>")
            return

        enabled = args[0].lower() in ["on", "yes", "true"]
        await self._update_config(chat_id, "kick_on_fail", enabled)
        status = "开启" if enabled else "关闭"
        await update.message.reply_text(f"✅ 验证失败踢人已{status}")

    async def handle_autoapprove(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /autoapprove command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        args = context.args
        if not args or args[0].lower() not in ["on", "off", "yes", "no", "true", "false"]:
            await update.message.reply_text("用法：/autoapprove <on|off>")
            return

        enabled = args[0].lower() in ["on", "yes", "true"]
        await self._update_config(chat_id, "auto_approve", enabled)
        status = "开启" if enabled else "关闭"
        await update.message.reply_text(f"✅ 自动通过入群申请已{status}")

    async def _update_config(self, chat_id: int, key: str, value):
        """Helper to update a single config key"""
        group = await self.db.get_group(chat_id)
        if not group:
            group = await self.db.upsert_group(chat_id, "Unknown")

        config = group.get("config", {})
        config[key] = value
        await self.db.update_group_config(chat_id, config)
        await self.redis.invalidate_group_cache(chat_id)

    async def handle_ban(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /ban command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        args = context.args
        if not args:
            await update.message.reply_text("用法：/ban <user_id> [reason]\n示例：/ban 123456789 垃圾广告")
            return

        try:
            target_user_id = int(args[0])
        except ValueError:
            await update.message.reply_text("❌ user_id 必须是数字")
            return

        reason = " ".join(args[1:]) if len(args) > 1 else None

        await self.db.add_to_blacklist(
            chat_id, target_user_id,
            reason=reason, banned_by=user_id
        )

        await update.message.reply_text(
            f"✅ 用户 `{target_user_id}` 已加入黑名单",
            parse_mode="Markdown"
        )

        await self.db.log_action(
            "user.banned",
            chat_id=chat_id,
            user_id=target_user_id,
            action_data={"reason": reason},
            operator_id=user_id
        )

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
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        args = context.args
        if not args:
            await update.message.reply_text("用法：/unban <user_id>\n示例：/unban 123456789")
            return

        try:
            target_user_id = int(args[0])
        except ValueError:
            await update.message.reply_text("❌ user_id 必须是数字")
            return

        removed = await self.db.remove_from_blacklist(chat_id, target_user_id)

        if removed:
            await update.message.reply_text(
                f"✅ 用户 `{target_user_id}` 已从黑名单移除",
                parse_mode="Markdown"
            )
        else:
            await update.message.reply_text(
                f"⚠️ 用户 `{target_user_id}` 不在黑名单中",
                parse_mode="Markdown"
            )

        await self.db.log_action(
            "user.unbanned",
            chat_id=chat_id,
            user_id=target_user_id,
            operator_id=user_id
        )

    async def handle_blacklist(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /blacklist command - show blacklist"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        async with self.db.pool.acquire() as conn:
            rows = await conn.fetch(
                "SELECT user_id, username, first_name, reason, created_at FROM blacklist WHERE chat_id = $1 ORDER BY created_at DESC LIMIT 20",
                chat_id
            )

        if not rows:
            await update.message.reply_text("📋 黑名单为空")
            return

        message = "📋 **黑名单列表**\n\n"
        for row in rows:
            username = row['username'] or row['first_name'] or 'Unknown'
            reason = f" - {row['reason']}" if row['reason'] else ""
            message += f"• `{row['user_id']}` (@{username}){reason}\n"

        if len(rows) == 20:
            message += "\n_仅显示最近 20 条记录_"

        await update.message.reply_text(message, parse_mode="Markdown")

    async def handle_webui(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /webui command"""
        # Get API URL from environment or default
        import os
        api_url = os.getenv("WEBUI_URL", "http://localhost:8000")

        await update.message.reply_text(
            "🌐 **Web 管理面板**\n\n"
            f"访问地址：{api_url}\n\n"
            "默认登录信息：\n"
            "用户名：admin\n"
            "密码：admin123\n\n"
            "⚠️ 请登录后立即修改默认密码！",
            parse_mode="Markdown"
        )

    async def handle_stats(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /stats command"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        # Get today's stats
        today_stats = await self.db.get_today_stats(chat_id)

        # Get all-time stats
        async with self.db.pool.acquire() as conn:
            all_time = await conn.fetchrow(
                """
                SELECT
                    COUNT(*) FILTER (WHERE status = 'success') as success,
                    COUNT(*) FILTER (WHERE status = 'failed') as failed,
                    COUNT(*) FILTER (WHERE status = 'timeout') as timeout,
                    COUNT(*) as total
                FROM verification_logs
                WHERE chat_id = $1
                """,
                chat_id
            )

        await update.message.reply_text(
            f"📊 **群组统计**\n\n"
            f"**今日统计**\n"
            f"✅ 验证成功：{today_stats.get('success', 0)}\n"
            f"❌ 验证失败：{today_stats.get('failed', 0)}\n"
            f"⏰ 验证超时：{today_stats.get('timeout', 0)}\n\n"
            f"**累计统计**\n"
            f"✅ 验证成功：{all_time['success'] or 0}\n"
            f"❌ 验证失败：{all_time['failed'] or 0}\n"
            f"⏰ 验证超时：{all_time['timeout'] or 0}\n"
            f"📝 总计：{all_time['total'] or 0}",
            parse_mode="Markdown"
        )

    async def handle_resetstats(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /resetstats command - clear verification logs"""
        chat_id = update.message.chat_id
        user_id = update.message.from_user.id

        if not await self.is_admin(context, chat_id, user_id):
            await update.message.reply_text("❌ 只有管理员可以执行此命令。")
            return

        # This is a destructive action, require confirmation
        args = context.args
        if not args or args[0] != "confirm":
            await update.message.reply_text(
                "⚠️ **危险操作**\n\n"
                "此命令将清空本群所有验证日志记录！\n\n"
                "如果确认，请使用：/resetstats confirm",
                parse_mode="Markdown"
            )
            return

        async with self.db.pool.acquire() as conn:
            await conn.execute(
                "DELETE FROM verification_logs WHERE chat_id = $1",
                chat_id
            )

        await update.message.reply_text("✅ 验证日志已清空")
        await self.db.log_action("stats.reset", chat_id=chat_id, operator_id=user_id)
