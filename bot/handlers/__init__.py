from telegram.ext import Application
from .verification import VerificationHandler
from .member import MemberHandler
from .admin import AdminHandler


def setup_handlers(app: Application, db, redis):
    """Setup all bot handlers"""
    # Initialize handlers
    verification_handler = VerificationHandler(db, redis)
    member_handler = MemberHandler(db, redis)
    admin_handler = AdminHandler(db, redis)

    # Register handlers
    verification_handler.register(app)
    member_handler.register(app)
    admin_handler.register(app)
