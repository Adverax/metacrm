# IAM Permission Sync Service

## Обзор

IAM Permission Sync Service предоставляет gRPC API для внешних сервисов, которые нуждаются в синхронизации разрешений пользователей через механизм lazy loading. Этот сервис особенно полезен для сервисов, которые получают события от IAM (например, добавление пользователя в группу) и должны обновить свои локальные кэши разрешений.

## Архитектура

```
┌─────────────────┐    Events     ┌─────────────────┐    gRPC     ┌─────────────────┐
│   IAM Service   │ ────────────► │ External Service│ ──────────► │ Permission Sync │
│                 │               │                 │             │     Service     │
└─────────────────┘               └─────────────────┘             └─────────────────┘
        │                                  │                                │
        │                                  │                                │
        ▼                                  ▼                                ▼
┌─────────────────┐               ┌─────────────────┐             ┌─────────────────┐
│   Outbox Table  │               │ Local Cache     │             │  Database       │
│   (Events)      │               │ (Permissions)   │             │  Functions      │
└─────────────────┘               └─────────────────┘             └─────────────────┘
```

## Основные возможности

### 1. Получение снимка разрешений пользователя
- **Метод**: `GetUserPermissionsSnapshot`
- **Назначение**: Получить полный снимок разрешений пользователя для кэширования
- **Использование**: Когда внешний сервис получает событие изменения разрешений

### 2. Получение изменений разрешений
- **Метод**: `GetUserPermissionsChanges`
- **Назначение**: Получить изменения разрешений пользователя с определенного момента
- **Использование**: Для инкрементальной синхронизации

### 3. Массовое получение разрешений
- **Метод**: `GetBulkUserPermissions`
- **Назначение**: Получить разрешения для нескольких пользователей одновременно
- **Использование**: Для пакетной обработки (максимум 100 пользователей)

### 4. Проверка изменений
- **Метод**: `CheckUserPermissionsChanged`
- **Назначение**: Эффективная проверка наличия изменений разрешений
- **Использование**: Для оптимизации - избежать ненужных запросов

### 5. Синхронизация разрешений группы
- **Метод**: `SyncGroupPermissions`
- **Назначение**: Синхронизировать разрешения для всех участников группы
- **Использование**: После изменения разрешений группы

## Типичные сценарии использования

### Сценарий 1: Добавление пользователя в группу

```
1. IAM Service: Создает событие iam.group_member.added
2. External Service: Получает событие из outbox
3. External Service: Вызывает GetUserPermissionsSnapshot(user_id)
4. External Service: Обновляет локальный кэш разрешений
5. External Service: Обслуживает запросы с обновленными разрешениями
```

### Сценарий 2: Изменение разрешений группы

```
1. IAM Service: Создает событие iam.permission_set.assigned_to_group
2. External Service: Получает событие
3. External Service: Вызывает SyncGroupPermissions(group_id)
4. External Service: Получает обновленные разрешения для всех участников
5. External Service: Обновляет локальный кэш
```

### Сценарий 3: Периодическая синхронизация

```
1. External Service: Вызывает CheckUserPermissionsChanged(user_id, since)
2. Если есть изменения:
   - Вызывает GetUserPermissionsChanges для получения деталей
   - Обновляет локальный кэш
3. Если изменений нет - использует существующий кэш
```

## Структуры данных

### UserPermissionsSnapshot
```protobuf
message UserPermissionsSnapshot {
  string user_id = 1;                      // ID пользователя
  string tenant_id = 2;                    // ID тенанта
  repeated ObjectPermission permissions = 3; // Разрешения на объекты
  repeated FieldPermission field_permissions = 4; // Разрешения на поля
  repeated GroupMembership group_memberships = 5; // Членство в группах
  repeated PermissionSource permission_sources = 6; // Источники разрешений
  google.protobuf.Timestamp snapshot_at = 7; // Время создания снимка
  string snapshot_version = 8;             // Версия снимка
}
```

### PermissionBitmask
```protobuf
message PermissionBitmask {
  int32 value = 1;                         // Сырое значение битовой маски
  bool can_read = 2;                       // Разрешение READ (бит 0)
  bool can_create = 3;                     // Разрешение CREATE (бит 1)
  bool can_update = 4;                     // Разрешение UPDATE (бит 2)
  bool can_delete = 5;                     // Разрешение DELETE (бит 3)
  repeated string permission_names = 6;    // Человекочитаемые названия разрешений
}
```

## Конфигурация и настройка

### Кэширование
- **TTL по умолчанию**: 3600 секунд (1 час)
- **Настраиваемый TTL**: Можно указать для каждого запроса
- **Принудительное обновление**: Доступно через параметр `force_refresh`

### Лимиты
- **Массовые запросы**: Максимум 100 пользователей за раз
- **Таймауты**: Настраиваемые таймауты для gRPC вызовов
- **Ретраи**: Встроенная логика повторных попыток

## Мониторинг и аналитика

### Метрики
- Общее количество событий
- Количество событий синхронизации
- Количество неудачных событий
- Среднее время обработки

### Логирование
- Все вызовы API логируются
- Отслеживание производительности
- Мониторинг ошибок

## Примеры использования

### Go клиент
```go
// Подключение к сервису
conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := pb.NewPermissionSyncServiceClient(conn)

// Получение снимка разрешений
resp, err := client.GetUserPermissionsSnapshot(ctx, &pb.GetUserPermissionsSnapshotRequest{
    TenantId: "550e8400-e29b-41d4-a716-446655440000",
    UserId:   "usr_a1b2c3d4e5f67890",
    ObjectApiNames: []string{"user", "order", "product"},
    IncludeGroupMemberships: true,
    TtlSeconds: &[]int32{3600}[0],
})
```

### Обработка событий
```go
// При получении события iam.group_member.added
func handleGroupMemberAdded(event GroupMemberAddedEvent) {
    // Получить обновленные разрешения пользователя
    snapshot, err := client.GetUserPermissionsSnapshot(ctx, &pb.GetUserPermissionsSnapshotRequest{
        TenantId: event.TenantId,
        UserId:   event.MemberId,
        ObjectApiNames: []string{"user", "order"}, // Только нужные объекты
        IncludeGroupMemberships: true,
    })
    
    if err != nil {
        log.Printf("Failed to get user permissions: %v", err)
        return
    }
    
    // Обновить локальный кэш
    updateLocalPermissionCache(snapshot)
}
```

## Интеграция с базой данных

### Функции PostgreSQL
- `iam.get_user_permissions_snapshot()` - Получение снимка разрешений
- `iam.get_user_permissions_changes()` - Получение изменений
- `iam.get_bulk_user_permissions()` - Массовое получение
- `iam.sync_group_permissions()` - Синхронизация группы
- `iam.check_user_permissions_changed()` - Проверка изменений

### Оптимизация
- Использование индексов для быстрого поиска
- Кэширование результатов на уровне базы данных
- Пакетная обработка для снижения нагрузки

## Безопасность

### Аутентификация
- Все запросы требуют валидный tenant_id
- Проверка прав доступа на уровне базы данных
- Аудит всех операций

### Валидация
- Проверка формата UUID для tenant_id
- Валидация существования пользователей
- Проверка лимитов запросов

## Производительность

### Оптимизации
- Кэширование результатов
- Инкрементальные обновления
- Пакетная обработка
- Асинхронная обработка событий

### Масштабирование
- Горизонтальное масштабирование gRPC серверов
- Распределенное кэширование
- Балансировка нагрузки

## Troubleshooting

### Частые проблемы

1. **Таймауты запросов**
   - Увеличить таймауты gRPC
   - Проверить производительность базы данных

2. **Неполные данные**
   - Проверить параметры запроса
   - Убедиться в корректности tenant_id и user_id

3. **Медленная работа**
   - Проверить индексы базы данных
   - Оптимизировать запросы
   - Использовать кэширование

### Логи и отладка
- Включить детальное логирование
- Мониторить метрики производительности
- Анализировать медленные запросы

## Roadmap

### Планируемые улучшения
- Поддержка WebSocket для real-time обновлений
- GraphQL API для гибких запросов
- Машинное обучение для предсказания изменений разрешений
- Автоматическая оптимизация кэша

### Интеграции
- Kafka для событий
- Redis для кэширования
- Prometheus для метрик
- Jaeger для трейсинга
