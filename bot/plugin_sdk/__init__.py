"""
TGBot Plugin SDK - 插件开发工具包

提供插件开发的基类和装饰器，让开发者可以轻松创建 Telegram Bot 插件。
"""

import asyncio
import functools
import inspect
from typing import Any, Callable, Dict, List, Optional, Union
from dataclasses import dataclass, field
from enum import Enum
import logging

logger = logging.getLogger(__name__)


class HookType(Enum):
    """钩子类型枚举"""
    ON_JOIN = "on_join"               # 用户入群
    ON_LEAVE = "on_leave"             # 用户退群
    ON_MESSAGE = "on_message"         # 收到消息
    ON_COMMAND = "on_command"         # 命令触发
    ON_CALLBACK = "on_callback"       # 按钮回调
    ON_VERIFY = "on_verify"           # 验证事件
    ON_EDITED_MESSAGE = "on_edited_message"  # 消息编辑
    ON_CHANNEL_POST = "on_channel_post"      # 频道消息
    ON_INLINE_QUERY = "on_inline_query"      # 内联查询
    ON_CHOSEN_INLINE_RESULT = "on_chosen_inline_result"  # 内联结果选择
    ON_SHIPPING_QUERY = "on_shipping_query"  # 配送查询
    ON_PRE_CHECKOUT_QUERY = "on_pre_checkout_query"  # 预结账查询
    ON_POLL = "on_poll"               # 投票
    ON_POLL_ANSWER = "on_poll_answer"  # 投票答案
    ON_MY_CHAT_MEMBER = "on_my_chat_member"  # Bot 成员状态变化
    ON_CHAT_MEMBER = "on_chat_member"  # 群成员状态变化
    ON_CHAT_JOIN_REQUEST = "on_chat_join_request"  # 入群申请
    ON_ERROR = "on_error"             # 错误处理


class Permission(Enum):
    """权限枚举"""
    READ_MESSAGES = "read_messages"
    SEND_MESSAGES = "send_messages"
    EDIT_MESSAGES = "edit_messages"
    DELETE_MESSAGES = "delete_messages"
    KICK_MEMBERS = "kick_members"
    RESTRICT_MEMBERS = "restrict_members"
    PROMOTE_MEMBERS = "promote_members"
    INVITE_MEMBERS = "invite_members"
    PIN_MESSAGES = "pin_messages"
    MANAGE_CHAT = "manage_chat"
    MANAGE_TOPICS = "manage_topics"


@dataclass
class User:
    """用户信息"""
    id: int
    first_name: str
    last_name: Optional[str] = None
    username: Optional[str] = None
    language_code: Optional[str] = None
    is_bot: bool = False
    is_premium: bool = False

    @property
    def full_name(self) -> str:
        if self.last_name:
            return f"{self.first_name} {self.last_name}"
        return self.first_name

    @property
    def mention(self) -> str:
        if self.username:
            return f"@{self.username}"
        return f"[{self.full_name}](tg://user?id={self.id})"


@dataclass
class Chat:
    """聊天/群组信息"""
    id: int
    type: str  # private, group, supergroup, channel
    title: Optional[str] = None
    username: Optional[str] = None
    first_name: Optional[str] = None
    last_name: Optional[str] = None

    @property
    def is_group(self) -> bool:
        return self.type in ("group", "supergroup")

    @property
    def is_private(self) -> bool:
        return self.type == "private"


@dataclass
class Message:
    """消息信息"""
    id: int
    chat: Chat
    from_user: Optional[User] = None
    text: Optional[str] = None
    caption: Optional[str] = None
    date: int = 0
    reply_to_message: Optional['Message'] = None
    forward_from: Optional[User] = None
    forward_date: Optional[int] = None
    edit_date: Optional[int] = None
    entities: List[Dict] = field(default_factory=list)
    photo: List[Dict] = field(default_factory=list)
    video: Optional[Dict] = None
    audio: Optional[Dict] = None
    document: Optional[Dict] = None
    sticker: Optional[Dict] = None
    animation: Optional[Dict] = None
    voice: Optional[Dict] = None
    video_note: Optional[Dict] = None
    contact: Optional[Dict] = None
    location: Optional[Dict] = None
    venue: Optional[Dict] = None
    poll: Optional[Dict] = None
    dice: Optional[Dict] = None
    new_chat_members: List[User] = field(default_factory=list)
    left_chat_member: Optional[User] = None
    new_chat_title: Optional[str] = None
    delete_chat_photo: bool = False
    group_chat_created: bool = False
    supergroup_chat_created: bool = False
    channel_chat_created: bool = False


@dataclass
class CallbackQuery:
    """回调查询"""
    id: str
    from_user: User
    chat_instance: str
    message: Optional[Message] = None
    inline_message_id: Optional[str] = None
    data: Optional[str] = None
    game_short_name: Optional[str] = None


@dataclass
class Context:
    """插件上下文 - 提供与 Telegram 交互的 API"""

    # 内部属性 (由运行时设置)
    _bot_id: int = 0
    _chat_id: int = 0
    _plugin_id: str = ""
    _db: Any = None
    _redis: Any = None
    _config: Dict = field(default_factory=dict)
    _api_client: Any = None

    @property
    def bot_id(self) -> int:
        return self._bot_id

    @property
    def chat_id(self) -> int:
        return self._chat_id

    @property
    def plugin_id(self) -> str:
        return self._plugin_id

    @property
    def config(self) -> Dict:
        return self._config

    def get_config(self, key: str, default: Any = None) -> Any:
        """获取插件配置项"""
        return self._config.get(key, default)

    async def send_message(
        self,
        text: str,
        parse_mode: Optional[str] = None,
        disable_web_page_preview: bool = False,
        disable_notification: bool = False,
        reply_to_message_id: Optional[int] = None,
        reply_markup: Optional[Dict] = None,
    ) -> Dict:
        """发送消息"""
        if self._api_client:
            return await self._api_client.send_message(
                chat_id=self._chat_id,
                text=text,
                parse_mode=parse_mode,
                disable_web_page_preview=disable_web_page_preview,
                disable_notification=disable_notification,
                reply_to_message_id=reply_to_message_id,
                reply_markup=reply_markup,
            )
        logger.warning("API client not available")
        return {}

    async def reply(
        self,
        text: str,
        parse_mode: Optional[str] = None,
        quote: bool = False,
        **kwargs
    ) -> Dict:
        """回复消息"""
        return await self.send_message(
            text=text,
            parse_mode=parse_mode,
            reply_to_message_id=self._config.get("_message_id") if quote else None,
            **kwargs
        )

    async def delete_message(self, message_id: int) -> bool:
        """删除消息"""
        if self._api_client:
            return await self._api_client.delete_message(
                chat_id=self._chat_id,
                message_id=message_id
            )
        return False

    async def kick_user(
        self,
        user_id: int,
        until_date: Optional[int] = None,
        revoke_messages: bool = False,
    ) -> bool:
        """踢出用户"""
        if self._api_client:
            return await self._api_client.kick_chat_member(
                chat_id=self._chat_id,
                user_id=user_id,
                until_date=until_date,
                revoke_messages=revoke_messages,
            )
        return False

    async def ban_user(self, user_id: int, **kwargs) -> bool:
        """封禁用户"""
        return await self.kick_user(user_id, **kwargs)

    async def unban_user(self, user_id: int, only_if_banned: bool = False) -> bool:
        """解封用户"""
        if self._api_client:
            return await self._api_client.unban_chat_member(
                chat_id=self._chat_id,
                user_id=user_id,
                only_if_banned=only_if_banned,
            )
        return False

    async def restrict_user(
        self,
        user_id: int,
        permissions: Dict,
        until_date: Optional[int] = None,
    ) -> bool:
        """限制用户权限"""
        if self._api_client:
            return await self._api_client.restrict_chat_member(
                chat_id=self._chat_id,
                user_id=user_id,
                permissions=permissions,
                until_date=until_date,
            )
        return False

    async def mute_user(self, user_id: int, duration: int = 0) -> bool:
        """禁言用户"""
        return await self.restrict_user(
            user_id=user_id,
            permissions={
                "can_send_messages": False,
                "can_send_media_messages": False,
                "can_send_polls": False,
                "can_send_other_messages": False,
                "can_add_web_page_previews": False,
            },
            until_date=duration if duration > 0 else None,
        )

    async def unmute_user(self, user_id: int) -> bool:
        """解除禁言"""
        return await self.restrict_user(
            user_id=user_id,
            permissions={
                "can_send_messages": True,
                "can_send_media_messages": True,
                "can_send_polls": True,
                "can_send_other_messages": True,
                "can_add_web_page_previews": True,
                "can_change_info": False,
                "can_invite_users": True,
                "can_pin_messages": False,
            },
        )

    async def get_chat_member(self, user_id: int) -> Optional[Dict]:
        """获取群成员信息"""
        if self._api_client:
            return await self._api_client.get_chat_member(
                chat_id=self._chat_id,
                user_id=user_id,
            )
        return None

    async def get_chat_administrators(self) -> List[Dict]:
        """获取群管理员列表"""
        if self._api_client:
            return await self._api_client.get_chat_administrators(
                chat_id=self._chat_id
            )
        return []

    async def promote_user(
        self,
        user_id: int,
        can_change_info: bool = False,
        can_post_messages: bool = False,
        can_edit_messages: bool = False,
        can_delete_messages: bool = False,
        can_invite_users: bool = False,
        can_restrict_members: bool = False,
        can_pin_messages: bool = False,
        can_promote_members: bool = False,
        can_manage_video_chats: bool = False,
        can_manage_chat: bool = False,
    ) -> bool:
        """提升用户为管理员"""
        if self._api_client:
            return await self._api_client.promote_chat_member(
                chat_id=self._chat_id,
                user_id=user_id,
                can_change_info=can_change_info,
                can_post_messages=can_post_messages,
                can_edit_messages=can_edit_messages,
                can_delete_messages=can_delete_messages,
                can_invite_users=can_invite_users,
                can_restrict_members=can_restrict_members,
                can_pin_messages=can_pin_messages,
                can_promote_members=can_promote_members,
                can_manage_video_chats=can_manage_video_chats,
                can_manage_chat=can_manage_chat,
            )
        return False

    async def answer_callback(
        self,
        callback_query_id: str,
        text: Optional[str] = None,
        show_alert: bool = False,
        url: Optional[str] = None,
        cache_time: int = 0,
    ) -> bool:
        """回答回调查询"""
        if self._api_client:
            return await self._api_client.answer_callback_query(
                callback_query_id=callback_query_id,
                text=text,
                show_alert=show_alert,
                url=url,
                cache_time=cache_time,
            )
        return False

    # 数据库操作
    async def db_query(self, query: str, *args) -> List:
        """执行数据库查询"""
        if self._db:
            return await self._db.fetch(query, *args)
        return []

    async def db_execute(self, query: str, *args) -> str:
        """执行数据库命令"""
        if self._db:
            return await self._db.execute(query, *args)
        return ""

    # Redis 操作
    async def cache_get(self, key: str) -> Optional[str]:
        """获取缓存"""
        if self._redis:
            return await self._redis.get(f"plugin:{self._plugin_id}:{key}")
        return None

    async def cache_set(self, key: str, value: str, expire: int = 3600) -> bool:
        """设置缓存"""
        if self._redis:
            return await self._redis.setex(
                f"plugin:{self._plugin_id}:{key}",
                expire,
                value
            )
        return False

    async def cache_delete(self, key: str) -> bool:
        """删除缓存"""
        if self._redis:
            return await self._redis.delete(f"plugin:{self._plugin_id}:{key}") > 0
        return False


class PluginMeta(type):
    """插件元类，用于收集装饰器注册的处理器"""

    def __new__(mcs, name, bases, namespace):
        cls = super().__new__(mcs, name, bases, namespace)

        # 收集所有通过装饰器注册的处理器
        handlers = {}
        for attr_name, attr_value in namespace.items():
            if hasattr(attr_value, '_hook_type'):
                hook_type = attr_value._hook_type
                if hook_type not in handlers:
                    handlers[hook_type] = []
                handlers[hook_type].append({
                    'handler': attr_value,
                    'name': attr_name,
                    'filters': getattr(attr_value, '_filters', {}),
                    'priority': getattr(attr_value, '_priority', 0),
                })

        cls._handlers = handlers
        return cls


class Plugin(metaclass=PluginMeta):
    """
    插件基类

    所有插件都应该继承此类并实现相应的钩子方法。

    示例:
        class MyPlugin(Plugin):
            id = "my_plugin"
            name = "我的插件"
            version = "1.0.0"
            author = "Developer"

            @Plugin.on_join
            async def on_join(self, ctx: Context, user: User):
                await ctx.send_message(f"欢迎 {user.full_name}!")

            @Plugin.on_command("hello")
            async def hello(self, ctx: Context, args: List[str]):
                await ctx.reply("Hello World!")
    """

    # 插件元信息 (子类必须覆盖)
    id: str = ""
    name: str = ""
    version: str = "1.0.0"
    author: str = ""
    description: str = ""
    permissions: List[Permission] = []

    # 内部属性
    _handlers: Dict[HookType, List[Dict]] = {}

    @staticmethod
    def on_join(priority: int = 0):
        """注册用户入群钩子"""
        def decorator(func):
            func._hook_type = HookType.ON_JOIN
            func._priority = priority
            func._filters = {}
            return func
        return decorator

    @staticmethod
    def on_leave(priority: int = 0):
        """注册用户退群钩子"""
        def decorator(func):
            func._hook_type = HookType.ON_LEAVE
            func._priority = priority
            func._filters = {}
            return func
        return decorator

    @staticmethod
    def on_message(priority: int = 0, **filters):
        """注册消息钩子"""
        def decorator(func):
            func._hook_type = HookType.ON_MESSAGE
            func._priority = priority
            func._filters = filters
            return func
        return decorator

    @staticmethod
    def on_command(command: str, priority: int = 0, **filters):
        """注册命令钩子"""
        def decorator(func):
            func._hook_type = HookType.ON_COMMAND
            func._priority = priority
            func._filters = {'command': command, **filters}
            return func
        return decorator

    @staticmethod
    def on_callback(priority: int = 0, **filters):
        """注册回调钩子"""
        def decorator(func):
            func._hook_type = HookType.ON_CALLBACK
            func._priority = priority
            func._filters = filters
            return func
        return decorator

    @staticmethod
    def on_error(priority: int = 0):
        """注册错误处理钩子"""
        def decorator(func):
            func._hook_type = HookType.ON_ERROR
            func._priority = priority
            func._filters = {}
            return func
        return decorator

    def get_handlers(self) -> Dict[HookType, List[Dict]]:
        """获取所有注册的处理器"""
        return self._handlers

    def get_manifest(self) -> Dict:
        """生成插件清单"""
        hooks = [hook.value for hook in self._handlers.keys()]
        return {
            'id': self.id,
            'name': self.name,
            'version': self.version,
            'author': self.author,
            'description': self.description,
            'hooks': hooks,
            'permissions': [p.value for p in self.permissions],
            'config_schema': getattr(self, 'config_schema', {}),
        }


def command(cmd: str):
    """命令装饰器的简写形式"""
    return Plugin.on_command(cmd)
