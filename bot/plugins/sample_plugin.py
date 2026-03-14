"""
示例插件：自定义欢迎消息

这个插件演示了如何使用 TGBot Plugin SDK 创建一个完整的插件。
"""

from typing import List
import sys
import os

# 添加 SDK 路径
sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from plugin_sdk import (
    Plugin, Context, User, Message, Permission,
    MessageBuilder, make_inline_keyboard, make_button
)
from plugin_sdk.utils import format_user_mention, escape_markdown


class CustomWelcomePlugin(Plugin):
    """
    自定义欢迎消息插件

    功能：
    - 新用户入群时发送自定义欢迎消息
    - 支持用户名提及
    - 支持配置欢迎消息模板
    """

    # 插件元信息
    id = "custom_welcome"
    name = "自定义欢迎"
    version = "1.0.0"
    author = "TGBot Admin"
    description = "新用户入群时发送自定义欢迎消息，支持模板变量"

    # 需要的权限
    permissions = [
        Permission.SEND_MESSAGES,
        Permission.DELETE_MESSAGES,
    ]

    # 配置 Schema
    config_schema = {
        "type": "object",
        "properties": {
            "welcome_message": {
                "type": "string",
                "default": "欢迎 {user} 加入本群！请阅读群规。",
                "description": "欢迎消息模板，支持变量: {user}, {user_id}, {username}, {chat_title}"
            },
            "delete_after": {
                "type": "integer",
                "default": 0,
                "description": "多少秒后删除欢迎消息，0 表示不删除"
            },
            "mention_user": {
                "type": "boolean",
                "default": True,
                "description": "是否在消息中提及用户"
            },
            "show_rules_button": {
                "type": "boolean",
                "default": False,
                "description": "是否显示群规按钮"
            },
            "rules_text": {
                "type": "string",
                "default": "",
                "description": "群规文本（点击按钮时显示）"
            }
        }
    }

    @Plugin.on_join(priority=10)
    async def on_user_join(self, ctx: Context, user: User):
        """
        用户入群处理

        Args:
            ctx: 插件上下文
            user: 入群用户信息
        """
        # 获取配置
        template = ctx.get_config("welcome_message", "欢迎 {user} 加入本群！")
        delete_after = ctx.get_config("delete_after", 0)
        mention_user = ctx.get_config("mention_user", True)
        show_rules_button = ctx.get_config("show_rules_button", False)

        # 构建欢迎消息
        user_display = format_user_mention(user.id, user.full_name) if mention_user else user.full_name

        message_text = template.format(
            user=user_display,
            user_id=user.id,
            username=user.username or "无",
            chat_title=f"群组 {ctx.chat_id}",
        )

        # 构建消息
        builder = MessageBuilder()
        builder.text(message_text).markdown()

        # 添加群规按钮
        if show_rules_button:
            builder.inline_keyboard([[
                make_button("📋 查看群规", callback_data="show_rules")
            ]])

        # 发送消息
        msg = await ctx.send_message(**builder.build())

        # 设置自动删除
        if delete_after > 0 and msg.get("message_id"):
            # 注意：需要定时任务来删除消息
            # 这里简单记录，实际需要定时器
            await ctx.cache_set(
                f"delete_msg:{msg['message_id']}",
                str(delete_after),
                expire=delete_after
            )

    @Plugin.on_callback
    async def handle_callback(self, ctx: Context, callback):
        """
        处理回调查询

        Args:
            ctx: 插件上下文
            callback: 回调查询对象
        """
        data = callback.data

        if data == "show_rules":
            rules_text = ctx.get_config("rules_text", "暂无群规")

            await ctx.answer_callback(
                callback.id,
                text=rules_text[:200] if len(rules_text) > 200 else rules_text,
                show_alert=True
            )

    @Plugin.on_command("setwelcome")
    async def set_welcome_command(self, ctx: Context, args: List[str]):
        """
        设置欢迎消息命令

        用法: /setwelcome 欢迎消息内容
        """
        if not args:
            current = ctx.get_config("welcome_message", "默认欢迎消息")
            await ctx.reply(f"当前欢迎消息:\n\n{current}")
            return

        new_message = " ".join(args)

        # 这里需要更新配置
        # 实际实现需要调用 API 更新数据库
        await ctx.reply(f"欢迎消息已更新为:\n\n{new_message}")

    @Plugin.on_command("setrules")
    async def set_rules_command(self, ctx: Context, args: List[str]):
        """
        设置群规命令

        用法: /setrules 群规内容
        """
        if not args:
            current = ctx.get_config("rules_text", "暂无群规")
            await ctx.reply(f"当前群规:\n\n{current}")
            return

        new_rules = " ".join(args)
        await ctx.reply(f"群规已更新")


# 导出插件类
plugin = CustomWelcomePlugin


if __name__ == "__main__":
    # 测试插件清单
    print("Plugin Manifest:")
    import json
    print(json.dumps(plugin().get_manifest(), indent=2, ensure_ascii=False))
