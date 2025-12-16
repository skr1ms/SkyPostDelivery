import logging
import sys
from typing import Optional


def setup_logger(name: Optional[str] = None, level: str = "INFO") -> logging.Logger:
    logger = logging.getLogger(name or __name__)

    if logger.handlers:
        return logger

    logger.setLevel(getattr(logging, level.upper()))

    handler = logging.StreamHandler(sys.stdout)
    handler.setLevel(getattr(logging, level.upper()))

    formatter = logging.Formatter(
        "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )
    handler.setFormatter(formatter)

    logger.addHandler(handler)

    return logger
