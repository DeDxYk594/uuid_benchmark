# PostgreSQL Benchmark на Go

Бенчмарк отвечает на вопрос: что лучше: хранить UUID файла как PRIMARY KEY или хранить UUID файла как атрибут и ставить автоинкрементный суррогатный PRIMARY KEY меньшей разрядности?

Тестируются две модели данных:

```sql
-- Поддиректория uuid
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "user"(
    u_id BIGINT PRIMARY KEY,
    avatar_file_uuid UUID
);
CREATE TABLE "file"(
    file_uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_extension TEXT,
    created_at TIMESTAMPTZ
);
ALTER TABLE "user" ADD CONSTRAINT user_file_fk FOREIGN KEY (avatar_file_uuid) REFERENCES "file"(file_uuid) ON UPDATE CASCADE ON DELETE CASCADE;

CREATE INDEX idx_user_avatar_file_uuid ON "user"(avatar_file_uuid);
```

```sql
-- Поддиректория bigint
CREATE TABLE "user"(
    u_id BIGINT PRIMARY KEY,
    avatar_file_id BIGINT,
    FOREIGN KEY (avatar_file_id) REFERENCES "file"(file_id) ON UPDATE CASCADE ON DELELE CASCADE
);
CREATE TABLE "file"(
    file_id BIGINT PRIMARY KEY ALWAYS GENERATED AS IDENTITY,
    file_uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_extension TEXT,
    created_at TIMESTAMPTZ,
    created_by
);
```

## Описание

Бенчмарк состоит из двух этапов:

#### Лютая вставка

- 15 параллельных горутин по КД инсёртят юзеров. `u_id` присваивается прямо из гошки и аватарка в виде записи в таблице `file`
- Для автоинкрементации `u_id` используется глобальная переменная, доступ к которой регулируется через `mutex` (чтобы можно было точно сказать, что ключ в диапазоне от 1 до N точно есть в таблице юзеров)
- Одна дополнительная горутина выполняет `SELECT` с использованием `JOIN`
- Не забываем, что у нас есть индекс по `FOREIGN KEY`
- В процессе выполнения бенчмарка добавляется 100K пользователей
- Результаты измерения задержек для `INSERT` и `SELECT` записываются в файлы `INSERT_1_latency.txt` и `SELECT_1_latency.txt` соответственно

#### Сумасшедший селект (условия, близкие к проду)

- 15 горутин по КД делают `SELECT` (с `JOIN` для получения файла аватарки)
- Одна дополнительная горутина делает `INSERT`
- Выполняется 1M селектов
- Результаты измерения задержек для `INSERT` и `SELECT` записываются в файлы `INSERT_2_latency.txt` и `SELECT_2_latency.txt` соответственно

Программа будет выводить прогресс выполнения этапов и сохранять данные о задержек в текстовые файлы.

# Результаты

ОС: Windows 11 Pro x64

Intel Core i5-12500H

16 GB DDR4

SSD 1 TB XPG GAMMIX S11 Pro

PostgreSQL 17

|Бенчмарк|Время первой части|Время второй части|
|-|-|-|
|`UUID`|38 с|403 с|
|`BIGNIT`|38 с|390 с|
