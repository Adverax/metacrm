# IAM Permissions gRPC Service

## Обзор

gRPC сервис для получения информации о правах доступа пользователей в IAM системе. Предоставляет быстрый доступ к информации о:
- Object permissions (разрешения на объекты)
- Field permissions (разрешения на поля)
- Group memberships (членство в группах)

## Особенности

- **Multi-tenant архитектура** - все запросы изолированы по tenant_id
- **Кэширование** - автоматическое кэширование с настраиваемым TTL
- **Batch операции** - массовое получение разрешений
- **Иерархические разрешения** - поддержка наследования от групп и ролей
- **Field-level security** - детальные разрешения на уровне полей

## Сервисы

### PermissionService

Основной сервис для работы с разрешениями пользователей.

#### Методы

##### 1. GetUserObjectPermissions
Получить разрешения пользователя на конкретный объект.

```protobuf
rpc GetUserObjectPermissions(GetUserObjectPermissionsRequest) returns (GetUserObjectPermissionsResponse);
```

**Запрос:**
- `tenant_id` - идентификатор тенанта (UUID)
- `user_id` - идентификатор пользователя (например, "usr_a1b2c3d4e5f67890")
- `object_api_name` - API имя объекта (например, "user", "order", "product")
- `ttl_seconds` - TTL кэша в секундах (по умолчанию: 3600)

**Ответ:**
- `ObjectPermission` - детали разрешений на объект
- `CacheInfo` - информация о кэше

##### 2. GetUserFieldPermissions
Получить разрешения пользователя на конкретное поле объекта.

```protobuf
rpc GetUserFieldPermissions(GetUserFieldPermissionsRequest) returns (GetUserFieldPermissionsResponse);
```

**Запрос:**
- `tenant_id` - идентификатор тенанта (UUID)
- `user_id` - идентификатор пользователя
- `object_api_name` - API имя объекта
- `field_api_name` - API имя поля (например, "email", "salary")
- `ttl_seconds` - TTL кэша в секундах

**Ответ:**
- `FieldPermission` - детали разрешений на поле
- `CacheInfo` - информация о кэше

##### 3. GetUserGroupMemberships
Получить список групп, в которых состоит пользователь.

```protobuf
rpc GetUserGroupMemberships(GetUserGroupMembershipsRequest) returns (GetUserGroupMembershipsResponse);
```

**Запрос:**
- `tenant_id` - идентификатор тенанта (UUID)
- `user_id` - идентификатор пользователя
- `include_inherited` - включить наследованные членства (по умолчанию: true)
- `include_nested` - включить вложенные группы (по умолчанию: true)

**Ответ:**
- `GroupMembership[]` - список членств в группах
- `total_count` - общее количество групп

##### 4. CheckUserObjectPermission
Проверить, есть ли у пользователя конкретное разрешение на объект.

```protobuf
rpc CheckUserObjectPermission(CheckUserObjectPermissionRequest) returns (CheckUserObjectPermissionResponse);
```

**Запрос:**
- `tenant_id` - идентификатор тенанта (UUID)
- `user_id` - идентификатор пользователя
- `object_api_name` - API имя объекта
- `permission` - требуемый тип разрешения
- `ttl_seconds` - TTL кэша в секундах

**Ответ:**
- `has_permission` - есть ли разрешение
- `required_permission` - требуемое разрешение
- `actual_permission` - фактическое разрешение пользователя
- `CacheInfo` - информация о кэше

##### 5. GetBulkObjectPermissions
Получить разрешения пользователя на несколько объектов одновременно.

```protobuf
rpc GetBulkObjectPermissions(GetBulkObjectPermissionsRequest) returns (GetBulkObjectPermissionsResponse);
```

**Запрос:**
- `tenant_id` - идентификатор тенанта (UUID)
- `user_id` - идентификатор пользователя
- `object_api_names` - список API имен объектов
- `ttl_seconds` - TTL кэша в секундах

**Ответ:**
- `ObjectPermission[]` - список разрешений на объекты
- `total_count` - общее количество объектов
- `CacheInfo` - информация о кэше

## Типы данных

### PermissionBitmask
Битовые маски для разрешений:

- **READ** (1) - чтение данных
- **CREATE** (2) - создание записей
- **UPDATE** (4) - обновление записей
- **DELETE** (8) - удаление записей

Комбинации:
- `7` = READ + CREATE + UPDATE
- `15` = все разрешения

### GroupType
Типы групп:

- **ROLE_BASED** - группы на основе ролей
- **TERRITORY_BASED** - группы на основе территорий
- **MANUAL** - ручное управление
- **DYNAMIC** - динамическое управление

### MembershipType
Типы членства:

- **DIRECT** - прямое членство
- **INHERITED** - наследованное через роль
- **AUTOMATIC** - автоматическое через тип группы

## Примеры использования

### Go Client

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    pb "github.com/adverax/metacrm/contracts/iam-permissions/v1"
)

func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    client := pb.NewPermissionServiceClient(conn)
    
    // Проверить разрешение на чтение пользователей
    resp, err := client.CheckUserObjectPermission(context.Background(), &pb.CheckUserObjectPermissionRequest{
        TenantId: "550e8400-e29b-41d4-a716-446655440000",
        UserId: "usr_a1b2c3d4e5f67890",
        ObjectApiName: "user",
        Permission: pb.PermissionType_PERMISSION_TYPE_READ,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    if resp.HasPermission {
        log.Println("User has READ permission on user objects")
    }
}
```

### JavaScript/TypeScript Client

```typescript
import { PermissionServiceClient } from './generated/iam-permissions_grpc_pb';
import { CheckUserObjectPermissionRequest, PermissionType } from './generated/iam-permissions_pb';

const client = new PermissionServiceClient('localhost:50051');

// Проверить разрешение на чтение пользователей
const request = new CheckUserObjectPermissionRequest();
request.setTenantId('550e8400-e29b-41d4-a716-446655440000');
request.setUserId('usr_a1b2c3d4e5f67890');
request.setObjectApiName('user');
request.setPermission(PermissionType.READ);

client.checkUserObjectPermission(request, (error, response) => {
    if (error) {
        console.error('Error:', error);
        return;
    }
    
    if (response.getHasPermission()) {
        console.log('User has READ permission on user objects');
    }
});
```

## Кэширование

Сервис использует многоуровневое кэширование:

1. **L1 Cache** - кэш в памяти (TTL: 5 минут)
2. **L2 Cache** - кэш в Redis (TTL: 1 час)
3. **Database** - PostgreSQL с кэшированными функциями

### Управление кэшем

- **TTL настраивается** для каждого запроса
- **Автоматическая инвалидация** при изменении разрешений
- **Lazy loading** - кэш создается по требованию
- **Batch invalidation** - массовая инвалидация при изменениях

## Безопасность

- **Multi-tenant изоляция** - все запросы проверяются по tenant_id
- **Rate limiting** - ограничение частоты запросов
- **Audit logging** - логирование всех обращений к разрешениям
- **Permission validation** - проверка прав на доступ к данным

## Мониторинг

### Метрики

- Количество запросов по типу
- Время ответа сервиса
- Hit rate кэша
- Количество ошибок

### Логирование

- Все запросы логируются с tenant_id и user_id
- Ошибки доступа логируются с полным контекстом
- Производительность отслеживается по каждому методу

## Генерация кода

### Go

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    iam-permissions.proto
```

### JavaScript/TypeScript

```bash
protoc --js_out=import_style=commonjs,binary:. \
    --grpc-web_out=import_style=typescript,mode=grpcwebtext:. \
    iam-permissions.proto
```

### Python

```bash
python -m grpc_tools.protoc --python_out=. --grpc_python_out=. iam-permissions.proto
```

## Ограничения

- Максимум 100 объектов в bulk запросах
- TTL кэша: минимум 60 секунд, максимум 24 часа
- Rate limit: 1000 запросов в минуту на пользователя
- Таймаут запроса: 30 секунд
