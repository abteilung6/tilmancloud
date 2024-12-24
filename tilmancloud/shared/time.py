
import datetime
from datetime import timezone

DEFAULT_TIMEZONE = datetime.timezone.utc

def now(tz: timezone | None = None) -> datetime.datetime:
    return datetime.datetime.now(tz=tz if tz else DEFAULT_TIMEZONE)
