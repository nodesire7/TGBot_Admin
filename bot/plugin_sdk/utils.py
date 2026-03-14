"""
插件 SDK 工具函数
"""

import re
import hashlib
import time
from typing import List, Optional, Dict, Any, Union
from . import Permission


def parse_command(text: str, prefix: str = "/") -> tuple:
    """
    解析命令文本

    返回: (command, args)
    """
    if not text or not text.startswith(prefix):
        return None, []

    parts = text[len(prefix):].split()
    if not parts:
        return None, []

    return parts[0].lower(), parts[1:]


def extract_entities(text: str, entity_type: str = "mention") -> List[str]:
    """
    从文本中提取实体

    支持类型: mention, hashtag, email, url
    """
    entities = []

    if entity_type == "mention":
        pattern = r'@[\w]+'
    elif entity_type == "hashtag":
        pattern = r'#[\w]+'
    elif entity_type == "email":
        pattern = r'[\w\.-]+@[\w\.-]+\.\w+'
    elif entity_type == "url":
        pattern = r'https?://[^\s]+'
    else:
        return entities

    entities = re.findall(pattern, text)
    return entities


def extract_urls(text: str) -> List[str]:
    """提取文本中的所有 URL"""
    return extract_entities(text, "url")


def extract_mentions(text: str) -> List[str]:
    """提取文本中的所有提及 (@用户名)"""
    return extract_entities(text, "mention")


def extract_hashtags(text: str) -> List[str]:
    """提取文本中的所有标签"""
    return extract_entities(text, "hashtag")


def is_admin(user_id: int, admin_ids: List[int]) -> bool:
    """检查用户是否为管理员"""
    return user_id in admin_ids


def generate_hash(text: str) -> str:
    """生成文本的哈希值"""
    return hashlib.sha256(text.encode()).hexdigest()[:16]


def format_duration(seconds: int) -> str:
    """格式化持续时间"""
    if seconds < 60:
        return f"{seconds}秒"
    elif seconds < 3600:
        return f"{seconds // 60}分钟"
    elif seconds < 86400:
        return f"{seconds // 3600}小时"
    else:
        return f"{seconds // 86400}天"


def format_user_mention(user_id: int, name: Optional[str] = None) -> str:
    """格式化用户提及"""
    if name:
        return f"[{name}](tg://user?id={user_id})"
    return f"[用户](tg://user?id={user_id})"


def escape_markdown(text: str) -> str:
    """转义 Markdown 特殊字符"""
    special_chars = ['_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!']
    for char in special_chars:
        text = text.replace(char, f'\\{char}')
    return text


def truncate_text(text: str, max_length: int = 4000) -> str:
    """截断文本以适应 Telegram 消息长度限制"""
    if len(text) <= max_length:
        return text
    return text[:max_length - 3] + "..."


def chunk_text(text: str, chunk_size: int = 4000) -> List[str]:
    """将长文本分割成多个块"""
    return [text[i:i + chunk_size] for i in range(0, len(text), chunk_size)]


def parse_bool(value: Any) -> bool:
    """解析布尔值"""
    if isinstance(value, bool):
        return value
    if isinstance(value, str):
        return value.lower() in ('true', 'yes', '1', 'on')
    if isinstance(value, int):
        return value != 0
    return False


def check_permissions(required: List[Permission], available: List[Permission]) -> bool:
    """检查权限是否满足"""
    required_set = {p.value for p in required}
    available_set = {p.value for p in available}
    return required_set.issubset(available_set)


class RateLimiter:
    """简单的速率限制器"""

    def __init__(self, max_requests: int = 10, window_seconds: int = 60):
        self.max_requests = max_requests
        self.window_seconds = window_seconds
        self._requests: Dict[int, List[float]] = {}

    def is_allowed(self, user_id: int) -> bool:
        """检查用户是否被允许"""
        now = time.time()
        window_start = now - self.window_seconds

        if user_id not in self._requests:
            self._requests[user_id] = []

        # 清理过期请求
        self._requests[user_id] = [
            ts for ts in self._requests[user_id]
            if ts > window_start
        ]

        if len(self._requests[user_id]) >= self.max_requests:
            return False

        self._requests[user_id].append(now)
        return True

    def remaining(self, user_id: int) -> int:
        """返回剩余请求次数"""
        now = time.time()
        window_start = now - self.window_seconds

        if user_id not in self._requests:
            return self.max_requests

        # 清理过期请求
        self._requests[user_id] = [
            ts for ts in self._requests[user_id]
            if ts > window_start
        ]

        return max(0, self.max_requests - len(self._requests[user_id]))

    def reset(self, user_id: int):
        """重置用户的限制"""
        if user_id in self._requests:
            del self._requests[user_id]


class Cooldown:
    """冷却时间管理器"""

    def __init__(self, default_seconds: int = 30):
        self.default_seconds = default_seconds
        self._cooldowns: Dict[str, Dict[int, float]] = {}  # {key: {user_id: expire_time}}

    def set(self, key: str, user_id: int, seconds: Optional[int] = None):
        """设置冷却时间"""
        if key not in self._cooldowns:
            self._cooldowns[key] = {}

        expire_time = time.time() + (seconds or self.default_seconds)
        self._cooldowns[key][user_id] = expire_time

    def is_on_cooldown(self, key: str, user_id: int) -> bool:
        """检查是否在冷却中"""
        if key not in self._cooldowns:
            return False
        if user_id not in self._cooldowns[key]:
            return False

        return time.time() < self._cooldowns[key][user_id]

    def remaining(self, key: str, user_id: int) -> int:
        """返回剩余冷却时间"""
        if not self.is_on_cooldown(key, user_id):
            return 0

        return int(self._cooldowns[key][user_id] - time.time())

    def reset(self, key: str, user_id: int):
        """重置冷却时间"""
        if key in self._cooldowns and user_id in self._cooldowns[key]:
            del self._cooldowns[key][user_id]


class MessageBuilder:
    """消息构建器"""

    def __init__(self):
        self._text = ""
        self._parse_mode = None
        self._disable_web_page_preview = False
        self._disable_notification = False
        self._reply_to_message_id = None
        self._reply_markup = None

    def text(self, text: str) -> 'MessageBuilder':
        self._text = text
        return self

    def markdown(self) -> 'MessageBuilder':
        self._parse_mode = "MarkdownV2"
        return self

    def html(self) -> 'MessageBuilder':
        self._parse_mode = "HTML"
        return self

    def no_preview(self) -> 'MessageBuilder':
        self._disable_web_page_preview = True
        return self

    def silent(self) -> 'MessageBuilder':
        self._disable_notification = True
        return self

    def reply_to(self, message_id: int) -> 'MessageBuilder':
        self._reply_to_message_id = message_id
        return self

    def keyboard(self, buttons: List[List[Dict]], one_time: bool = False, resize: bool = True) -> 'MessageBuilder':
        self._reply_markup = {
            "type": "ReplyKeyboardMarkup",
            "keyboard": buttons,
            "one_time_keyboard": one_time,
            "resize_keyboard": resize,
        }
        return self

    def inline_keyboard(self, buttons: List[List[Dict]]) -> 'MessageBuilder':
        self._reply_markup = {
            "type": "InlineKeyboardMarkup",
            "inline_keyboard": buttons,
        }
        return self

    def build(self) -> Dict:
        return {
            "text": self._text,
            "parse_mode": self._parse_mode,
            "disable_web_page_preview": self._disable_web_page_preview,
            "disable_notification": self._disable_notification,
            "reply_to_message_id": self._reply_to_message_id,
            "reply_markup": self._reply_markup,
        }


# 常用键盘构建函数
def make_button(text: str, callback_data: Optional[str] = None, url: Optional[str] = None) -> Dict:
    """创建单个按钮"""
    button = {"text": text}
    if callback_data:
        button["callback_data"] = callback_data
    if url:
        button["url"] = url
    return button


def make_inline_keyboard(buttons: List[List[Dict]]) -> Dict:
    """创建内联键盘"""
    return {
        "type": "InlineKeyboardMarkup",
        "inline_keyboard": buttons,
    }


def make_reply_keyboard(buttons: List[List[str]], resize: bool = True) -> Dict:
    """创建回复键盘"""
    keyboard = [[{"text": text} for text in row] for row in buttons]
    return {
        "type": "ReplyKeyboardMarkup",
        "keyboard": keyboard,
        "resize_keyboard": resize,
    }
