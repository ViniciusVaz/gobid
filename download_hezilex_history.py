#!/usr/bin/env python3
"""Baixa o histórico paginado da sua conta Hezilex em um arquivo JSON.

Instalação:
    python3 -m pip install requests

Antes de executar, preencha XSRF_TOKEN e LARAVEL_SESSION com os valores
atuais da sua própria sessão autenticada no navegador. Não compartilhe esses
valores: eles podem permitir acesso à sua conta enquanto estiverem válidos.
"""

from __future__ import annotations

import json
import random
import sys
import time
from pathlib import Path
from typing import Any

import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry


BASE_URL = "https://app.hezilex.com/binary/history"
OUTPUT_FILE = Path("history.json")
START_PAGE = 1

# Pausa aleatória entre páginas para não enviar chamadas em frequência fixa.
MIN_DELAY_SECONDS = 1.5
MAX_DELAY_SECONDS = 3.5

REQUEST_TIMEOUT_SECONDS = 30
MAX_RETRIES = 5
BACKOFF_FACTOR = 1.5

# Preencha somente com credenciais da sua sessão. Não cole cookies neste arquivo
# se ele for ser compartilhado ou enviado para um repositório.
XSRF_TOKEN = "eyJpdiI6IjAyTTZrNVdBL3l6K1QwVWV0Ujk5ZUE9PSIsInZhbHVlIjoidU9zQ0xCUE1EWVBjdmdRdURlaHpqMkQzbEoyWGxlL25qOEN3cHVVa01aVU1nT283ZXdndm5yZmZMME14OU15b0FONEVWZ1htdFZkN3RhRVJSQ0hOUlZob3l0VmxhVjBLa0wwWGtwZXZ0ZnRVOFVjTFFkb1RRMy9NSmpZdloyMDciLCJtYWMiOiIwMjQ2ZjJkNTZhYTdlMjhmNTM3MjA5MmYxMTJkZTYzODJlY2Y1MzhiMTU3YWZiMjJlZjc5ODhlMTQ5Y2UzYjgyIiwidGFnIjoiIn0"
LARAVEL_SESSION = "eyJpdiI6InIyaEkyMitoSDdJRTVtMzc5UE1mTGc9PSIsInZhbHVlIjoiTWtMM0dTd3pCQ1RuR3N6V0VZVVd3RkJOQXhQUVZmT1YxcXJMVzIvVTJlZithMG1heEVmekZsOHNYQzZYV1p2VTdiRzBJVDdnckU3K2pUK2xrbkczaHJIRVNPOGhFV2NncmJ2NUlEVjhTNXFrc2xIY3B4eDQ1dUh4aTBQeUVBVWEiLCJtYWMiOiJmZmE5M2M0NDcwMzRlOWQ0YWU0Y2M3NGE2MjJmN2YyZWI1MzdkMzcyNjg5MWIzZjA2MTliYWNmNGMzM2M2ZWI2IiwidGFnIjoiIn0%3D;"

HEADERS = {
    "Accept": "application/json, text/plain, */*",
    "Accept-Language": "pt-BR,pt;q=0.9",
    "Referer": "https://app.hezilex.com/traderoom",
    "User-Agent": (
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
        "AppleWebKit/605.1.15 (KHTML, like Gecko) "
        "Version/26.4 Safari/605.1.15"
    ),
    # Em aplicações Laravel, este normalmente é o mesmo valor do cookie XSRF,
    # já decodificado. Se a sua chamada no navegador usa outro valor, substitua.
    "X-XSRF-TOKEN": XSRF_TOKEN,
}

COOKIES = {
    "XSRF-TOKEN": XSRF_TOKEN,
    "laravel_session": LARAVEL_SESSION,
}


def validate_credentials() -> None:
    """Evita uma execução acidental usando os valores de exemplo."""
    placeholders = ("COLE_AQUI", "SEU_", "TOKEN_AQUI")
    if any(marker in XSRF_TOKEN for marker in placeholders) or any(
        marker in LARAVEL_SESSION for marker in placeholders
    ):
        sys.exit(
            "Preencha XSRF_TOKEN e LARAVEL_SESSION com os valores atuais da sua sessão "
            "antes de executar o script."
        )


def build_session() -> requests.Session:
    """Cria uma sessão com repetição automática para falhas temporárias."""
    retry = Retry(
        total=MAX_RETRIES,
        connect=MAX_RETRIES,
        read=MAX_RETRIES,
        status=MAX_RETRIES,
        backoff_factor=BACKOFF_FACTOR,
        status_forcelist=(429, 500, 502, 503, 504),
        allowed_methods=frozenset({"GET"}),
        respect_retry_after_header=True,
        raise_on_status=False,
    )
    adapter = HTTPAdapter(max_retries=retry)
    session = requests.Session()
    session.mount("https://", adapter)
    session.headers.update(HEADERS)
    session.cookies.update(COOKIES)
    return session


def page_items(payload: Any) -> list[Any]:
    """Extrai itens de respostas que sejam listas ou objetos com campo data."""
    if isinstance(payload, list):
        return payload
    if isinstance(payload, dict):
        data = payload.get("data")
        if isinstance(data, list):
            return data
    raise ValueError(
        "Formato de resposta inesperado. Esperava uma lista JSON ou um objeto com 'data'."
    )


def save_json(items: list[Any], path: Path) -> None:
    """Salva de forma atômica, preservando um JSON válido se houver interrupção."""
    temporary_path = path.with_suffix(path.suffix + ".tmp")
    with temporary_path.open("w", encoding="utf-8") as file:
        json.dump(items, file, ensure_ascii=False, indent=2)
        file.write("\n")
    temporary_path.replace(path)


def fetch_history() -> list[Any]:
    validate_credentials()
    collected: list[Any] = []

    with build_session() as session:
        page = START_PAGE
        while True:
            url = f"{BASE_URL}/{page}"
            print(f"Buscando página {page}: {url}")

            try:
                response = session.get(url, timeout=REQUEST_TIMEOUT_SECONDS)
                response.raise_for_status()
                payload = response.json()
                items = page_items(payload)
            except requests.RequestException as error:
                save_json(collected, OUTPUT_FILE)
                raise RuntimeError(
                    f"Falha ao buscar a página {page}. O progresso foi salvo em "
                    f"{OUTPUT_FILE.resolve()}."
                ) from error
            except (json.JSONDecodeError, ValueError) as error:
                save_json(collected, OUTPUT_FILE)
                raise RuntimeError(
                    f"Resposta inválida na página {page}. O progresso foi salvo em "
                    f"{OUTPUT_FILE.resolve()}."
                ) from error

            if not items:
                print(f"Página {page} está vazia. Paginação concluída.")
                break

            collected.extend(items)
            save_json(collected, OUTPUT_FILE)
            print(f"  {len(items)} itens recebidos; total salvo: {len(collected)}")

            page += 1
            delay = random.uniform(MIN_DELAY_SECONDS, MAX_DELAY_SECONDS)
            print(f"  Aguardando {delay:.1f}s antes da próxima página...")
            time.sleep(delay)

    return collected


if __name__ == "__main__":
    try:
        history = fetch_history()
    except RuntimeError as error:
        print(f"Erro: {error}", file=sys.stderr)
        sys.exit(1)

    print(f"Concluído: {len(history)} itens salvos em {OUTPUT_FILE.resolve()}")
