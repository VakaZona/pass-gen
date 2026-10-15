# pass-gen

Генератор случайных паролей для macOS и Linux на Go. Каждый запуск выводит новый пароль и сразу копирует его в буфер обмена. Пароли не сохраняются; в конфиге находятся только настройки. Буфер обмена содержит пароль до следующего изменения буфера, а вывод остаётся в истории терминала.

## Сборка и запуск

Требуется Go 1.25 или новее.

```sh
make build
./bin/pass-gen
```

Чтобы вызывать команду из любого каталога, добавьте каталог `bin` проекта в `PATH` (для текущего терминала):

```sh
export PATH="/Users/valery/Projects/password-mac/bin:$PATH"
pass-gen
```

Для постоянного действия добавьте строку `export` в `~/.zshrc`.

## Настройки

```sh
pass-gen config
pass-gen config set length 24
pass-gen config set symbols false
pass-gen config set exclude_ambiguous false
pass-gen config path
pass-gen config reset
```

По умолчанию: длина 20, строчные и заглавные латинские буквы, цифры и спецсимволы. Неоднозначные символы `i l I o O 0 1` исключены. Спецсимволы: `!@#$%^&*()-_=+[]{};:,.?/~`.

Настройки хранятся в `~/.config/pass-gen/config.json`, либо в `$XDG_CONFIG_HOME/pass-gen/config.json`, если переменная задана абсолютным путём. Файл создаётся при сохранении настройки или сбросе; без файла используются значения по умолчанию. Изменения вступают в силу при следующем запуске.

Ключи: `length`, `uppercase`, `lowercase`, `digits`, `symbols`, `exclude_ambiguous`. Для логических значений используйте `true` или `false`. Длина: от 1 до 4096, не меньше числа включённых категорий; должна быть включена хотя бы одна категория. Каждый пароль содержит хотя бы один символ каждой включённой категории. Генерация использует `crypto/rand` с равномерной выборкой среди паролей, удовлетворяющих настройкам.

Параметры команды переопределяют настройки только для текущего запуска:

```sh
pass-gen --length 32 --symbols=false
pass-gen --no-copy
pass-gen --help
```

Для булевых параметров команды нужен синтаксис `--symbols=false`, а не `--symbols false`. При ошибке копирования пароль уже выведен, сообщение об ошибке поступает в stderr, код завершения — 1. При успешном запуске stdout содержит только пароль и перевод строки; в буфер копируется пароль без перевода строки.

## Проверка

```sh
make test
```

Тесты проверяют состав паролей, ограничения настроек, сохранение и сброс, временные переопределения, повреждённый конфиг и ошибки буфера обмена.

## Установка из GitHub Releases

После публикации первого релиза с установщиком:

```sh
curl -fsSL https://github.com/VakaZona/pass-gen/releases/latest/download/install.sh | sh
```

Установщик выбирает macOS/Linux и ARM64/x86_64, скачивает готовый бинарник, проверяет SHA-256 и устанавливает его в `~/.local/bin/pass-gen`. Go, Git и Homebrew на целевом компьютере не нужны. Нужен `curl` и `shasum` либо `sha256sum`. Для zsh и bash установщик добавляет этот каталог в PATH, если его ещё нет. После первой установки открой новый терминал:

```sh
pass-gen --version
pass-gen
```

Повторный запуск установки обновляет программу и сохраняет настройки. Можно скачать `install.sh` из релиза, прочитать его и выполнить `sh install.sh` вместо запуска через pipe. Для конкретной версии:

```sh
curl -fsSL https://github.com/VakaZona/pass-gen/releases/download/v0.1.1/install.sh | PASS_GEN_VERSION=v0.1.1 sh
```

Переменная `PASS_GEN_INSTALL_DIR` задаёт другой каталог установки; для него PATH нужно настроить самостоятельно. `PASS_GEN_NO_PATH=1` отключает изменение настроек оболочки.

### Буфер обмена на Linux

В графической сессии Wayland установи `wl-clipboard` (команда `wl-copy`), в X11 — `xclip` или `xsel`. Например, на Ubuntu/Debian:

```sh
sudo apt install wl-clipboard  # Wayland
sudo apt install xclip         # X11, вместо wl-clipboard
```

На сервере без графической сессии используй `pass-gen --no-copy`. При отсутствии доступного буфера пароль будет выведен, а команда сообщит об ошибке копирования.

## Публикация новой версии

В проект добавлен workflow `.github/workflows/release.yml`: после push тега он запускает тесты, собирает четыре бинарника и публикует GitHub Release с установщиком и контрольными суммами. Используется встроенный GitHub Actions token; дополнительный токен не нужен.

Опубликуй изменения и новый тег (не переиспользуй уже существующий `v0.1.0`):

```sh
git add main.go README.md install.sh scripts .github
git commit -m "Add GitHub release builds and installer for macOS and Linux"
git push origin main
git tag v0.1.1
git push origin v0.1.1
```

Дождись успешного workflow в https://github.com/VakaZona/pass-gen/actions и проверь файлы в https://github.com/VakaZona/pass-gen/releases. После этого команда установки выше работает на другом компьютере. Репозиторий должен быть публичным, GitHub Actions — включён.

Локальная сборка release-файлов: `sh scripts/build-release.sh v0.1.1`. Поддерживаются стабильные версии с тегами `vMAJOR.MINOR.PATCH`.
