"""AwqatSalah (Diyanet Namaz Vakti) API — Python istemcisi (sıfır bağımlılık)."""

from .config import Config, load_config
from .client import AwqatClient, unwrap

__all__ = ["Config", "load_config", "AwqatClient", "unwrap"]
