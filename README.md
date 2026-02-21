# wav2mp3

Конвертер WAV → MP3 с максимальным качеством. Использует [libmp3lame](https://lame.sourceforge.io/) через CGo с поддержкой VBR, ID3v2-тегов и обложки альбома.

## Требования

- macOS (Apple Silicon или Intel)
- Go 1.21+
- libmp3lame:

```bash
brew install lame
```

## Установка

```bash
git clone ...
cd svk-wav2mp3
make install        # собирает и копирует в /usr/local/bin/wav2mp3
```

Или только сборка:

```bash
make build          # ./bin/wav2mp3
```

## Использование

```
wav2mp3 -i INPUT [flags]
```

### Примеры

```bash
# VBR V2 по умолчанию — лучший баланс качество/размер
wav2mp3 -i song.wav

# Указать выходной файл
wav2mp3 -i song.wav -o song_hq.mp3

# С тегами и обложкой
wav2mp3 -i song.wav \
  --title "Song Title" \
  --artist "Artist Name" \
  --album "Album" \
  --year "2026" \
  --genre "Electronic" \
  --track "3" \
  --cover cover.jpg

# CBR 320 kbps
wav2mp3 -i song.wav --bitrate 320

# VBR максимальное качество (V0)
wav2mp3 -i song.wav --vbr-quality 0

# Тихий режим (без прогресс-бара и статистики)
wav2mp3 -i song.wav -q

# Подробный вывод (параметры энкодера)
wav2mp3 -i song.wav -v
```

### Флаги

| Флаг | По умолчанию | Описание |
|------|--------------|----------|
| `-i, --input` | — | Входной WAV файл **(обязательный)** |
| `-o, --output` | рядом с входным | Выходной MP3 файл |
| `--title` | | Название трека |
| `--artist` | | Исполнитель |
| `--album` | | Альбом |
| `--year` | | Год |
| `--genre` | | Жанр |
| `--track` | | Номер трека |
| `--comment` | | Комментарий |
| `--cover` | | Обложка (JPEG, PNG, GIF, WebP) |
| `--vbr-quality` | `2.0` | VBR качество 0.0 (лучше) – 9.9; несовместим с `--bitrate` |
| `--bitrate` | — | CBR битрейт kbps (32–320); при указании отключает VBR |
| `--quality` | `2` | Алгоритмическое качество LAME 0 (лучше) – 9 |
| `-v, --verbose` | | Подробный вывод |
| `-q, --quiet` | | Без прогресс-бара и статистики |
| `--version` | | Версия |

### Поддерживаемые форматы WAV

| Формат | Каналы |
|--------|--------|
| 8-bit PCM | Mono, Stereo |
| 16-bit PCM | Mono, Stereo |
| 24-bit PCM | Mono, Stereo |
| 32-bit PCM | Mono, Stereo |

## Вывод

```
Вход:  song.wav (44100 Hz, Stereo, 16-bit, 3m 42s, 39.1 MB)
Выход: song.mp3 (VBR V2, elapsed 12.3s, 8.4 MB, сжатие 4.65x)
Теги:  Title="Song Title", Artist="Artist Name", Cover=cover.jpg
```

## Разработка

```bash
make test           # все тесты
make testdata       # регенерировать WAV-фикстуры
make fmt            # go fmt
```

## Лицензия

MIT
