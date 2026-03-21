"""Centralised definitions for available PDF parsing engines."""

from __future__ import annotations

import importlib
from typing import Callable, Dict

import pdf_parser_opendataloader as opendataloader
import pdf_parser_opendataloader_hybrid_docling_fast as opendataloader_hybrid_docling_fast
import pdf_parser_opendataloader_hybrid_hancom as opendataloader_hybrid_hancom

EngineHandler = Callable[..., None]


def _maybe_import(module_name: str):
    """Import an optional engine adapter without breaking the registry."""
    try:
        return importlib.import_module(module_name)
    except ModuleNotFoundError:
        return None


ENGINES: Dict[str, str] = {
    "opendataloader": "local",
    "opendataloader-hybrid-docling-fast": "local+docling-fast",
    "opendataloader-hybrid-hancom": "local+hancom",
}


ENGINE_DISPATCH: Dict[str, EngineHandler] = {
    "opendataloader": opendataloader.to_markdown,
    "opendataloader-hybrid-docling-fast": opendataloader_hybrid_docling_fast.to_markdown,
    "opendataloader-hybrid-hancom": opendataloader_hybrid_hancom.to_markdown,
}

opendataloader_go = _maybe_import("pdf_parser_opendataloader_go")
if opendataloader_go is not None:
    ENGINES["opendataloader-go"] = "local+go"
    ENGINE_DISPATCH["opendataloader-go"] = opendataloader_go.to_markdown
