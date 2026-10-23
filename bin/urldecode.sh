#!/bin/bash

# НАСТРОЙКА ЦВЕТОВ (Проверяем поддержку терминалом)
if [ -t 1 ] && command -v tput &> /dev/null && [ "$(tput colors)" -ge 8 ]; then
    COLOR_GREEN=$(tput setaf 2)
    COLOR_RED=$(tput setaf 1)
    COLOR_GRAY=$(tput setaf 8)
    COLOR_RESET=$(tput sgr0)
else
    COLOR_GREEN=""
    COLOR_RED=""
    COLOR_GRAY=""
    COLOR_RESET=""
fi

# Путь к программе декодирования
GO_BINARY="${URLDECODE_BIN:-urldecode}"
if ! command -v "$GO_BINARY" &> /dev/null; then
    echo "Ошибка: Не удается вызвать '$GO_BINARY'." >&2
    echo "Причина: файл отсутствует в PATH, не создан или не имеет прав на выполнение (chmod +x)." >&2
    exit 1
fi

# Собираем все аргументы скрипта в одну строку
INPUT_TEXT="$*"
if [ $# -eq 0 ]; then
    echo "Ошибка: Не передан текст для декодирования." >&2
    echo "Использование: $0 <закодированный_url>" >&2
    exit 1
fi

# Выбираем утилиту для работы с буфером обмена
if command -v pbcopy &> /dev/null; then
    # Для macOS
    CLIP_CMD="pbcopy"
elif command -v xclip &> /dev/null; then
    # Для Linux (X11)
    CLIP_CMD="xclip -selection clipboard"
elif command -v wl-copy &> /dev/null; then
    # Для Linux (Wayland)
    CLIP_CMD="wl-copy"
else
    echo "Ошибка: Не найдена утилита буфера обмена (pbcopy, xclip или wl-copy)." >&2
    exit 1
fi

# Временный файл для ошибки
STDERR_FILE=$(mktemp)
trap 'rm -f "$STDERR_FILE"' EXIT

# Передаем текст в Go-программу и копируем результат
set -o pipefail
DECODED_TEXT=$(echo -n "$INPUT_TEXT" | "$GO_BINARY" 2> "$STDERR_FILE")
DECODE_STATUS=$?
ERROR_TEXT=$(cat "$STDERR_FILE")
if [ $DECODE_STATUS -eq 0 ]; then
    # Выводим результат в терминал
    echo ""
    echo "${COLOR_GREEN}$DECODED_TEXT${COLOR_RESET}"
    
    # Отправляем результат в буфер обмена
    echo -n "$DECODED_TEXT" | $CLIP_CMD
    
    echo -e "\n${COLOR_GRAY}Текст скопирован в буфер обмена.${COLOR_RESET}"
else
    echo -e "${COLOR_RED}Ошибка при выполнении декодирования.${COLOR_RESET}" >&2
    echo "$ERROR_TEXT" >&2
    exit 1
fi
