#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.12"
# dependencies = []
# ///
"""Validate a plan bundle and generate its index.md and log.md files.

Usage: uv run plan-bundle.py <validate|index|log> <bundle-dir> [options]

The core takes the files of a bundle and the schemas as values in memory and returns findings and
generated text. Only the functions in the file system edge at the end read directories, read and
write files, and print.
"""

from __future__ import annotations

import bisect
import json
import math
import os
import posixpath
import re
import sys
from collections.abc import Callable, Iterable, Mapping, Sequence
from dataclasses import dataclass, replace
from functools import cmp_to_key
from typing import Literal
from urllib.parse import quote

type YamlValue = None | bool | int | float | str | list[YamlValue] | dict[str, YamlValue]
type Severity = Literal["error", "warning"]
type Path = tuple[str | int, ...]

OKF_VERSION = "0.2"
RESERVED_NAMES = frozenset({"index.md", "log.md"})
UNINDEXED_DIRECTORIES = frozenset({"diagrams"})

DECISION = "Workflow Decision"
WORK_ITEM = "Workflow Work Item"

# The white space of ECMAScript, so trimming and collapsing match the Node validator byte for byte.
JS_WHITESPACE = (
    "\t\n\x0b\x0c\r \xa0           "
    "     　﻿"
)
_WS = (
    r"\t\n\x0b\x0c\r \xa0  -     　﻿"
)
_DOT = r"[^\n\r  ]"
_WHITESPACE_RUN = re.compile(f"[{_WS}]+")


# ---------------------------------------------------------------------------
# Text helpers with ECMAScript semantics
# ---------------------------------------------------------------------------


def js_trim(text: str) -> str:
    return text.strip(JS_WHITESPACE)


def js_trim_end(text: str) -> str:
    return text.rstrip(JS_WHITESPACE)


def char_at(text: str, pos: int) -> str:
    return text[pos] if 0 <= pos < len(text) else ""


def text_key(text: str) -> bytes:
    """UTF-16 code unit order, so sorting never depends on the locale of the machine."""
    return text.encode("utf-16-be", "surrogatepass")


def compare_text(a: str, b: str) -> int:
    x, y = text_key(a), text_key(b)
    return -1 if x < y else (1 if x > y else 0)


def compare_numbers(a: float, b: float) -> int:
    return -1 if a < b else (1 if a > b else 0)


def collapse(text: str) -> str:
    return js_trim(_WHITESPACE_RUN.sub(" ", text))


def is_number(value: object) -> bool:
    return isinstance(value, (int, float)) and not isinstance(value, bool)


def is_js_integer(value: object) -> bool:
    if isinstance(value, bool):
        return False
    if isinstance(value, int):
        return True
    return isinstance(value, float) and value.is_integer()


def _is_array_index(key: str) -> bool:
    return (
        key.isascii()
        and key.isdigit()
        and (key == "0" or not key.startswith("0"))
        and int(key) < 2**32 - 1
    )


def js_keys(mapping: Mapping[str, object]) -> list[str]:
    """Object key order: integer-like keys ascending, then the others in insertion order."""
    keys = list(mapping)
    indices = sorted((key for key in keys if _is_array_index(key)), key=int)
    return indices + [key for key in keys if not _is_array_index(key)]


def js_number(value: float) -> str:
    """Number::toString, so a YAML 1.0 prints as 1 and 1e21 as 1e+21."""
    if isinstance(value, int):
        return str(value)
    if math.isnan(value):
        return "NaN"
    if math.isinf(value):
        return "Infinity" if value > 0 else "-Infinity"
    if value == 0:
        return "0"
    sign = "-" if value < 0 else ""
    mantissa, _, exponent = repr(abs(value)).partition("e")
    whole, _, fraction = mantissa.partition(".")
    digits = whole + fraction
    point = len(whole) + (int(exponent) if exponent else 0)
    stripped = digits.lstrip("0")
    point -= len(digits) - len(stripped)
    digits = stripped.rstrip("0")
    k, n = len(digits), point
    if k <= n <= 21:
        return sign + digits + "0" * (n - k)
    if 0 < n <= 21:
        return f"{sign}{digits[:n]}.{digits[n:]}"
    if -6 < n <= 0:
        return f"{sign}0.{'0' * -n}{digits}"
    e = n - 1
    tail = f"e{'+' if e >= 0 else '-'}{abs(e)}"
    return f"{sign}{digits}{tail}" if k == 1 else f"{sign}{digits[0]}.{digits[1:]}{tail}"


_JSON_ESCAPES = {'"': '\\"', "\\": "\\\\", "\b": "\\b", "\f": "\\f", "\n": "\\n", "\r": "\\r", "\t": "\\t"}
_JSON_ESCAPED = re.compile('["\\\\\x00-\x1f\ud800-\udfff]')


def _json_quote(text: str) -> str:
    return '"' + _JSON_ESCAPED.sub(
        lambda m: _JSON_ESCAPES.get(m.group(0), f"\\u{ord(m.group(0)):04x}"), text
    ) + '"'


def js_stringify(value: object) -> str:
    """JSON.stringify without indentation."""
    if value is None:
        return "null"
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, (int, float)):
        return "null" if isinstance(value, float) and not math.isfinite(value) else js_number(value)
    if isinstance(value, str):
        return _json_quote(value)
    if isinstance(value, list):
        return "[" + ",".join(js_stringify(item) for item in value) + "]"
    if isinstance(value, dict):
        return "{" + ",".join(f"{_json_quote(k)}:{js_stringify(value[k])}" for k in js_keys(value)) + "}"
    return "null"


def js_string(value: object) -> str:
    """String(value)."""
    if isinstance(value, str):
        return value
    if value is None:
        return "null"
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, (int, float)):
        return js_number(value)
    if isinstance(value, list):
        return ",".join("" if item is None else js_string(item) for item in value)
    return "[object Object]"


def js_pattern(source: str) -> str:
    """Translates an ECMAScript pattern to Python: $ ends the input, . stops at line terminators."""
    out: list[str] = []
    in_class = False
    index = 0
    while index < len(source):
        c = source[index]
        if c == "\\":
            escape = source[index : index + 2]
            if escape == "\\s":
                out.append(_WS if in_class else f"[{_WS}]")
            elif escape == "\\S" and not in_class:
                out.append(f"[^{_WS}]")
            else:
                out.append(escape)
            index += 2
            continue
        if in_class:
            in_class = c != "]"
            out.append(c)
        elif c == "[":
            in_class = True
            out.append(c)
        elif c == "$":
            out.append(r"\Z")
        elif c == ".":
            out.append(_DOT)
        elif source.startswith("(?<", index) and not source.startswith(("(?<=", "(?<!"), index):
            out.append("(?P<")
            index += 3
            continue
        else:
            out.append(c)
        index += 1
    return "".join(out)


def compile_js_pattern(source: str) -> re.Pattern[str]:
    return re.compile(js_pattern(source), re.ASCII)


# ---------------------------------------------------------------------------
# Findings
# ---------------------------------------------------------------------------


@dataclass(frozen=True)
class Finding:
    path: str
    line: int
    severity: Severity
    message: str


def error(path: str, line: int, message: str) -> Finding:
    return Finding(path, line, "error", message)


def warning(path: str, line: int, message: str) -> Finding:
    return Finding(path, line, "warning", message)


def _finding_key(finding: Finding) -> tuple[bytes, int, bytes, bytes]:
    return (
        text_key(finding.path),
        finding.line,
        text_key(finding.severity),
        text_key(finding.message),
    )


def finish(findings: Iterable[Finding]) -> tuple[Finding, ...]:
    """Removes duplicates and sorts by path, line, severity and message."""
    return tuple(sorted(dict.fromkeys(findings), key=_finding_key))


# ---------------------------------------------------------------------------
# YAML subset parser
# ---------------------------------------------------------------------------


class YamlError(Exception):
    def __init__(self, message: str, line: int) -> None:
        super().__init__(message)
        self.message = message
        self.line = line


@dataclass(frozen=True)
class YamlOk:
    value: YamlValue
    lines: Mapping[str, int]


@dataclass(frozen=True)
class YamlFail:
    line: int
    message: str


type YamlResult = YamlOk | YamlFail


@dataclass(frozen=True)
class _SourceLine:
    raw: str
    number: int
    indent: int


class _ParserState:
    """The cursor of one parse; it lives only for the duration of parse_yaml."""

    def __init__(self, lines: list[_SourceLine]) -> None:
        self.lines = lines
        self.index = 0
        self.locations: dict[str, int] = {}


@dataclass(frozen=True)
class _Span:
    text: str
    starts: tuple[int, ...]
    numbers: tuple[int, ...]
    owner: int
    locations: dict[str, int]


def _unsupported(name: str, line: int) -> YamlError:
    return YamlError(f"unsupported YAML construct: {name}", line)


def _is_white(c: str) -> bool:
    return c == " " or c == "\t"


def _is_separator(c: str) -> bool:
    return c == "" or _is_white(c)


def _is_flow_separator(c: str) -> bool:
    return c == "" or c == "\n" or _is_white(c) or c in ",[]{}"


def _leading_spaces(raw: str) -> int:
    return len(raw) - len(raw.lstrip(" "))


def _skip_inline(text: str, pos: int) -> int:
    while _is_white(char_at(text, pos)):
        pos += 1
    return pos


def pointer_segment(key: str) -> str:
    return key.replace("~", "~0").replace("/", "~1")


def _locate(locations: dict[str, int], pointer: str, line: int) -> None:
    locations.setdefault(pointer, line)


_INTEGER = re.compile(r"[-+]?[0-9]+|0o[0-7]+|0x[0-9a-fA-F]+")
_FLOAT = re.compile(r"[-+]?(?:\.[0-9]+|[0-9]+(?:\.[0-9]*)?)(?:[eE][-+]?[0-9]+)?")
_INFINITY = re.compile(r"[-+]?\.(?:inf|Inf|INF)")
_NAN = re.compile(r"\.(?:nan|NaN|NAN)")
_MAX_SAFE_INTEGER = 2**53 - 1


def _resolve_plain(text: str, line: int) -> YamlValue:
    """Resolves a plain scalar with the YAML 1.2 core schema, so a timestamp stays a string."""
    if text in ("", "~", "null", "Null", "NULL"):
        return None
    if text in ("true", "True", "TRUE"):
        return True
    if text in ("false", "False", "FALSE"):
        return False
    if _INTEGER.fullmatch(text):
        if text.startswith("0o"):
            value = int(text[2:], 8)
        elif text.startswith("0x"):
            value = int(text[2:], 16)
        else:
            value = int(text)
        if abs(value) > _MAX_SAFE_INTEGER:
            raise YamlError(f"integer {text} is too large to hold exactly", line)
        return value
    if _FLOAT.fullmatch(text):
        return float(text)
    if _INFINITY.fullmatch(text):
        return -math.inf if text.startswith("-") else math.inf
    if _NAN.fullmatch(text):
        return math.nan
    return text


def _reject_indicator(c: str, nxt: str, line: int, separates: Callable[[str], bool]) -> None:
    if c == "&":
        raise _unsupported("anchor", line)
    if c == "*":
        raise _unsupported("alias", line)
    if c == "!":
        raise _unsupported("tag", line)
    if c == "?" and separates(nxt):
        raise _unsupported("complex key", line)
    if c != "" and c in "@`%,]}":
        raise YamlError(f'a plain scalar cannot start with "{c}"', line)


def _line_at(span: _Span, pos: int) -> int:
    return span.numbers[_span_line_index(span, pos)]


def _span_line_index(span: _Span, pos: int) -> int:
    return max(bisect.bisect_right(span.starts, pos) - 1, 0)


def _check_continuation(span: _Span, pos: int) -> None:
    """A closing bracket may sit at the indentation of its parent, as formatters write it."""
    indent = 0
    while char_at(span.text, pos + indent) == " ":
        indent += 1
    content = char_at(span.text, _skip_inline(span.text, pos + indent))
    if content in ("", "\n"):
        return
    if indent == span.owner and content in ("]", "}"):
        return
    if indent <= span.owner:
        raise YamlError(
            "a continuation line must be indented more than its parent", _line_at(span, pos)
        )


def _fold_breaks(span: _Span, pos: int) -> tuple[int, str]:
    breaks = 0
    while char_at(span.text, pos) == "\n":
        breaks += 1
        pos += 1
        _check_continuation(span, pos)
        pos = _skip_inline(span.text, pos)
    return pos, " " if breaks == 1 else "\n" * (breaks - 1)


_SIMPLE_ESCAPES = {
    "0": "\0",
    "a": "\x07",
    "b": "\b",
    "t": "\t",
    "\t": "\t",
    "n": "\n",
    "v": "\v",
    "f": "\f",
    "r": "\r",
    "e": "\x1b",
    " ": " ",
    '"': '"',
    "/": "/",
    "\\": "\\",
    "N": "\x85",
    "_": "\xa0",
    "L": " ",
    "P": " ",
}
_HEX_ESCAPES = {"x": 2, "u": 4, "U": 8}
_HEX_DIGITS = re.compile(r"[0-9a-fA-F]+")


def _append_code_point(out: str, code: int) -> str:
    """Joins an escaped surrogate pair into one character, as a UTF-16 string does."""
    if 0xDC00 <= code <= 0xDFFF and out and 0xD800 <= ord(out[-1]) <= 0xDBFF:
        return out[:-1] + chr(0x10000 + ((ord(out[-1]) - 0xD800) << 10) + (code - 0xDC00))
    return out + chr(code)


def _read_double_quoted(span: _Span, start: int) -> tuple[str, int] | None:
    text = span.text
    pos = start + 1
    out = ""
    kept = 0
    while pos < len(text):
        c = text[pos]
        if c == '"':
            return out, pos + 1
        if c == "\\":
            escape = char_at(text, pos + 1)
            if escape == "\n":
                pos += 2
                _check_continuation(span, pos)
                pos = _skip_inline(text, pos)
                kept = len(out)
                continue
            simple = _SIMPLE_ESCAPES.get(escape)
            width = _HEX_ESCAPES.get(escape, 0)
            digits = text[pos + 2 : pos + 2 + width]
            if simple is not None:
                out += simple
                pos += 2
            elif width > 0 and _HEX_DIGITS.fullmatch(digits) and len(digits) == width:
                code = int(digits, 16)
                if code > 0x10FFFF:
                    raise YamlError(
                        f"escape \\{escape}{digits} is not a code point", _line_at(span, pos)
                    )
                out = _append_code_point(out, code)
                pos += 2 + width
            else:
                raise YamlError(
                    f"invalid escape sequence \\{escape} in a double-quoted scalar",
                    _line_at(span, pos),
                )
            kept = len(out)
            continue
        if c == "\n":
            pos, fold = _fold_breaks(span, pos)
            out = out[:kept] + fold
            kept = len(out)
            continue
        out += c
        if not _is_white(c):
            kept = len(out)
        pos += 1
    return None


def _read_single_quoted(span: _Span, start: int) -> tuple[str, int] | None:
    text = span.text
    pos = start + 1
    out = ""
    kept = 0
    while pos < len(text):
        c = text[pos]
        if c == "'":
            if char_at(text, pos + 1) != "'":
                return out, pos + 1
            out += "'"
            kept = len(out)
            pos += 2
            continue
        if c == "\n":
            pos, fold = _fold_breaks(span, pos)
            out = out[:kept] + fold
            kept = len(out)
            continue
        out += c
        if not _is_white(c):
            kept = len(out)
        pos += 1
    return None


def _read_quoted(span: _Span, start: int) -> tuple[str, int] | None:
    if char_at(span.text, start) == '"':
        return _read_double_quoted(span, start)
    return _read_single_quoted(span, start)


def _quote_kind(c: str) -> str:
    return "double" if c == '"' else "single"


def _skip_flow_space(span: _Span, pos: int) -> int:
    text = span.text
    while True:
        c = char_at(text, pos)
        if _is_white(c):
            pos += 1
        elif c == "\n":
            pos += 1
            _check_continuation(span, pos)
        elif c == "#" and (_is_white(char_at(text, pos - 1)) or char_at(text, pos - 1) == "\n"):
            while pos < len(text) and text[pos] != "\n":
                pos += 1
        else:
            return pos


def _read_flow_plain(span: _Span, start: int) -> tuple[str, int]:
    text = span.text
    pos = start
    out = ""
    kept = 0
    while pos < len(text):
        c = text[pos]
        if c == ":" and _is_flow_separator(char_at(text, pos + 1)):
            break
        if c in ",[]{}":
            break
        if c == "#" and (_is_white(char_at(text, pos - 1)) or char_at(text, pos - 1) == "\n"):
            break
        if c == "\n":
            pos, fold = _fold_breaks(span, pos)
            if char_at(text, pos) == "#":
                break
            out = out[:kept] + fold
            continue
        out += c
        if not _is_white(c):
            kept = len(out)
        pos += 1
    value = out[:kept]
    if value == "":
        raise YamlError("expected a value", _line_at(span, start))
    return value, pos


def _read_flow_node(span: _Span, pos: int, pointer: str) -> tuple[YamlValue, int, bool]:
    text = span.text
    c = char_at(text, pos)
    line = _line_at(span, pos)
    _locate(span.locations, pointer, line)
    if c == "[":
        return _read_flow_sequence(span, pos, pointer)
    if c == "{":
        return _read_flow_mapping(span, pos, pointer)
    if c in ('"', "'"):
        quoted = _read_quoted(span, pos)
        if quoted is None:
            raise YamlError(f"{_quote_kind(c)}-quoted scalar is not closed", line)
        return quoted[0], quoted[1], True
    _reject_indicator(c, char_at(text, pos + 1), line, _is_flow_separator)
    if c in ("|", ">"):
        raise YamlError("a block scalar cannot appear inside a flow collection", line)
    if c == "-" and _is_flow_separator(char_at(text, pos + 1)):
        raise YamlError("a block sequence entry cannot appear inside a flow collection", line)
    plain, end = _read_flow_plain(span, pos)
    return _resolve_plain(plain, line), end, False


def _read_flow_sequence(span: _Span, start: int, pointer: str) -> tuple[YamlValue, int, bool]:
    text = span.text
    items: list[YamlValue] = []
    pos = start + 1
    while True:
        pos = _skip_flow_space(span, pos)
        c = char_at(text, pos)
        if c == "":
            raise YamlError("flow sequence is not closed", _line_at(span, start))
        if c == "]":
            return items, pos + 1, True
        if c == ",":
            raise YamlError("flow sequence has an empty entry", _line_at(span, pos))
        value, end, is_json = _read_flow_node(span, pos, f"{pointer}/{len(items)}")
        pos = _skip_flow_space(span, end)
        if char_at(text, pos) == ":" and (is_json or _is_flow_separator(char_at(text, pos + 1))):
            raise _unsupported("flow pair", _line_at(span, pos))
        items.append(value)
        nxt = char_at(text, pos)
        if nxt == ",":
            pos += 1
        elif nxt != "]":
            raise YamlError(
                "flow sequence is not closed" if nxt == "" else "expected , or ] in a flow sequence",
                _line_at(span, start if nxt == "" else pos),
            )


def _read_flow_key(span: _Span, pos: int) -> tuple[str, int, bool]:
    text = span.text
    c = char_at(text, pos)
    line = _line_at(span, pos)
    if c in ("[", "{"):
        raise _unsupported("complex key", line)
    if c in ('"', "'"):
        quoted = _read_quoted(span, pos)
        if quoted is None:
            raise YamlError(f"{_quote_kind(c)}-quoted scalar is not closed", line)
        return quoted[0], quoted[1], True
    _reject_indicator(c, char_at(text, pos + 1), line, _is_flow_separator)
    if c == ":":
        raise YamlError("a mapping entry has no key", line)
    plain, end = _read_flow_plain(span, pos)
    return plain, end, False


def _read_flow_mapping(span: _Span, start: int, pointer: str) -> tuple[YamlValue, int, bool]:
    text = span.text
    mapping: dict[str, YamlValue] = {}
    pos = start + 1
    while True:
        pos = _skip_flow_space(span, pos)
        c = char_at(text, pos)
        if c == "":
            raise YamlError("flow mapping is not closed", _line_at(span, start))
        if c == "}":
            return mapping, pos + 1, True
        if c == ",":
            raise YamlError("flow mapping has an empty entry", _line_at(span, pos))
        key_line = _line_at(span, pos)
        key, end, key_json = _read_flow_key(span, pos)
        if key in mapping:
            raise YamlError(f'duplicate key "{key}"', key_line)
        entry_pointer = f"{pointer}/{pointer_segment(key)}"
        _locate(span.locations, entry_pointer, key_line)
        pos = _skip_flow_space(span, end)
        value: YamlValue = None
        if char_at(text, pos) == ":" and (key_json or _is_flow_separator(char_at(text, pos + 1))):
            pos = _skip_flow_space(span, pos + 1)
            if char_at(text, pos) not in (",", "}"):
                value, end, _ = _read_flow_node(span, pos, entry_pointer)
                pos = _skip_flow_space(span, end)
        mapping[key] = value
        nxt = char_at(text, pos)
        if nxt == ",":
            pos += 1
        elif nxt != "}":
            raise YamlError(
                "flow mapping is not closed" if nxt == "" else "expected , or } in a flow mapping",
                _line_at(span, start if nxt == "" else pos),
            )


def _span_from(state: _ParserState, owner: int) -> _Span:
    rest = state.lines[state.index :]
    starts: list[int] = []
    offset = 0
    for line in rest:
        starts.append(offset)
        offset += len(line.raw) + 1
    return _Span(
        "\n".join(line.raw for line in rest),
        tuple(starts),
        tuple(line.number for line in rest),
        owner,
        state.locations,
    )


def _finish_inline(state: _ParserState, span: _Span, end: int) -> None:
    pos = _skip_inline(span.text, end)
    c = char_at(span.text, pos)
    if c not in ("", "\n") and not (c == "#" and pos > end):
        raise YamlError("unexpected text after the value", _line_at(span, pos))
    state.index += _span_line_index(span, pos) + 1


def _closing_bracket(raw: str, start: int) -> int:
    depth = 0
    i = start
    while i < len(raw):
        c = raw[i]
        if c in ('"', "'"):
            i += 1
            while i < len(raw) and raw[i] != c:
                i += 2 if c == '"' and raw[i] == "\\" else 1
        elif c in ("[", "{"):
            depth += 1
        elif c in ("]", "}"):
            depth -= 1
            if depth == 0:
                return i
        i += 1
    return -1


def _mapping_key(raw: str, start: int, number: int) -> tuple[str, int] | None:
    """Returns the key and the column after its colon, or None when the line is no mapping entry."""
    c = char_at(raw, start)
    if c == "?" and _is_separator(char_at(raw, start + 1)):
        raise _unsupported("complex key", number)
    if c == "&":
        raise _unsupported("anchor", number)
    if c == "*":
        raise _unsupported("alias", number)
    if c == "!":
        raise _unsupported("tag", number)
    if c in ("[", "{"):
        close = _closing_bracket(raw, start)
        after = -1 if close < 0 else _skip_inline(raw, close + 1)
        if after >= 0 and char_at(raw, after) == ":" and _is_separator(char_at(raw, after + 1)):
            raise _unsupported("complex key", number)
        return None
    if c in ('"', "'"):
        span = _Span(raw, (0,), (number,), -1, {})
        quoted = _read_quoted(span, start)
        after = -1 if quoted is None else _skip_inline(raw, quoted[1])
        if quoted is not None and char_at(raw, after) == ":" and _is_separator(char_at(raw, after + 1)):
            return quoted[0], after + 1
        return None
    for i in range(start, len(raw)):
        ch = raw[i]
        if ch == "#" and i > start and _is_white(raw[i - 1]):
            return None
        if ch == ":" and _is_separator(char_at(raw, i + 1)):
            if i == start:
                raise YamlError("a mapping entry has no key", number)
            if c in "|>%@`,]}":
                raise YamlError(f'a plain scalar cannot start with "{c}"', number)
            return js_trim_end(raw[start:i]), i + 1
    return None


def _is_sequence_entry(line: _SourceLine) -> bool:
    return char_at(line.raw, line.indent) == "-" and _is_separator(char_at(line.raw, line.indent + 1))


def _peek(state: _ParserState) -> _SourceLine | None:
    """Returns the next line that is not blank or a comment."""
    while True:
        if state.index >= len(state.lines):
            return None
        line = state.lines[state.index]
        content = js_trim(line.raw[line.indent :])
        if content != "" and not content.startswith("#"):
            if char_at(line.raw, line.indent) == "\t":
                raise YamlError("tab character in indentation", line.number)
            return line
        state.index += 1


def _scan_plain_line(raw: str, start: int, number: int) -> tuple[str, bool]:
    for i in range(start, len(raw)):
        c = raw[i]
        if c == "#" and i > start and _is_white(raw[i - 1]):
            return js_trim_end(raw[start:i]), True
        if c == ":" and _is_separator(char_at(raw, i + 1)):
            raise YamlError("a mapping value is not allowed here", number)
    return js_trim_end(raw[start:]), False


def _parse_plain_block(state: _ParserState, line: _SourceLine, start: int, owner: int) -> YamlValue:
    text, comment = _scan_plain_line(line.raw, start, line.number)
    state.index += 1
    if comment:
        return _resolve_plain(text, line.number)
    blank = 0
    for index in range(state.index, len(state.lines)):
        nxt = state.lines[index]
        trimmed = js_trim(nxt.raw)
        if trimmed == "":
            blank += 1
            continue
        if nxt.indent <= owner or trimmed.startswith("#"):
            break
        segment, segment_comment = _scan_plain_line(
            nxt.raw, _skip_inline(nxt.raw, nxt.indent), nxt.number
        )
        text += f" {segment}" if blank == 0 else "\n" * blank + segment
        blank = 0
        state.index = index + 1
        if segment_comment:
            break
    return _resolve_plain(text, line.number)


def _fold_block_lines(content: Sequence[tuple[str, bool]]) -> str:
    out = ""
    pending_empty = 0
    started = False
    previous_spaced = False
    for text, empty in content:
        if empty:
            pending_empty += 1
            continue
        spaced = text.startswith(" ") or text.startswith("\t")
        if not started:
            out += "\n" * pending_empty
        elif spaced or previous_spaced:
            out += "\n" * (pending_empty + 1)
        else:
            out += " " if pending_empty == 0 else "\n" * pending_empty
        out += text
        started = True
        previous_spaced = spaced
        pending_empty = 0
    return out


_ALL_SPACES = re.compile(" *")


def _parse_block_scalar(state: _ParserState, line: _SourceLine, start: int, owner: int) -> str:
    """Parses a | or > scalar; owner is the indentation of the key or entry that holds it."""
    raw = line.raw
    literal = char_at(raw, start) == "|"
    pos = start + 1
    chomping = "clip"
    explicit = 0
    for _ in range(2):
        c = char_at(raw, pos)
        if c in ("-", "+") and chomping == "clip":
            chomping = "strip" if c == "-" else "keep"
            pos += 1
        elif c != "" and "1" <= c <= "9" and explicit == 0:
            explicit = int(c)
            pos += 1
    after = _skip_inline(raw, pos)
    if char_at(raw, after) != "" and not (char_at(raw, after) == "#" and after > pos):
        raise YamlError("invalid block scalar header", line.number)
    state.index += 1
    indent = owner + explicit if explicit > 0 else -1
    if indent < 0:
        longest_empty = 0
        longest_line = line.number
        for index in range(state.index, len(state.lines)):
            nxt = state.lines[index].raw
            if _ALL_SPACES.fullmatch(nxt):
                if len(nxt) > longest_empty:
                    longest_line = state.lines[index].number
                longest_empty = max(longest_empty, len(nxt))
                continue
            spaces = _leading_spaces(nxt)
            if spaces > owner:
                indent = spaces
            if spaces > owner and longest_empty > spaces:
                raise YamlError(
                    "a leading empty line of a block scalar is indented more than its content",
                    longest_line,
                )
            break
        if indent < 0:
            indent = max(owner + 1, longest_empty)
    body: list[tuple[str, bool]] = []
    while state.index < len(state.lines):
        nxt = state.lines[state.index].raw
        if _ALL_SPACES.fullmatch(nxt) and len(nxt) <= indent:
            body.append(("", True))
        elif _leading_spaces(nxt) >= indent:
            body.append((nxt[indent:], False))
        else:
            break
        state.index += 1
    last = len(body)
    while last > 0 and body[last - 1][1]:
        last -= 1
    content = body[:last]
    trailing = len(body) - last
    value = "\n".join(text for text, _ in content) if literal else _fold_block_lines(content)
    if chomping == "strip":
        return value
    if not content:
        return "\n" * trailing if chomping == "keep" else value
    return value + "\n" * (trailing + 1) if chomping == "keep" else value + "\n"


def _parse_scalar_node(
    state: _ParserState, line: _SourceLine, start: int, owner: int, pointer: str
) -> YamlValue:
    _locate(state.locations, pointer, line.number)
    c = char_at(line.raw, start)
    if c in ("|", ">"):
        return _parse_block_scalar(state, line, start, owner)
    if c in ("[", "{", '"', "'"):
        span = _span_from(state, owner)
        value, end, _ = _read_flow_node(span, start, pointer)
        _finish_inline(state, span, end)
        return value
    _reject_indicator(c, char_at(line.raw, start + 1), line.number, _is_separator)
    return _parse_plain_block(state, line, start, owner)


def _parse_nested(
    state: _ParserState, owner: int, pointer: str, same_indent_sequence: bool
) -> YamlValue:
    """same_indent_sequence: whether a sequence at the owner's indentation belongs to it."""
    nxt = _peek(state)
    if nxt is None:
        return None
    if nxt.indent > owner:
        return _parse_block_node_at(state, nxt, owner, pointer)
    if same_indent_sequence and nxt.indent == owner and _is_sequence_entry(nxt):
        return _parse_block_sequence(state, owner, pointer)
    return None


def _parse_block_node_at(
    state: _ParserState, line: _SourceLine, owner: int, pointer: str
) -> YamlValue:
    if _is_sequence_entry(line):
        return _parse_block_sequence(state, line.indent, pointer)
    if _mapping_key(line.raw, line.indent, line.number) is not None:
        return _parse_block_mapping(state, line.indent, pointer)
    return _parse_scalar_node(state, line, line.indent, owner, pointer)


def _parse_block_mapping(state: _ParserState, indent: int, pointer: str) -> dict[str, YamlValue]:
    mapping: dict[str, YamlValue] = {}
    while True:
        line = _peek(state)
        if line is None or line.indent < indent:
            return mapping
        if line.indent > indent:
            raise YamlError("unexpected indentation", line.number)
        if _is_sequence_entry(line):
            raise YamlError("expected a mapping entry, found a sequence entry", line.number)
        entry = _mapping_key(line.raw, indent, line.number)
        if entry is None:
            raise YamlError("expected a mapping entry", line.number)
        key, value_column = entry
        if key in mapping:
            raise YamlError(f'duplicate key "{key}"', line.number)
        entry_pointer = f"{pointer}/{pointer_segment(key)}"
        _locate(state.locations, entry_pointer, line.number)
        start = _skip_inline(line.raw, value_column)
        c = char_at(line.raw, start)
        if c in ("", "#"):
            state.index += 1
            value = _parse_nested(state, indent, entry_pointer, True)
        elif c == "-" and _is_separator(char_at(line.raw, start + 1)):
            raise YamlError("a block sequence cannot start on the line of its key", line.number)
        else:
            value = _parse_scalar_node(state, line, start, indent, entry_pointer)
        mapping[key] = value


def _parse_block_sequence(state: _ParserState, indent: int, pointer: str) -> list[YamlValue]:
    items: list[YamlValue] = []
    while True:
        line = _peek(state)
        if line is None or line.indent < indent:
            return items
        if line.indent > indent:
            raise YamlError("unexpected indentation", line.number)
        if not _is_sequence_entry(line):
            return items
        item_pointer = f"{pointer}/{len(items)}"
        _locate(state.locations, item_pointer, line.number)
        start = _skip_inline(line.raw, indent + 1)
        c = char_at(line.raw, start)
        # A compact entry such as "- key: value" continues on the next lines at the column after the dash.
        compact = _SourceLine(line.raw, line.number, start)
        if c in ("", "#"):
            state.index += 1
            items.append(_parse_nested(state, indent, item_pointer, False))
        elif c == "-" and _is_separator(char_at(line.raw, start + 1)):
            state.lines[state.index] = compact
            items.append(_parse_block_sequence(state, start, item_pointer))
        elif _mapping_key(line.raw, start, line.number) is not None:
            state.lines[state.index] = compact
            items.append(_parse_block_mapping(state, start, item_pointer))
        else:
            items.append(_parse_scalar_node(state, line, start, indent, item_pointer))


_DOCUMENT_MARKER = re.compile(r"(?:---|\.\.\.)(?:[ \t]|\Z)")
_LINE_BREAK = re.compile(r"\r\n?")


def parse_yaml(text: str, first_line: int = 1) -> YamlResult:
    """Parses one YAML document in the strict subset a plan bundle uses.

    A construct outside the subset is an error with its line, never a best guess. first_line is the
    line number of the first line of text in its file.
    """
    raws = _LINE_BREAK.sub("\n", text.removeprefix("﻿")).split("\n")
    if raws[-1] == "":
        raws.pop()
    state = _ParserState(
        [_SourceLine(raw, first_line + index, _leading_spaces(raw)) for index, raw in enumerate(raws)]
    )
    try:
        for line in state.lines:
            if _DOCUMENT_MARKER.match(line.raw):
                raise _unsupported("multi-document stream", line.number)
            if line.raw.startswith("%"):
                raise _unsupported("directive", line.number)
        first = _peek(state)
        if first is None:
            return YamlOk(None, state.locations)
        _locate(state.locations, "", first.number)
        value = _parse_block_node_at(state, first, -1, "")
        rest = _peek(state)
        if rest is not None:
            raise YamlError("unexpected content after the document", rest.number)
        return YamlOk(value, state.locations)
    except YamlError as caught:
        return YamlFail(caught.line, caught.message)


# ---------------------------------------------------------------------------
# Dates
# ---------------------------------------------------------------------------

_DATETIME_OFFSET = re.compile(
    r"([0-9]{4})-([0-9]{2})-([0-9]{2})T([0-9]{2}):([0-9]{2})(?::([0-9]{2})(?:\.([0-9]+))?)?"
    r"(?:Z|([+-])([0-9]{2}):([0-9]{2}))"
)
_CALENDAR_DATE = re.compile(r"([0-9]{4})-([0-9]{2})-([0-9]{2})")


def _is_leap(year: int) -> bool:
    return year % 4 == 0 and (year % 100 != 0 or year % 400 == 0)


def _days_in_month(year: int, month: int) -> int:
    # Date.UTC reads the years 0 to 99 as 1900 to 1999, and the Node validator inherits that.
    full_year = year + 1900 if 0 <= year <= 99 else year
    if month == 2:
        return 29 if _is_leap(full_year) else 28
    return 30 if month in (4, 6, 9, 11) else 31


def _is_valid_date(year: int, month: int, day: int) -> bool:
    return 1 <= month <= 12 and 1 <= day <= _days_in_month(year, month)


def is_calendar_date(text: str) -> bool:
    match = _CALENDAR_DATE.fullmatch(text)
    return match is not None and _is_valid_date(*(int(group) for group in match.groups()))


def is_datetime_offset(value: object) -> bool:
    if not isinstance(value, str):
        return False
    match = _DATETIME_OFFSET.fullmatch(value)
    if match is None:
        return False
    year, month, day, hour, minute, second, _, _, offset_hour, offset_minute = match.groups()
    return (
        _is_valid_date(int(year), int(month), int(day))
        and int(hour) <= 23
        and int(minute) <= 59
        and (second is None or int(second) <= 59)
        and (offset_hour is None or int(offset_hour) <= 23)
        and (offset_minute is None or int(offset_minute) <= 59)
    )


def _days_from_civil(year: int, month: int, day: int) -> int:
    year -= month <= 2
    era = (year if year >= 0 else year - 399) // 400
    year_of_era = year - era * 400
    day_of_year = (153 * (month + (-3 if month > 2 else 9)) + 2) // 5 + day - 1
    day_of_era = year_of_era * 365 + year_of_era // 4 - year_of_era // 100 + day_of_year
    return era * 146097 + day_of_era - 719468


def instant_ms(value: str) -> int | None:
    """Milliseconds since the epoch of a datetime with an offset, as Date.parse reads it."""
    if not is_datetime_offset(value):
        return None
    match = _DATETIME_OFFSET.fullmatch(value)
    assert match is not None
    year, month, day, hour, minute, second, fraction, sign, offset_hour, offset_minute = match.groups()
    days = _days_from_civil(int(year), int(month), int(day))
    milliseconds = int(((fraction or "") + "000")[:3])
    local = (
        ((days * 24 + int(hour)) * 60 + int(minute)) * 60 + int(second or 0)
    ) * 1000 + milliseconds
    offset = 0 if sign is None else (int(offset_hour) * 60 + int(offset_minute)) * 60000
    return local - offset if sign != "-" else local + offset


# ---------------------------------------------------------------------------
# Bundle model
# ---------------------------------------------------------------------------


@dataclass(frozen=True)
class FrontmatterBlock:
    kind: Literal["absent", "unclosed", "present"]
    yaml: str = ""
    first_line: int = 0
    body: str = ""
    body_line: int = 0


@dataclass(frozen=True)
class BundleFile:
    path: str
    id: str
    directory: str
    name: str
    reserved: bool
    text: str
    block: FrontmatterBlock
    yaml: YamlResult | None
    data: dict[str, YamlValue] | None
    lines: Mapping[str, int]
    body: str
    body_line: int


@dataclass(frozen=True)
class BundleSource:
    """The files of a bundle as values: bundle-relative paths with / and the text of each file.

    files holds every .md regular file, links every symbolic link, and directories every directory;
    the parents of files and links count as directories without being listed.
    """

    files: tuple[tuple[str, str], ...]
    directories: tuple[str, ...] = ()
    links: tuple[str, ...] = ()


@dataclass(frozen=True)
class Bundle:
    directories: tuple[str, ...]
    files: tuple[BundleFile, ...]
    concepts: Mapping[str, BundleFile]
    findings: tuple[Finding, ...]


def split_frontmatter(text: str) -> FrontmatterBlock:
    lines = _LINE_BREAK.sub("\n", text.removeprefix("﻿")).split("\n")
    if js_trim_end(lines[0]) != "---":
        return FrontmatterBlock("absent")
    close = next((i for i in range(1, len(lines)) if js_trim_end(lines[i]) == "---"), -1)
    if close < 0:
        return FrontmatterBlock("unclosed")
    return FrontmatterBlock(
        "present",
        "".join(f"{line}\n" for line in lines[1:close]),
        2,
        "\n".join(lines[close + 1 :]),
        close + 2,
    )


def read_bundle_file(path: str, text: str) -> BundleFile:
    slash = path.rfind("/")
    name = path[slash + 1 :]
    block = split_frontmatter(text)
    yaml = parse_yaml(block.yaml, block.first_line) if block.kind == "present" else None
    data = yaml.value if isinstance(yaml, YamlOk) and isinstance(yaml.value, dict) else None
    return BundleFile(
        path=path,
        id=path[:-3],
        directory="" if slash < 0 else path[:slash],
        name=name,
        reserved=name in RESERVED_NAMES,
        text=text,
        block=block,
        yaml=yaml,
        data=data,
        lines=yaml.lines if isinstance(yaml, YamlOk) else {},
        body=block.body if block.kind == "present" else _LINE_BREAK.sub("\n", text),
        body_line=block.body_line if block.kind == "present" else 1,
    )


def _walk_key(path: str) -> tuple[bytes, ...]:
    """Sorting by path segments gives the order of a depth-first walk with sorted entries."""
    return tuple(text_key(segment) for segment in path.split("/")) if path else ()


def _ancestors(path: str) -> Iterable[str]:
    parts = path.split("/")[:-1]
    return ("/".join(parts[:count]) for count in range(1, len(parts) + 1))


def load_bundle(source: BundleSource) -> Bundle:
    directories = {"", *source.directories}
    for path in (*(path for path, _ in source.files), *source.links):
        directories.update(_ancestors(path))
    files = tuple(
        read_bundle_file(path, text)
        for path, text in sorted(source.files, key=lambda item: _walk_key(item[0]))
    )
    return Bundle(
        directories=tuple(sorted(directories, key=_walk_key)),
        files=files,
        concepts={file.id: file for file in files if not file.reserved},
        findings=tuple(warning(path, 1, "symbolic link is not followed") for path in source.links),
    )


def line_of(file: BundleFile, pointer: str) -> int:
    current = pointer
    while True:
        line = file.lines.get(current)
        if line is not None:
            return line
        if current == "":
            return 1
        current = current[: current.rfind("/")]


def type_of(file: BundleFile) -> str:
    value = None if file.data is None else file.data.get("type")
    return js_trim(value) if isinstance(value, str) else ""


def is_workflow(file: BundleFile) -> bool:
    return type_of(file).startswith("Workflow ")


# ---------------------------------------------------------------------------
# Layer 1: Open Knowledge Format conformance
# ---------------------------------------------------------------------------


def _frontmatter_problem(file: BundleFile) -> Finding | None:
    if file.block.kind == "absent":
        return error(file.path, 1, "no frontmatter: a concept document starts with a --- line")
    if file.block.kind == "unclosed":
        return error(file.path, 1, "frontmatter is not closed by a --- line")
    if isinstance(file.yaml, YamlFail):
        return error(file.path, file.yaml.line, file.yaml.message)
    if file.data is None:
        return error(file.path, 1, "frontmatter must be a YAML mapping")
    return None


def concept_findings(file: BundleFile) -> tuple[Finding, ...]:
    problem = _frontmatter_problem(file)
    if problem is not None:
        return (problem,)
    assert file.data is not None
    value = file.data.get("type")
    if value is None:
        return (error(file.path, 1, "frontmatter has no type"),)
    if not isinstance(value, str):
        return (error(file.path, line_of(file, "/type"), "type must be a string"),)
    if js_trim(value) == "":
        return (error(file.path, line_of(file, "/type"), "type must not be empty"),)
    return ()


_HEADING = re.compile(f"#{{1,6}} [^{_WS}]")
_INDEX_ENTRY = re.compile(
    rf"[*-] \[(?:[^\]\\]|\\{_DOT})+\]\([^){_WS}]+\)(?: - [^{_WS}]{_DOT}*)?"
)
_DATE_HEADING = re.compile(f"## ({_DOT}*[^{_WS}])")
_LOG_TITLE = re.compile(f"# [^{_WS}]")
_LOG_ENTRY = re.compile(f"[*-] [^{_WS}]")


def index_structure(file: BundleFile) -> tuple[Finding, ...]:
    findings: list[Finding] = []
    headings = 0
    for index, text in enumerate(file.body.split("\n")):
        line = file.body_line + index
        if js_trim(text) == "":
            continue
        if _HEADING.match(text):
            headings += 1
        elif not _INDEX_ENTRY.fullmatch(text):
            findings.append(
                error(
                    file.path,
                    line,
                    "an index line is a # heading or an entry * [Title](url) - description",
                )
            )
        elif headings == 0:
            findings.append(error(file.path, line, "an index entry must follow a # heading"))
    if headings == 0:
        findings.append(error(file.path, file.body_line, "an index needs a # heading"))
    return tuple(findings)


def log_structure(file: BundleFile) -> tuple[Finding, ...]:
    findings: list[Finding] = []
    titled = False
    dated = False
    previous = ""
    for index, text in enumerate(file.body.split("\n")):
        line = file.body_line + index
        if js_trim(text) == "":
            continue
        if not titled:
            titled = True
            if _LOG_TITLE.match(text):
                continue
            findings.append(
                error(file.path, line, "a log starts with a # heading: # Directory Update Log")
            )
        match = _DATE_HEADING.fullmatch(text)
        if match is not None:
            date = match.group(1)
            if not is_calendar_date(date):
                findings.append(error(file.path, line, f"date heading {date} is not a YYYY-MM-DD date"))
            elif previous != "" and compare_text(date, previous) >= 0:
                findings.append(
                    error(
                        file.path, line, f"date headings run newest first, but {date} follows {previous}"
                    )
                )
            previous = date
            dated = True
        elif not _LOG_ENTRY.match(text):
            findings.append(
                error(file.path, line, "a log line is a ## YYYY-MM-DD heading or an entry starting with *")
            )
        elif not dated:
            findings.append(error(file.path, line, "a log entry must follow a ## YYYY-MM-DD heading"))
    if not titled:
        findings.append(
            error(file.path, file.body_line, "a log starts with a # heading: # Directory Update Log")
        )
    return tuple(findings)


def reserved_findings(file: BundleFile) -> tuple[Finding, ...]:
    if file.block.kind == "unclosed":
        return (error(file.path, 1, "frontmatter is not closed by a --- line"),)
    if isinstance(file.yaml, YamlFail):
        return (error(file.path, file.yaml.line, file.yaml.message),)
    findings: list[Finding] = []
    value = file.yaml.value if isinstance(file.yaml, YamlOk) else None
    if file.block.kind == "present":
        if isinstance(value, dict) and "type" in value:
            findings.append(
                error(
                    file.path,
                    line_of(file, "/type"),
                    f"{file.name} is a reserved file name and cannot be a concept document",
                )
            )
        elif file.path != "index.md":
            findings.append(
                error(file.path, 1, f"only the bundle root index.md has frontmatter, {file.path} has none")
            )
        elif not isinstance(value, dict):
            findings.append(
                error(file.path, 1, 'the root index.md frontmatter holds only okf_version: "0.2"')
            )
        else:
            for key in js_keys(value):
                if key != "okf_version":
                    findings.append(
                        error(
                            file.path,
                            line_of(file, f"/{pointer_segment(key)}"),
                            f"the root index.md frontmatter holds only okf_version, not {key}",
                        )
                    )
                elif not isinstance(value[key], str):
                    findings.append(
                        error(
                            file.path,
                            line_of(file, "/okf_version"),
                            'okf_version is a quoted string such as "0.2"',
                        )
                    )
    structure = index_structure(file) if file.name == "index.md" else log_structure(file)
    return (*findings, *structure)


def okf_findings(bundle: Bundle) -> tuple[Finding, ...]:
    return tuple(
        finding
        for file in bundle.files
        for finding in (reserved_findings(file) if file.reserved else concept_findings(file))
    )


# ---------------------------------------------------------------------------
# Layer 2: schemas
# ---------------------------------------------------------------------------

SCHEMA_KEYWORDS = frozenset(
    {
        "type",
        "required",
        "properties",
        "items",
        "enum",
        "pattern",
        "minLength",
        "minItems",
        "uniqueItems",
        "oneOf",
        "allOf",
        "$ref",
        "const",
        "format",
        "x-concept",
        "additionalProperties",
    }
)
SCHEMA_ANNOTATIONS = frozenset(
    {"$schema", "$id", "$comment", "$defs", "title", "description", "examples", "default"}
)
SCHEMA_TYPES = ("string", "integer", "number", "boolean", "array", "object", "null")
_REF_PATTERN = re.compile(r"(?P<target>[A-Za-z0-9][A-Za-z0-9._/-]*)?#/\$defs/(?P<name>[^/~]+)")
_SCHEMA_NAME = re.compile(r"workflow/[a-z0-9]+(?:-[a-z0-9]+)*/[0-9]+")
_CONCEPT_PREFIX = re.compile(r"(?:[a-z0-9][a-z0-9-]*/)+")


@dataclass(frozen=True)
class SchemaSource:
    """A *.schema.json file as a value: its absolute path, its text and its tier.

    Tier 0 holds the --schemas directories and tier 1 the defaults; tier 0 wins on a shared $id.
    """

    file: str
    text: str
    tier: int


@dataclass(frozen=True)
class SchemaFile:
    file: str
    document: object
    lines: Mapping[str, int]
    id: str | None
    refs: tuple[str, ...]
    tier: int


@dataclass(frozen=True)
class SchemaRegistry:
    files: Mapping[str, SchemaFile]
    by_id: Mapping[str, SchemaFile]
    broken: frozenset[str]
    findings: tuple[Finding, ...]


EMPTY_REGISTRY = SchemaRegistry({}, {}, frozenset(), ())


@dataclass(frozen=True)
class SchemaIssue:
    path: Path
    message: str


@dataclass(frozen=True)
class SchemaContext:
    registry: SchemaRegistry
    concepts: frozenset[str]


def json_lines(text: str) -> dict[str, int]:
    """Maps each JSON pointer of a valid JSON text to the line it starts on."""
    lines: dict[str, int] = {}
    pos = 0
    line = 1

    def space() -> None:
        nonlocal pos, line
        while pos < len(text) and text[pos] in " \t\n\r":
            if text[pos] == "\n":
                line += 1
            pos += 1

    def string() -> str:
        nonlocal pos
        start = pos
        pos += 1
        while pos < len(text) and text[pos] != '"':
            pos += 2 if text[pos] == "\\" else 1
        pos += 1
        return js_string(json.loads(text[start:pos]))

    def value(pointer: str) -> None:
        nonlocal pos
        space()
        _locate(lines, pointer, line)
        c = char_at(text, pos)
        if c in ("{", "["):
            pos += 1
            space()
            more = char_at(text, pos) != ("}" if c == "{" else "]")
            if not more:
                pos += 1
            index = 0
            while more:
                space()
                if c == "{":
                    key_line = line
                    child = f"{pointer}/{pointer_segment(string())}"
                    _locate(lines, child, key_line)
                    space()
                    pos += 1
                    value(child)
                else:
                    value(f"{pointer}/{index}")
                space()
                more = char_at(text, pos) == ","
                pos += 1
                index += 1
        elif c == '"':
            string()
        else:
            while pos < len(text) and text[pos] not in " \t\n\r,]}":
                pos += 1

    value("")
    return lines


def _schema_line(schema: SchemaFile, pointer: str) -> int:
    return schema.lines.get(pointer, 1)


def _check_schema_node(
    schema: SchemaFile, node: object, pointer: str, findings: list[Finding], refs: list[str]
) -> None:
    def report(at: str, message: str) -> None:
        findings.append(error(schema.file, _schema_line(schema, at), message))

    if not isinstance(node, dict):
        report(pointer, "a schema must be a JSON object")
        return

    def subschemas(value: object, at: str) -> None:
        if not isinstance(value, list) or len(value) == 0:
            report(at, "must be a non-empty array of schemas")
        else:
            for index, item in enumerate(value):
                _check_schema_node(schema, item, f"{at}/{index}", findings, refs)

    def schema_map(value: object, at: str) -> None:
        if not isinstance(value, dict):
            report(at, "must be an object of schemas")
        else:
            for key in js_keys(value):
                _check_schema_node(schema, value[key], f"{at}/{pointer_segment(key)}", findings, refs)

    for key in js_keys(node):
        value = node[key]
        at = f"{pointer}/{pointer_segment(key)}"
        if key not in SCHEMA_KEYWORDS and key not in SCHEMA_ANNOTATIONS:
            report(at, f'unsupported schema keyword "{key}"')
            continue
        if key == "type":
            types = value if isinstance(value, list) else [value]
            if len(types) == 0 or not all(isinstance(t, str) and t in SCHEMA_TYPES for t in types):
                report(at, f"type must name {', '.join(SCHEMA_TYPES)}")
        elif key == "required":
            if not isinstance(value, list) or not all(isinstance(name, str) for name in value):
                report(at, "required must be an array of key names")
        elif key in ("properties", "$defs"):
            schema_map(value, at)
        elif key == "items":
            _check_schema_node(schema, value, at, findings, refs)
        elif key in ("oneOf", "allOf"):
            subschemas(value, at)
        elif key == "enum":
            if not isinstance(value, list) or len(value) == 0:
                report(at, "enum must be a non-empty array")
        elif key == "pattern":
            try:
                compile_js_pattern(js_string(value))
            except re.error:
                report(at, f"pattern is not a valid regular expression: {js_string(value)}")
            else:
                if not isinstance(value, str):
                    report(at, "pattern must be a string")
        elif key in ("minLength", "minItems"):
            if not is_js_integer(value) or value < 0:  # type: ignore[operator]
                report(at, f"{key} must be a non-negative integer")
        elif key in ("uniqueItems", "additionalProperties"):
            if not isinstance(value, bool):
                report(at, f"{key} must be true or false")
        elif key == "format":
            if value != "datetime-offset":
                report(at, 'format supports only "datetime-offset"')
        elif key == "x-concept":
            if not isinstance(value, str) or (value != "" and not _CONCEPT_PREFIX.fullmatch(value)):
                report(at, 'x-concept is "" or a directory prefix ending in /, such as "decisions/"')
        elif key == "$ref":
            if not isinstance(value, str) or not _REF_PATTERN.fullmatch(value):
                report(
                    at,
                    "$ref must be #/$defs/<name>, <file>.schema.json#/$defs/<name> or <schema-id>#/$defs/<name>",
                )
            else:
                refs.append(at)
        elif key == "$id":
            if not isinstance(value, str):
                report(at, "$id must be a string")


@dataclass(frozen=True)
class _Resolved:
    schema: SchemaFile
    node: object


def resolve_ref(registry: SchemaRegistry, schema: SchemaFile, ref: str) -> _Resolved | None:
    """Resolves in this file, then in a sibling file by name, then in any schema by its $id.

    That order lets a schema in another skill extend the shared definitions without a copy.
    """
    match = _REF_PATTERN.fullmatch(ref)
    if match is None:
        return None
    name = match.group("name")
    reference = match.group("target")
    sibling = (
        registry.files.get(posixpath.join(posixpath.dirname(schema.file), reference))
        if reference is not None and "/" not in reference and reference.endswith(".schema.json")
        else None
    )
    target = schema if reference is None else (sibling or registry.by_id.get(reference))
    defs = target.document.get("$defs") if target is not None and isinstance(target.document, dict) else None
    if target is None or not isinstance(defs, dict) or name not in defs:
        return None
    return _Resolved(target, defs[name])


def _value_at(schema: SchemaFile, pointer: str) -> object:
    node = schema.document
    for segment in pointer.split("/")[1:]:
        key = segment.replace("~1", "/").replace("~0", "~")
        if isinstance(node, dict):
            node = node.get(key)
        elif isinstance(node, list) and key.isdigit() and int(key) < len(node):
            node = node[int(key)]
        else:
            node = None
    return node


def declared_schema_name(document: object) -> str | None:
    """A type schema names the schema it validates in $id or as the const of its schema property."""
    if not isinstance(document, dict):
        return None
    if isinstance(document.get("$id"), str):
        return document["$id"]
    properties = document.get("properties")
    schema = properties.get("schema") if isinstance(properties, dict) else None
    const = schema.get("const") if isinstance(schema, dict) else None
    return const if isinstance(const, str) else None


def _reject_constant(name: str) -> object:
    raise ValueError(f"{name} is not valid JSON")


def build_schema_registry(sources: Iterable[SchemaSource]) -> SchemaRegistry:
    """Loads the schema files and checks that each uses only the supported schema vocabulary.

    A tier 0 schema replaces a default schema of the same $id, so a schema can be tried out before
    it ships.
    """
    files: dict[str, SchemaFile] = {}
    broken: set[str] = set()
    findings: list[Finding] = []
    node_findings: dict[str, list[Finding]] = {}
    for source in sorted(sources, key=lambda item: item.tier):
        if source.file in files or source.file in broken:
            continue
        try:
            document = json.loads(source.text, parse_constant=_reject_constant)
        except ValueError as caught:
            line = caught.lineno if isinstance(caught, json.JSONDecodeError) else 1
            findings.append(error(source.file, line, f"invalid JSON: {caught}"))
            broken.add(source.file)
            continue
        shell = SchemaFile(
            source.file, document, json_lines(source.text), declared_schema_name(document), (), source.tier
        )
        found: list[Finding] = []
        refs: list[str] = []
        _check_schema_node(shell, document, "", found, refs)
        files[source.file] = replace(shell, refs=tuple(refs))
        node_findings[source.file] = found
    declared: dict[tuple[int, str], SchemaFile] = {}
    by_id: dict[str, SchemaFile] = {}
    for schema in files.values():
        if schema.id is None:
            continue
        twin = declared.get((schema.tier, schema.id))
        if twin is None:
            declared[(schema.tier, schema.id)] = schema
            by_id.setdefault(schema.id, schema)
        else:
            at = "/$id" if "/$id" in schema.lines else "/properties/schema/const"
            findings.append(
                error(schema.file, _schema_line(schema, at), f"schema {schema.id} is also declared by {twin.file}")
            )
    partial = SchemaRegistry(files, by_id, frozenset(broken), ())
    for schema in files.values():
        found = list(node_findings[schema.file])
        for at in schema.refs:
            ref = js_string(_value_at(schema, at))
            if resolve_ref(partial, schema, ref) is None:
                found.append(error(schema.file, _schema_line(schema, at), f"$ref target {ref} does not exist"))
        if found:
            broken.add(schema.file)
        findings.extend(found)
    return SchemaRegistry(files, by_id, frozenset(broken), tuple(findings))


def _is_usable(registry: SchemaRegistry, schema: SchemaFile, seen: set[str]) -> bool:
    """Whether the schema and every file it references loaded without errors."""
    if schema.file in registry.broken:
        return False
    if schema.file in seen:
        return True
    seen.add(schema.file)
    for at in schema.refs:
        target = resolve_ref(registry, schema, js_string(_value_at(schema, at)))
        if target is None or not _is_usable(registry, target.schema, seen):
            return False
    return True


def type_name(value: object) -> str:
    if value is None:
        return "null"
    if isinstance(value, bool):
        return "boolean"
    if isinstance(value, list):
        return "array"
    if isinstance(value, (int, float)):
        return "integer" if is_js_integer(value) else "number"
    if isinstance(value, dict):
        return "object"
    return "string"


def has_type(value: object, kind: str) -> bool:
    if kind == "integer":
        return is_js_integer(value)
    if kind == "number":
        return is_number(value)
    return type_name(value) == kind


def deep_equal(a: object, b: object) -> bool:
    if isinstance(a, list) and isinstance(b, list):
        return len(a) == len(b) and all(deep_equal(x, y) for x, y in zip(a, b, strict=True))
    if isinstance(a, dict) and isinstance(b, dict):
        return len(a) == len(b) and all(key in b and deep_equal(a[key], b[key]) for key in a)
    if isinstance(a, bool) or isinstance(b, bool):
        return isinstance(a, bool) and isinstance(b, bool) and a == b
    if is_number(a) and is_number(b):
        return a == b
    if isinstance(a, str) and isinstance(b, str):
        return a == b
    return a is None and b is None


def _issue_path(path: Path) -> str:
    return "".join(
        f"[{segment}]" if isinstance(segment, int) else (segment if index == 0 else f".{segment}")
        for index, segment in enumerate(path)
    )


def issue_text(issue: SchemaIssue) -> str:
    return issue.message if len(issue.path) == 0 else f"{_issue_path(issue.path)}: {issue.message}"


def _concept_problem(concept: str, prefix: str, context: SchemaContext) -> str | None:
    if concept.endswith(".md"):
        return f"is a file name, a concept ID has no .md: {concept}"
    if concept.startswith("/") or any(segment in ("", ".", "..") for segment in concept.split("/")):
        return f"is not a concept ID, a bundle path without a leading / or .md: {concept}"
    if not concept.startswith(prefix):
        return f"must be a concept in {prefix}: {concept}"
    if concept not in context.concepts:
        return f"does not resolve to a file in the bundle: {concept}"
    return None


def check_value(
    value: object, node: object, schema: SchemaFile, context: SchemaContext, path: Path, depth: int
) -> list[SchemaIssue]:
    """depth counts the references followed without descending into the value."""
    if not isinstance(node, dict):
        return []
    issues: list[SchemaIssue] = []

    def issue(message: str) -> None:
        issues.append(SchemaIssue(path, message))

    ref = node.get("$ref")
    if isinstance(ref, str):
        target = resolve_ref(context.registry, schema, ref)
        if depth > 32:
            issue(f"schema references {ref} in a loop")
        elif target is not None:
            issues.extend(check_value(value, target.node, target.schema, context, path, depth + 1))
    if "type" in node:
        declared = node["type"]
        types = [js_string(t) for t in (declared if isinstance(declared, list) else [declared])]
        if not any(has_type(value, t) for t in types):
            return [*issues, SchemaIssue(path, f"expected {' or '.join(types)}, got {type_name(value)}")]
    if "const" in node and not deep_equal(value, node["const"]):
        issue(f"must be {js_stringify(node['const'])}")
    options = node.get("enum")
    if isinstance(options, list) and not any(deep_equal(value, option) for option in options):
        issue(f"must be one of {', '.join(js_stringify(option) for option in options)}")
    if isinstance(value, str):
        pattern = node.get("pattern")
        if isinstance(pattern, str) and compile_js_pattern(pattern).search(value) is None:
            issue(f"{js_stringify(value)} does not match the pattern {pattern}")
        min_length = node.get("minLength")
        if is_number(min_length) and len(value) < min_length:  # type: ignore[operator]
            issue(
                "must not be empty"
                if min_length == 1
                else f"must be at least {js_number(min_length)} characters long"  # type: ignore[arg-type]
            )
        if node.get("format") == "datetime-offset" and not is_datetime_offset(value):
            issue(f"{js_stringify(value)} is not an ISO 8601 datetime with an explicit UTC offset")
    prefix = node.get("x-concept")
    if isinstance(prefix, str):
        problem = (
            _concept_problem(value, prefix, context)
            if isinstance(value, str)
            else f"must be a concept ID, got {type_name(value)}"
        )
        if problem is not None:
            issue(problem)
    if isinstance(value, dict):
        required = node.get("required")
        if isinstance(required, list):
            for name in (js_string(item) for item in required):
                if name not in value:
                    issue(f"missing required key {name}")
        properties = node["properties"] if isinstance(node.get("properties"), dict) else {}
        for key in js_keys(properties):
            if key in value:
                issues.extend(check_value(value[key], properties[key], schema, context, (*path, key), 0))
        if node.get("additionalProperties") is False:
            for key in js_keys(value):
                if key not in properties:
                    issues.append(SchemaIssue((*path, key), "is not an allowed key"))
    if isinstance(value, list):
        if "items" in node:
            for index, item in enumerate(value):
                issues.extend(check_value(item, node["items"], schema, context, (*path, index), 0))
        min_items = node.get("minItems")
        if is_number(min_items) and len(value) < min_items:  # type: ignore[operator]
            issue(
                "must not be empty"
                if min_items == 1
                else f"must have at least {js_number(min_items)} items"  # type: ignore[arg-type]
            )
        if node.get("uniqueItems") is True:
            for index, item in enumerate(value):
                first = next(i for i, other in enumerate(value) if deep_equal(other, item))
                if first < index:
                    issues.append(SchemaIssue((*path, index), f"repeats item [{first}]"))
    all_of = node.get("allOf")
    if isinstance(all_of, list):
        for sub in all_of:
            issues.extend(check_value(value, sub, schema, context, path, depth))
    one_of = node.get("oneOf")
    if isinstance(one_of, list):
        results = [check_value(value, sub, schema, context, path, depth) for sub in one_of]
        matches = sum(1 for result in results if len(result) == 0)
        if matches == 0:
            shapes = []
            for index, result in enumerate(results):
                text = (
                    issue_text(SchemaIssue(result[0].path[len(path) :], result[0].message))
                    if result
                    else ""
                )
                shapes.append(f"shape {index + 1}: {text}")
            issue(f"matches none of the allowed shapes ({'; '.join(shapes)})")
        elif matches > 1:
            issue("matches more than one allowed shape")
    return issues


def _path_pointer(path: Path) -> str:
    return "".join(f"/{pointer_segment(js_string(segment))}" for segment in path)


_ABSENT = object()


def schema_findings(bundle: Bundle, registry: SchemaRegistry) -> tuple[Finding, ...]:
    context = SchemaContext(registry, frozenset(bundle.concepts))
    findings: list[Finding] = []
    for file in bundle.concepts.values():
        if file.data is None or not is_workflow(file):
            continue
        name = file.data.get("schema", _ABSENT)
        if not isinstance(name, str) or not _SCHEMA_NAME.fullmatch(name):
            found = "nothing" if name is _ABSENT else js_stringify(name)
            findings.append(
                error(
                    file.path,
                    line_of(file, "/type" if name is _ABSENT else "/schema"),
                    f"schema must name workflow/<kebab-name>/<integer>, found {found}",
                )
            )
            continue
        schema = registry.by_id.get(name)
        if schema is None:
            findings.append(warning(file.path, line_of(file, "/schema"), f"schema file for {name} not found"))
        elif not _is_usable(registry, schema, set()):
            findings.append(
                warning(
                    file.path,
                    line_of(file, "/schema"),
                    f"schema file for {name} has errors and was not applied",
                )
            )
        else:
            for issue in check_value(file.data, schema.document, schema, context, (), 0):
                findings.append(error(file.path, line_of(file, _path_pointer(issue.path)), issue_text(issue)))
    return tuple(findings)


# ---------------------------------------------------------------------------
# Layer 3: cross-file rules
# ---------------------------------------------------------------------------

RELATION_KEYS = (
    "depends_on",
    "blocks",
    "supersedes",
    "superseded_by",
    "waits_on",
    "decisions",
    "parent",
    "decision",
    "touched",
)


@dataclass(frozen=True)
class _Target:
    id: str
    pointer: str


def relation_targets(file: BundleFile, key: str) -> tuple[_Target, ...]:
    value = None if file.data is None else file.data.get(key)
    if isinstance(value, str):
        return (_Target(value, f"/{key}"),)
    if not isinstance(value, list):
        return ()
    return tuple(_Target(item, f"/{key}/{index}") for index, item in enumerate(value) if isinstance(item, str))


def _lists_id(value: object, concept: str) -> bool:
    return isinstance(value, list) and any(isinstance(item, str) and item == concept for item in value)


_SLUG = re.compile(r"[a-z0-9]+(?:-[a-z0-9]+)*")
_SUFFIX = re.compile(r"[a-z2-7]{4}")
_SLUG_SUFFIX = re.compile(f"({_DOT}+)-([a-z2-7]{{4}})")
_LOG_NAME = re.compile(f"([0-9]{{4}}-[0-9]{{2}}-[0-9]{{2}})-({_DOT}+)-([a-z2-7]{{4}})")
SLUG_SUFFIX_RULE = (
    "<slug>-<suffix>: a lowercase kebab-case slug of at most 48 characters, a hyphen and 4 "
    "characters from a-z and 2-7"
)
SUFFIX_RULE = "<suffix>: 4 characters from a-z and 2-7"


def is_slug_suffix(name: str) -> bool:
    match = _SLUG_SUFFIX.fullmatch(name)
    return match is not None and _SLUG.fullmatch(match.group(1)) is not None and len(match.group(1)) <= 48


def is_log_name(name: str) -> bool:
    match = _LOG_NAME.fullmatch(name)
    return match is not None and is_calendar_date(match.group(1)) and _SLUG.fullmatch(match.group(2)) is not None


@dataclass(frozen=True)
class Location:
    type: str
    place: str
    directory: re.Pattern[str]
    name: Callable[[str], bool]
    rule: str


def _named(expected: str) -> Callable[[str], bool]:
    return lambda name: name == expected


def _is_suffix(name: str) -> bool:
    return _SUFFIX.fullmatch(name) is not None


LOCATIONS = (
    Location("Workflow Plan", "plan.md", re.compile(""), _named("plan"), ""),
    Location("Workflow Close-out Record", "close-out.md", re.compile(""), _named("close-out"), ""),
    Location("Workflow Research", "research/", re.compile("research"), is_slug_suffix, SLUG_SUFFIX_RULE),
    Location(DECISION, "decisions/", re.compile("decisions"), is_slug_suffix, SLUG_SUFFIX_RULE),
    Location("Workflow Critique", "critiques/", re.compile("critiques"), is_slug_suffix, SLUG_SUFFIX_RULE),
    Location(WORK_ITEM, "work/", re.compile("work"), is_slug_suffix, SLUG_SUFFIX_RULE),
    Location("Workflow Verdict", "verdicts/", re.compile("verdicts"), is_slug_suffix, SLUG_SUFFIX_RULE),
    Location("Workflow Log Entry", "log/", re.compile("log"), is_log_name, "<YYYY-MM-DD>-<skill>-<suffix>"),
    Location(
        "Workflow Implementation Notes",
        "implementation/<work-id>/notes.md",
        re.compile("implementation/[^/]+"),
        _named("notes"),
        "",
    ),
    Location(
        "Workflow Commit Message",
        "deliveries/<work-id>/commit-message.md",
        re.compile("deliveries/[^/]+"),
        _named("commit-message"),
        "",
    ),
    Location(
        "Workflow Pull Request Description",
        "deliveries/<work-id>/pull-request.md",
        re.compile("deliveries/[^/]+"),
        _named("pull-request"),
        "",
    ),
    Location(
        "Workflow Inline Comment",
        "deliveries/<work-id>/comments/",
        re.compile("deliveries/[^/]+/comments"),
        _is_suffix,
        SUFFIX_RULE,
    ),
)


def location_findings(file: BundleFile) -> tuple[Finding, ...]:
    location = next((candidate for candidate in LOCATIONS if candidate.type == type_of(file)), None)
    if location is None:
        return ()
    line = line_of(file, "/type")
    if not location.directory.fullmatch(file.directory):
        return (error(file.path, line, f"a {location.type} belongs at {location.place}"),)
    segments = file.directory.split("/")
    work_directory = segments[1] if len(segments) > 1 else ""
    if re.match("(?:implementation|deliveries)/", file.directory) and not is_slug_suffix(work_directory):
        return (
            error(file.path, line, f"work item directory {work_directory} does not match {SLUG_SUFFIX_RULE}"),
        )
    if not location.name(file.name[:-3]):
        message = (
            f"a {location.type} belongs at {location.place}"
            if location.place.endswith(".md")
            else f"file name {file.name} does not match {location.rule}"
        )
        return (error(file.path, line, message),)
    return ()


def decision_findings(bundle: Bundle, file: BundleFile) -> tuple[Finding, ...]:
    findings: list[Finding] = []
    data = file.data or {}
    replacement = data.get("superseded_by")
    if data.get("status") == "deprecated" and not isinstance(replacement, str):
        findings.append(
            error(file.path, line_of(file, "/status"), "a deprecated decision names its replacement in superseded_by")
        )
    if isinstance(replacement, str):
        line = line_of(file, "/superseded_by")
        target = bundle.concepts.get(replacement)
        if data.get("status") != "deprecated":
            findings.append(error(file.path, line, "superseded_by is set but status is not deprecated"))
        if target is not None and target.data is not None and replacement != file.id:
            if type_of(target) != DECISION:
                findings.append(error(file.path, line, f"superseded_by does not resolve to a decision: {replacement}"))
            elif not _lists_id(target.data.get("supersedes"), file.id):
                findings.append(error(file.path, line, f"{replacement} does not list this decision in supersedes"))
    for target_ref in relation_targets(file, "supersedes"):
        target = bundle.concepts.get(target_ref.id)
        if target is None or target.data is None or target_ref.id == file.id:
            continue
        line = line_of(file, target_ref.pointer)
        if type_of(target) != DECISION:
            findings.append(error(file.path, line, f"supersedes does not resolve to a decision: {target_ref.id}"))
        elif target.data.get("status") != "deprecated":
            findings.append(error(file.path, line, f"supersedes {target_ref.id}, which is not deprecated"))
        elif target.data.get("superseded_by") != file.id:
            findings.append(
                error(file.path, line, f"supersedes {target_ref.id}, whose superseded_by is not this decision")
            )
    return tuple(findings)


def work_findings(bundle: Bundle, file: BundleFile) -> tuple[Finding, ...]:
    findings: list[Finding] = []
    data = file.data or {}
    waits = data.get("waits_on")
    waiting = isinstance(waits, list) and len(waits) > 0
    if data.get("readiness") == "waiting" and not waiting:
        findings.append(
            error(file.path, line_of(file, "/readiness"), "readiness is waiting but waits_on names no decision")
        )
    if data.get("readiness") == "ready" and waiting:
        findings.append(
            warning(file.path, line_of(file, "/waits_on"), "readiness is ready but waits_on still names decisions")
        )
    for target_ref in relation_targets(file, "parent"):
        target = bundle.concepts.get(target_ref.id)
        if (
            target is not None
            and target.data is not None
            and target_ref.id != file.id
            and (type_of(target) != WORK_ITEM or target.data.get("kind") != "work-package")
        ):
            findings.append(
                error(file.path, line_of(file, target_ref.pointer), f"parent is not a work package: {target_ref.id}")
            )
    return tuple(findings)


def target_type_findings(bundle: Bundle, file: BundleFile) -> tuple[Finding, ...]:
    findings: list[Finding] = []
    for key, kind in (("decision", DECISION),):
        for target_ref in relation_targets(file, key):
            target = bundle.concepts.get(target_ref.id)
            if target is not None and target.data is not None and target_ref.id != file.id and type_of(target) != kind:
                findings.append(
                    error(
                        file.path,
                        line_of(file, target_ref.pointer),
                        f"{key} does not resolve to a {kind}: {target_ref.id}",
                    )
                )
    return tuple(findings)


def unique_id_findings(file: BundleFile) -> tuple[Finding, ...]:
    findings: list[Finding] = []
    for key in ("sources", "criteria"):
        entries = None if file.data is None else file.data.get(key)
        if not isinstance(entries, list):
            continue
        seen: dict[str, int] = {}
        for index, entry in enumerate(entries):
            entry_id = entry.get("id") if isinstance(entry, dict) else None
            if not isinstance(entry_id, str):
                continue
            first = seen.get(entry_id)
            if first is None:
                seen[entry_id] = index
            else:
                findings.append(
                    error(
                        file.path,
                        line_of(file, f"/{key}/{index}/id"),
                        f"id {entry_id} is already used by {key}[{first}]",
                    )
                )
    return tuple(findings)


_FENCE = re.compile(r" {0,3}(`{3,}|~{3,})")
_CODE_SPAN = re.compile(r"`+[^`]*`+")


def prose_lines(body: str) -> tuple[str, ...]:
    """Blanks fenced code blocks and inline code spans but keeps every line.

    Footnote positions then still map to file lines.
    """
    fence = ""
    out: list[str] = []
    for line in body.split("\n"):
        match = _FENCE.match(line)
        marker = match.group(1) if match else None
        if fence == "" and marker is not None:
            fence = marker
            out.append("")
            continue
        if fence != "":
            if marker is not None and marker.startswith(fence[0]) and len(marker) >= len(fence):
                fence = ""
            out.append("")
            continue
        out.append(_CODE_SPAN.sub(lambda code: " " * len(code.group(0)), line))
    return tuple(out)


_FOOTNOTE_DEFINITION = re.compile(rf" {{0,3}}\[\^([^\]{_WS}]+)\]:")
_FOOTNOTE_REFERENCE = re.compile(rf"\[\^([^\]{_WS}]+)\]")


def footnote_findings(file: BundleFile) -> tuple[Finding, ...]:
    findings: list[Finding] = []
    source_ids: dict[str, str] = {}
    sources = None if file.data is None else file.data.get("sources")
    if isinstance(sources, list):
        for index, source in enumerate(sources):
            source_id = source.get("id") if isinstance(source, dict) else None
            if isinstance(source_id, str) and source_id not in source_ids:
                source_ids[source_id] = f"/sources/{index}/id"
    cited: dict[str, int] = {}
    defined: set[str] = set()

    def check_label(label: str, line: int) -> None:
        if label not in source_ids:
            findings.append(error(file.path, line, f"footnote [^{label}] matches no sources id"))

    for index, text in enumerate(prose_lines(file.body)):
        line = file.body_line + index
        definition = _FOOTNOTE_DEFINITION.match(text)
        if definition is not None:
            defined.add(definition.group(1))
            check_label(definition.group(1), line)
        references = text if definition is None else text[definition.end() :]
        for match in _FOOTNOTE_REFERENCE.finditer(references):
            label = match.group(1)
            if label not in cited:
                cited[label] = line
                check_label(label, line)
    for label, line in cited.items():
        if label not in defined:
            findings.append(error(file.path, line, f"footnote [^{label}] has no definition"))
    for source_id, pointer in source_ids.items():
        if source_id not in cited:
            findings.append(
                warning(file.path, line_of(file, pointer), f"source {source_id} is never cited in the body")
            )
    return tuple(findings)


def work_cycle_findings(bundle: Bundle) -> tuple[Finding, ...]:
    """Finds depends_on cycles between work items with Tarjan's algorithm.

    A self-reference is left to the relation rule that reports it.
    """
    items = [file for file in bundle.concepts.values() if file.data is not None and type_of(file) == WORK_ITEM]
    ids = {file.id for file in items}
    marks: dict[str, list[int]] = {}
    stack: list[str] = []
    cycles: list[list[str]] = []

    def visit(file: BundleFile) -> None:
        mark = [len(marks), len(marks)]
        marks[file.id] = mark
        stack.append(file.id)
        for target_ref in relation_targets(file, "depends_on"):
            target_id = target_ref.id
            if target_id not in ids or target_id == file.id:
                continue
            known = marks.get(target_id)
            target = bundle.concepts.get(target_id)
            if known is None and target is not None:
                visit(target)
                mark[1] = min(mark[1], marks[target_id][1])
            elif known is not None and target_id in stack:
                mark[1] = min(mark[1], known[0])
        if mark[1] != mark[0]:
            return
        component: list[str] = []
        while stack:
            popped = stack.pop()
            component.append(popped)
            if popped == file.id:
                break
        if len(component) > 1:
            cycles.append(sorted(component, key=text_key))

    for file in items:
        if file.id not in marks:
            visit(file)
    findings: list[Finding] = []
    for cycle in cycles:
        first = bundle.concepts.get(cycle[0])
        findings.append(
            error(
                "" if first is None else first.path,
                1 if first is None else line_of(first, "/depends_on"),
                f"depends_on forms a cycle: {', '.join(cycle)}",
            )
        )
    return tuple(findings)


def cross_file_findings(bundle: Bundle) -> tuple[Finding, ...]:
    findings: list[Finding] = []
    for file in bundle.concepts.values():
        if file.data is None or not is_workflow(file):
            continue
        for key in RELATION_KEYS:
            for target_ref in relation_targets(file, key):
                if target_ref.id == file.id:
                    findings.append(error(file.path, line_of(file, target_ref.pointer), f"{key} refers to this file itself"))
        kind = type_of(file)
        if kind == DECISION:
            findings.extend(decision_findings(bundle, file))
        if kind == WORK_ITEM:
            findings.extend(work_findings(bundle, file))
        findings.extend(target_type_findings(bundle, file))
        findings.extend(unique_id_findings(file))
        findings.extend(footnote_findings(file))
        findings.extend(location_findings(file))
    return (*findings, *work_cycle_findings(bundle))


# ---------------------------------------------------------------------------
# Generated files
# ---------------------------------------------------------------------------


@dataclass(frozen=True)
class Generated:
    files: tuple[tuple[str, str], ...]
    obsolete: tuple[str, ...]
    findings: tuple[Finding, ...]


@dataclass(frozen=True)
class Section:
    heading: str
    entries: tuple[str, ...]


@dataclass(frozen=True)
class DirectoryRole:
    heading: str
    first: tuple[str, ...]


DIRECTORY_DESCRIPTIONS = {
    "research": "Research behind the plan",
    "critiques": "Adversarial reviews of the plan",
    "verdicts": "Validation verdicts for each work item",
    "log": "One entry for each skill run",
    "implementation": "Implementation notes for each work item",
    "deliveries": "Delivery text for each work item",
    "decisions": "Decisions that shape the plan",
    "work": "Work packages and work items",
}

_DIRECTORY_ROLES = (
    (re.compile("research"), DirectoryRole("Research", ())),
    (re.compile("critiques"), DirectoryRole("Critiques", ())),
    (re.compile("verdicts"), DirectoryRole("Verdicts", ())),
    (re.compile("log"), DirectoryRole("Log entries", ())),
    (re.compile("deliveries/[^/]+"), DirectoryRole("Delivery", ("commit-message", "pull-request"))),
    (re.compile("deliveries/[^/]+/comments"), DirectoryRole("Comments", ())),
    (re.compile("implementation/[^/]+"), DirectoryRole("Implementation notes", ("notes",))),
)


def directory_role(directory: str) -> DirectoryRole:
    if directory == "":
        return DirectoryRole("Documents", ("plan", "close-out"))
    return next(
        (role for pattern, role in _DIRECTORY_ROLES if pattern.fullmatch(directory)),
        DirectoryRole("Documents", ()),
    )


def escape_link_text(text: str) -> str:
    return re.sub(r"[\\\[\]]", lambda match: "\\" + match.group(0), text)


def encode_path(path: str) -> str:
    """encodeURIComponent of each segment."""
    return "/".join(quote(segment, safe="!'()*") for segment in path.split("/"))


def entry_line(title: str, url: str, description: str) -> str:
    return f"* [{escape_link_text(title)}]({url})" + ("" if description == "" else f" - {description}")


def _title_or(file: BundleFile | None, fallback: str) -> str:
    title = None if file is None or file.data is None else file.data.get("title")
    text = collapse(title) if isinstance(title, str) else ""
    return fallback if text == "" else text


def title_of(file: BundleFile) -> str:
    return _title_or(file, file.name[:-3])


def description_of(file: BundleFile | None) -> str:
    description = None if file is None or file.data is None else file.data.get("description")
    return collapse(description) if isinstance(description, str) else ""


def generated_at(file: BundleFile) -> int | None:
    generated = None if file.data is None else file.data.get("generated")
    at = generated.get("at") if isinstance(generated, dict) else None
    return instant_ms(at) if isinstance(at, str) else None


def compare_creation(a: BundleFile, b: BundleFile, direction: int) -> int:
    """Untimed files sort after timed ones in either direction."""
    x, y = generated_at(a), generated_at(b)
    if x is None or y is None:
        return ((1 if x is None else 0) - (1 if y is None else 0)) or compare_text(a.id, b.id)
    return compare_numbers(x, y) * direction or compare_text(a.id, b.id)


def _compare_oldest_first(a: BundleFile, b: BundleFile) -> int:
    return compare_creation(a, b, 1)


def _compare_newest_first(a: BundleFile, b: BundleFile) -> int:
    return compare_creation(a, b, -1)


_OLDEST_FIRST = cmp_to_key(_compare_oldest_first)
_NEWEST_FIRST = cmp_to_key(_compare_newest_first)


def _parent_of(directory: str) -> str:
    return directory[: directory.rfind("/")] if "/" in directory else ""


def _base_name(directory: str) -> str:
    return directory[directory.rfind("/") + 1 :]


def _concepts_in(bundle: Bundle, directory: str) -> list[BundleFile]:
    return [file for file in bundle.files if not file.reserved and file.directory == directory]


def _children_of(bundle: Bundle, directory: str) -> list[str]:
    return [child for child in bundle.directories if child != "" and _parent_of(child) == directory]


def _depth_of(directory: str) -> int:
    return 0 if directory == "" else len(directory.split("/"))


def is_unindexed(directory: str) -> bool:
    return directory.split("/")[0] in UNINDEXED_DIRECTORIES


def indexed_directories(bundle: Bundle) -> frozenset[str]:
    """The directories that get an index.md."""
    indexed: set[str] = set()
    deepest_first = sorted(bundle.directories, key=lambda d: (-_depth_of(d), text_key(d)))
    for directory in deepest_first:
        if directory != "" and is_unindexed(directory):
            continue
        if _concepts_in(bundle, directory) or any(child in indexed for child in _children_of(bundle, directory)):
            indexed.add(directory)
    return frozenset(indexed)


def directory_description(bundle: Bundle, directory: str) -> str:
    fixed = DIRECTORY_DESCRIPTIONS.get(directory)
    if fixed is not None:
        return fixed
    match = re.fullmatch("(?:implementation|deliveries)/([^/]+)", directory)
    if match is not None:
        work = bundle.concepts.get(f"work/{match.group(1)}")
        return "" if work is None or work.data is None else title_of(work)
    if re.fullmatch("deliveries/[^/]+/comments", directory):
        return "Inline comments on the pull request"
    return ""


def _relative_url(directory: str, file: BundleFile) -> str:
    return encode_path(file.path if directory == "" else file.path[len(directory) + 1 :])


def _concept_entry(directory: str, file: BundleFile, prefix: str = "") -> str:
    return entry_line(f"{prefix}{title_of(file)}", _relative_url(directory, file), description_of(file))


DECISION_GROUPS = (
    ("Settled", "settled"),
    ("Open", "open"),
    ("Paused", "paused"),
    ("Partly settled", "partly-settled"),
)


def decision_sections(decisions: Sequence[BundleFile]) -> tuple[tuple[Section, ...], tuple[Finding, ...]]:
    findings: list[Finding] = []
    groups: dict[str, list[BundleFile]] = {heading: [] for heading, _ in DECISION_GROUPS}
    groups["Superseded"] = []
    for file in decisions:
        data = file.data or {}
        state = data.get("state")
        group = (
            "Superseded"
            if data.get("status") == "deprecated"
            else next((heading for heading, value in DECISION_GROUPS if isinstance(state, str) and value == state), None)
        )
        if group is None:
            findings.append(
                error(
                    file.path,
                    line_of(file, "/state"),
                    "state must be settled, open, paused or partly-settled to group the decision",
                )
            )
        else:
            groups[group].append(file)
    sections: list[Section] = []
    number = 0
    for heading, members in groups.items():
        if not members:
            continue
        entries: list[str] = []
        for file in sorted(members, key=_OLDEST_FIRST):
            number += 1
            entries.append(_concept_entry("decisions", file, f"{number}. "))
        sections.append(Section(heading, tuple(entries)))
    return tuple(sections), tuple(findings)


def work_sections(items: Sequence[BundleFile]) -> tuple[tuple[Section, ...], tuple[Finding, ...]]:
    """Orders work by depends_on, ties by creation time; a cycle leaves its members out."""
    findings: list[Finding] = []
    valid: list[BundleFile] = []
    for file in items:
        readiness = (file.data or {}).get("readiness")
        if readiness in ("ready", "waiting") and isinstance(readiness, str):
            valid.append(file)
        else:
            findings.append(
                error(
                    file.path,
                    line_of(file, "/readiness"),
                    "readiness must be ready or waiting to group the work item",
                )
            )
    ids = {file.id for file in valid}
    remaining = sorted(valid, key=_OLDEST_FIRST)
    placed: set[str] = set()
    order: list[BundleFile] = []
    while True:
        index = next(
            (
                i
                for i, file in enumerate(remaining)
                if all(
                    t.id in placed or t.id not in ids or t.id == file.id
                    for t in relation_targets(file, "depends_on")
                )
            ),
            -1,
        )
        if index < 0:
            break
        file = remaining.pop(index)
        order.append(file)
        placed.add(file.id)
    sections = tuple(
        Section(
            heading,
            tuple(_concept_entry("work", file) for file in order if (file.data or {}).get("readiness") == readiness),
        )
        for heading, readiness in (("Ready now", "ready"), ("Waiting on a decision", "waiting"))
    )
    return tuple(section for section in sections if section.entries), tuple(findings)


def render_index(sections: Sequence[Section], root: bool) -> str:
    body = "\n\n".join(
        f"# {section.heading}" if not section.entries else f"# {section.heading}\n\n" + "\n".join(section.entries)
        for section in sections
    )
    front = f'---\nokf_version: "{OKF_VERSION}"\n---\n\n' if root else ""
    return f"{front}{body}\n"


def directory_index(
    bundle: Bundle, directory: str, indexed: frozenset[str]
) -> tuple[str | None, tuple[Finding, ...]]:
    """Returns the index text, or None with the findings that block grouping or ordering."""
    findings: list[Finding] = []
    listable: list[BundleFile] = []
    for file in _concepts_in(bundle, directory):
        if file.data is None:
            findings.append(error(file.path, 1, "an index cannot list this file until its frontmatter parses"))
        elif type_of(file) == "":
            findings.append(error(file.path, 1, "an index cannot list this file until it has a type"))
        else:
            listable.append(file)
    children = [child for child in _children_of(bundle, directory) if child in indexed]
    sections: list[Section] = []
    documents = listable
    if directory == "decisions":
        found, problems = decision_sections([file for file in listable if type_of(file) == DECISION])
        sections.extend(found)
        findings.extend(problems)
        documents = [file for file in listable if type_of(file) != DECISION]
    elif directory == "work":
        found, problems = work_sections([file for file in listable if type_of(file) == WORK_ITEM])
        sections.extend(found)
        findings.extend(problems)
        documents = [file for file in listable if type_of(file) != WORK_ITEM]
    role = directory_role(directory)
    pinned = [file for name in role.first for file in documents if file.name == f"{name}.md"]
    rest = sorted((file for file in documents if file not in pinned), key=_NEWEST_FIRST)
    heading = "Documents" if directory in ("decisions", "work") else role.heading
    if pinned or rest:
        sections.append(Section(heading, tuple(_concept_entry(directory, file) for file in (*pinned, *rest))))
    if children:
        sections.append(
            Section(
                "Directories",
                tuple(
                    entry_line(_base_name(child), f"{encode_path(_base_name(child))}/", directory_description(bundle, child))
                    for child in sorted(children, key=text_key)
                ),
            )
        )
    if findings:
        return None, tuple(findings)
    return render_index(sections, directory == ""), ()


def generate_indexes(bundle: Bundle) -> Generated:
    """Builds every index.md; a directory whose files block grouping or ordering gets no content."""
    findings: list[Finding] = []
    files: list[tuple[str, str]] = []
    indexed = indexed_directories(bundle)
    cycles = work_cycle_findings(bundle)
    for directory in bundle.directories:
        if directory not in indexed:
            continue
        content, problems = directory_index(bundle, directory, indexed)
        findings.extend(problems)
        path = "index.md" if directory == "" else f"{directory}/index.md"
        if directory == "work" and cycles:
            findings.extend(cycles)
        elif content is not None:
            files.append((path, content))
    obsolete = tuple(
        file.path
        for file in bundle.files
        if file.name == "index.md" and file.directory not in indexed and not is_unindexed(file.directory)
    )
    return Generated(tuple(files), obsolete, tuple(findings))


def first_paragraph(body: str) -> str:
    """The summary of a log entry, so a table or notes after it stay out of the one-line log."""
    lines = prose_lines(body)
    start = next((i for i, line in enumerate(lines) if js_trim(line) != ""), -1)
    if start < 0:
        return ""
    end = next((i for i in range(start + 1, len(lines)) if js_trim(lines[i]) == ""), len(lines))
    return collapse(" ".join(lines[start:end]))


def generate_log(bundle: Bundle) -> Generated:
    findings: list[Finding] = []
    dates: dict[str, list[tuple[BundleFile, str]]] = {}
    for file in _concepts_in(bundle, "log"):
        if file.data is None:
            findings.append(error(file.path, 1, "log.md cannot list this entry until its frontmatter parses"))
            continue
        generated = file.data.get("generated")
        at = generated.get("at") if isinstance(generated, dict) else None
        action = file.data.get("action")
        summary = first_paragraph(file.body)
        before = len(findings)
        if not isinstance(at, str) or not is_datetime_offset(at):
            findings.append(
                error(
                    file.path,
                    line_of(file, "/generated"),
                    "generated.at must be a datetime with a UTC offset to date the log entry",
                )
            )
        if not isinstance(action, str) or collapse(action) == "":
            findings.append(error(file.path, line_of(file, "/action"), "action must name the log entry's action"))
        if summary == "":
            findings.append(error(file.path, file.body_line, "the body must hold the one-sentence summary"))
        if len(findings) > before or not isinstance(at, str) or not isinstance(action, str):
            continue
        name = file.name[:-3]
        line = (
            f"* **{collapse(action)}**: {summary} [{escape_link_text(name)}](/log/{encode_path(file.name)})"
        )
        dates.setdefault(at[:10], []).append((file, line))
    groups = []
    for date in sorted(dates, key=text_key, reverse=True):
        entries = sorted(dates[date], key=lambda entry: _NEWEST_FIRST(entry[0]))
        groups.append(f"\n## {date}\n" + "\n".join(line for _, line in entries) + "\n")
    return Generated((("log.md", "# Directory Update Log\n" + "".join(groups)),), (), tuple(findings))


def stale_findings(bundle: Bundle, generated: Generated, command: str) -> tuple[Finding, ...]:
    findings: list[Finding] = []
    advice = f"run the {command} command with --write"
    texts = {file.path: file.text for file in bundle.files}
    for path, content in generated.files:
        existing = texts.get(path)
        if existing is None:
            findings.append(error(path, 1, f"generated file is missing; {advice}"))
        elif existing != content:
            actual = existing.split("\n")
            expected = content.split("\n")
            line = next(
                (i for i, text in enumerate(expected) if i >= len(actual) or actual[i] != text),
                len(expected),
            )
            findings.append(error(path, line + 1, f"generated file is stale; {advice}"))
    for path in generated.obsolete:
        findings.append(
            error(path, 1, f"generated file is not needed in a directory without concept documents; {advice}")
        )
    return tuple(findings)


# ---------------------------------------------------------------------------
# Validation
# ---------------------------------------------------------------------------


def validate_bundle(bundle: Bundle, registry: SchemaRegistry) -> tuple[Finding, ...]:
    """Runs the three layers and checks the generated files.

    A generator finding is dropped for a file that already has an error, because that error is its
    cause.
    """
    findings = [
        *registry.findings,
        *bundle.findings,
        *okf_findings(bundle),
        *schema_findings(bundle, registry),
        *cross_file_findings(bundle),
    ]
    failing = {finding.path for finding in findings if finding.severity == "error"}
    for command, generated in (("index", generate_indexes(bundle)), ("log", generate_log(bundle))):
        findings.extend(finding for finding in generated.findings if finding.path not in failing)
        findings.extend(stale_findings(bundle, generated, command))
    return finish(findings)


# ---------------------------------------------------------------------------
# Command line core
# ---------------------------------------------------------------------------

USAGE = """Usage: uv run plan-bundle.py <command> <bundle-dir> [options]

Commands:
  validate  check Open Knowledge Format conformance, Workflow schemas and cross-file rules
  index     generate index.md in the bundle directories
  log       generate log.md from the entries in log/

Options:
  --check             report stale or missing generated files (index, log)
  --write             rewrite stale or missing generated files (index, log)
  --schemas <dir>     add a schema directory, repeatable (validate)
  --format text|json  output format, text by default
  -h, --help          print this help

A <bundle-dir> that holds roadmap.md and no plan.md and no log/ is a family: the command runs on
each child directory that holds log/.

Exit status: 0 clean or warnings only, 1 errors, 2 usage error.
"""

type Command = Literal["validate", "index", "log"]
type Mode = Literal["check", "write"]


@dataclass(frozen=True)
class Options:
    command: Command
    bundle: str
    schemas: tuple[str, ...]
    format: Literal["text", "json"]
    mode: Mode | None


@dataclass(frozen=True)
class HelpRequest:
    pass


@dataclass(frozen=True)
class UsageError:
    message: str


def parse_arguments(argv: Sequence[str]) -> Options | HelpRequest | UsageError:
    if "-h" in argv or "--help" in argv:
        return HelpRequest()
    command = argv[0] if argv else None
    rest = argv[1:]
    if command not in ("validate", "index", "log"):
        return UsageError("missing command" if command is None else f"unknown command: {command}")
    positional: list[str] = []
    schemas: list[str] = []
    output: Literal["text", "json"] = "text"
    mode: Mode | None = None
    index = 0
    while index < len(rest):
        argument = rest[index]
        if argument in ("--schemas", "--format"):
            value = rest[index + 1] if index + 1 < len(rest) else None
            index += 1
            if value is None:
                return UsageError(f"{argument} needs a value")
            if argument == "--schemas":
                if command != "validate":
                    return UsageError("--schemas applies only to validate")
                schemas.append(value)
            elif value in ("text", "json"):
                output = "text" if value == "text" else "json"
            else:
                return UsageError(f"--format is text or json, not {value}")
        elif argument in ("--check", "--write"):
            if command == "validate":
                return UsageError(f"{argument} applies only to index and log")
            requested: Mode = "check" if argument == "--check" else "write"
            if mode is not None and mode != requested:
                return UsageError("use --check or --write, not both")
            mode = requested
        elif argument.startswith("-"):
            return UsageError(f"unknown option: {argument}")
        else:
            positional.append(argument)
        index += 1
    if not positional:
        return UsageError("missing bundle directory")
    if len(positional) > 1:
        return UsageError(f"unexpected argument: {positional[1]}")
    if command != "validate" and mode is None:
        return UsageError(f"{command} needs --check or --write")
    return Options(command, positional[0], tuple(schemas), output, mode)  # type: ignore[arg-type]


@dataclass(frozen=True)
class DirectoryEntry:
    """One entry of a directory listing; subdirectories names the directories inside a directory."""

    name: str
    kind: Literal["file", "directory", "other"]
    subdirectories: tuple[str, ...] = ()


def family_members(entries: Iterable[DirectoryEntry]) -> tuple[str, ...] | None:
    """Returns the bundles of a family in name order, or None when the directory is no family.

    A family holds roadmap.md, no plan.md and no log/ directory; its bundles are the child
    directories that hold log/, because every bundle starts with an Initialization entry there.
    """
    listing = tuple(entries)
    names = {entry.name: entry for entry in listing}
    roadmap = names.get("roadmap.md")
    log = names.get("log")
    if roadmap is None or roadmap.kind != "file" or "plan.md" in names or (log is not None and log.kind == "directory"):
        return None
    members = (
        entry.name
        for entry in listing
        if entry.kind == "directory" and not entry.name.startswith(".") and "log" in entry.subdirectories
    )
    return tuple(sorted(members, key=text_key))


@dataclass(frozen=True)
class BundleRun:
    """The outcome of a command on one bundle: findings, and the files a --write run changes."""

    findings: tuple[Finding, ...]
    writes: tuple[tuple[str, str], ...]
    deletes: tuple[str, ...]
    writing: bool


def run_bundle(options: Options, bundle: Bundle, registry: SchemaRegistry) -> BundleRun:
    if options.command == "validate":
        return BundleRun(validate_bundle(bundle, registry), (), (), False)
    generated = generate_indexes(bundle) if options.command == "index" else generate_log(bundle)
    writing = options.mode == "write"
    if writing and not generated.findings:
        texts = {file.path: file.text for file in bundle.files}
        writes = tuple((path, content) for path, content in generated.files if texts.get(path) != content)
        return BundleRun((), writes, generated.obsolete, True)
    return BundleRun(finish((*generated.findings, *stale_findings(bundle, generated, options.command))), (), (), writing)


def prefix_finding(finding: Finding, member: str) -> Finding:
    """Puts the bundle directory of a family member before a bundle-relative path."""
    if finding.path.startswith("/"):
        return finding
    return replace(finding, path=f"{member}/{finding.path}")


def display_path(path: str, cwd: str) -> str:
    if not path.startswith("/"):
        return path
    inside = posixpath.relpath(path, cwd)
    return path if inside.startswith("..") or inside.startswith("/") else inside


def plural(count: int, noun: str) -> str:
    return f"{count} {noun}{'' if count == 1 else 's'}"


_LONE_SURROGATE = re.compile("[\ud800-\udfff]")


def render_report(
    findings: Sequence[Finding],
    written: Sequence[str] | None,
    output: Literal["text", "json"],
    cwd: str,
    family: tuple[str, tuple[str, ...]] | None = None,
) -> str:
    """Renders the report; family holds the bundle argument and the members of a family."""
    errors = sum(1 for finding in findings if finding.severity == "error")
    warnings = len(findings) - errors
    if output == "json":
        report: dict[str, object] = {
            "findings": [
                {
                    "path": display_path(finding.path, cwd),
                    "line": finding.line,
                    "severity": finding.severity,
                    "message": finding.message,
                }
                for finding in findings
            ],
            "summary": {"findings": len(findings), "errors": errors, "warnings": warnings},
        }
        if written is not None:
            report["written"] = list(written)
        if family is not None:
            report["bundles"] = list(family[1])
        text = json.dumps(report, indent=2, ensure_ascii=False)
        return _LONE_SURROGATE.sub(lambda m: f"\\u{ord(m.group(0)):04x}", text) + "\n"
    if family is not None and not family[1]:
        return f"{family[0]} is a family without bundles: no child directory holds log/\n"
    lines = [
        f"{display_path(finding.path, cwd)}:{finding.line}: {finding.severity}: {finding.message}"
        for finding in findings
    ]
    summary = f"{plural(len(findings), 'finding')}: {plural(errors, 'error')}, {plural(warnings, 'warning')}"
    lines.append(summary if written is None else f"{summary}; wrote {plural(len(written), 'file')}")
    return "\n".join(lines) + "\n"


def exit_status(findings: Iterable[Finding]) -> int:
    return 1 if any(finding.severity == "error" for finding in findings) else 0


# ---------------------------------------------------------------------------
# File system edge
# ---------------------------------------------------------------------------


def _read_text(path: str) -> str:
    with open(path, "rb") as handle:
        return handle.read().decode("utf-8", errors="replace")


def _well_formed(text: str) -> bytes:
    # Node writes a lone surrogate as U+FFFD, and so does this edge.
    return _LONE_SURROGATE.sub("�", text).encode("utf-8")


def read_bundle_source(root: str) -> BundleSource:
    """Reads every markdown file of a bundle.

    Hidden entries are skipped and symbolic links are reported rather than followed, so the result
    depends only on the files inside the bundle.
    """
    directories: list[str] = []
    files: list[tuple[str, str]] = []
    links: list[str] = []
    pending = [""]
    while pending:
        directory = pending.pop()
        directories.append(directory)
        with os.scandir(os.path.join(root, directory)) as entries:
            for entry in entries:
                if entry.name.startswith("."):
                    continue
                path = entry.name if directory == "" else f"{directory}/{entry.name}"
                if entry.is_symlink():
                    links.append(path)
                elif entry.is_dir(follow_symlinks=False):
                    pending.append(path)
                elif entry.is_file(follow_symlinks=False) and entry.name.endswith(".md"):
                    files.append((path, _read_text(entry.path)))
    return BundleSource(tuple(files), tuple(directories), tuple(links))


def read_listing(root: str) -> tuple[DirectoryEntry, ...]:
    entries: list[DirectoryEntry] = []
    with os.scandir(root) as listing:
        for entry in listing:
            if entry.is_dir(follow_symlinks=False):
                with os.scandir(entry.path) as inner:
                    subdirectories = tuple(child.name for child in inner if child.is_dir(follow_symlinks=False))
                entries.append(DirectoryEntry(entry.name, "directory", subdirectories))
            elif entry.is_file(follow_symlinks=False):
                entries.append(DirectoryEntry(entry.name, "file"))
            else:
                entries.append(DirectoryEntry(entry.name, "other"))
    return tuple(entries)


def default_schema_directories(skills_directory: str) -> tuple[str, ...]:
    """The assets/schemas directory of every skill next to this one, so an owner skill's schema
    needs no registration."""
    found: list[str] = []
    with os.scandir(skills_directory) as entries:
        for entry in entries:
            if entry.is_dir(follow_symlinks=False) and not entry.name.startswith("."):
                path = os.path.join(skills_directory, entry.name, "assets", "schemas")
                if os.path.isdir(path):
                    found.append(path)
    return tuple(sorted(found, key=text_key))


def read_schema_sources(directories: Sequence[str], defaults: Sequence[str]) -> tuple[SchemaSource, ...]:
    sources: list[SchemaSource] = []
    for tier, group in ((0, directories), (1, defaults)):
        for directory in group:
            names = sorted((name for name in os.listdir(directory) if name.endswith(".schema.json")), key=text_key)
            for name in names:
                path = os.path.normpath(os.path.join(directory, name))
                if os.path.isfile(path):
                    sources.append(SchemaSource(path, _read_text(path), tier))
    return tuple(sources)


def apply_run(root: str, run: BundleRun) -> tuple[str, ...]:
    """Writes and deletes the files of a --write run and returns the paths it changed."""
    written: list[str] = []
    for path, content in run.writes:
        with open(os.path.join(root, path), "wb") as handle:
            handle.write(_well_formed(content))
        written.append(path)
    for path in run.deletes:
        target = os.path.join(root, path)
        # lstat keeps a symbolic link named index.md from removing its target.
        if os.path.isfile(target) and not os.path.islink(target):
            os.unlink(target)
            written.append(path)
    return tuple(written)


def _emit(stream: object, text: str) -> None:
    buffer = getattr(stream, "buffer")
    buffer.write(_well_formed(text))
    buffer.flush()


def main(argv: Sequence[str]) -> int:
    cwd = os.getcwd()
    parsed = parse_arguments(argv)
    if isinstance(parsed, HelpRequest):
        _emit(sys.stdout, USAGE)
        return 0
    if isinstance(parsed, UsageError):
        _emit(sys.stderr, f"plan-bundle: {parsed.message}\n\n{USAGE}")
        return 2
    root = os.path.normpath(os.path.join(cwd, parsed.bundle))
    if not os.path.isdir(root):
        _emit(sys.stderr, f"plan-bundle: bundle directory not found: {parsed.bundle}\n")
        return 2
    schema_directories = [os.path.normpath(os.path.join(cwd, directory)) for directory in parsed.schemas]
    missing = next((given for given, path in zip(parsed.schemas, schema_directories, strict=True) if not os.path.isdir(path)), None)
    if missing is not None:
        _emit(sys.stderr, f"plan-bundle: schema directory not found: {missing}\n")
        return 2
    registry = EMPTY_REGISTRY
    if parsed.command == "validate":
        skills = os.path.normpath(os.path.join(os.path.dirname(os.path.realpath(__file__)), "..", ".."))
        registry = build_schema_registry(read_schema_sources(schema_directories, default_schema_directories(skills)))
    members = family_members(read_listing(root))
    targets = (("", root),) if members is None else tuple((name, os.path.join(root, name)) for name in members)
    findings: list[Finding] = []
    written: list[str] = []
    writing = False
    for member, path in targets:
        run = run_bundle(parsed, load_bundle(read_bundle_source(path)), registry)
        changed = apply_run(path, run) if run.writing else ()
        writing = writing or run.writing
        findings.extend(run.findings if member == "" else (prefix_finding(f, member) for f in run.findings))
        written.extend(changed if member == "" else (f"{member}/{p}" for p in changed))
    if members is not None and parsed.mode == "write":
        writing = True
    result = finish(findings)
    family = None if members is None else (parsed.bundle, members)
    _emit(sys.stdout, render_report(result, written if writing else None, parsed.format, cwd, family))
    return exit_status(result)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
