# Identity Core — API Documentation

> Tài liệu dành cho Frontend / Mobile team tích hợp với **identity_core** service.

**Base URL:** `http://localhost:8080` (development)

**API version:** `v1` — tất cả endpoint nằm dưới prefix `/api/v1`

---

## Mục lục

1. [Tổng quan](#1-tổng-quan)
2. [Quy ước chung](#2-quy-ước-chung)
3. [User API](#3-user-api) — dành cho app người dùng cuối
4. [Admin API](#4-admin-api) — dành cho trang quản trị
5. [Internal API](#5-internal-api) — service-to-service (tham khảo)
6. [Data Models](#6-data-models)
7. [Enum / Constants](#7-enum--constants)
8. [Luồng tích hợp gợi ý](#8-luồng-tích-hợp-gợi-ý)

---

## 1. Tổng quan

Service chia API thành **3 nhóm** tách biệt:

| Nhóm | Prefix | Đối tượng sử dụng | Frontend cần tích hợp? |
|------|--------|-------------------|------------------------|
| **User** | `/api/v1/user` | App người dùng cuối | **Có** |
| **Admin** | `/api/v1/admin` | Trang quản trị / CMS | **Có** (admin panel) |
| **Internal** | `/api/v1/internal` | Backend service khác | Không (gọi qua BFF/gateway) |

**Multi-tenant:** Mọi request (trừ quản lý tenants ở Admin) **bắt buộc** gửi header `X-Tenant-Code` để xác định tenant. Không gửi hoặc gửi sai → lỗi, không truy cập được data tenant khác.

---

## 2. Quy ước chung

### 2.1 Response format

Mọi response đều theo cấu trúc:

```json
// Thành công
{
  "success": true,
  "data": { ... }
}

// Lỗi
{
  "success": false,
  "error": {
    "code": "TENANT_REQUIRED",
    "message": "X-Tenant-Code header is required"
  }
}
```

### 2.2 HTTP Status codes

| Status | Ý nghĩa |
|--------|---------|
| `200` | OK |
| `201` | Created |
| `204` | No Content (xóa/cập nhật thành công, không trả body) |
| `400` | Bad Request — dữ liệu không hợp lệ |
| `401` | Unauthorized — thiếu/sai auth |
| `403` | Forbidden — không có quyền / tenant inactive |
| `404` | Not Found |
| `409` | Conflict — trùng code (tenant, role, permission) |
| `500` | Internal Server Error |

### 2.3 Error codes thường gặp

| Code | HTTP | Mô tả |
|------|------|-------|
| `TENANT_REQUIRED` | 400 | Thiếu header `X-Tenant-Code` |
| `TENANT_INVALID` | 403 | Tenant không tồn tại hoặc không active |
| `UNAUTHORIZED` | 401 | Sai API key hoặc thiếu user ID |
| `NOT_FOUND` | 404 | Resource không tồn tại |
| `BAD_REQUEST` | 400 | Request không hợp lệ |
| `CONFLICT` | 409 | Dữ liệu trùng (code đã tồn tại) |
| `FORBIDDEN` | 403 | Không có quyền thực hiện |

### 2.4 Headers

| Header | Bắt buộc | Dùng ở | Mô tả |
|--------|----------|--------|-------|
| `Content-Type` | Có (POST/PUT) | Tất cả | `application/json` |
| `X-Tenant-Code` | Có* | User, Admin (scoped), Internal | Mã tenant, ví dụ `galaxy` |
| `X-Admin-API-Key` | Có | Admin API | API key quản trị |
| `X-Internal-API-Key` | Có | Internal API | API key service-to-service |
| `X-User-ID` | Legacy | User API | Chỉ khi `ALLOW_LEGACY_USER_ID_HEADER=true` |
| `Authorization` | Có | User API (sau login) / Admin | `Bearer <user-jwt>` hoặc admin JWT |

> \* Admin API **không** cần `X-Tenant-Code` khi quản lý tenants (`/admin/tenants`). Các endpoint còn lại trong Admin **cần** header này.

### 2.5 ID

Tất cả ID là **số nguyên** (`int64`), tự tăng. Ví dụ: `1`, `42`, `100`.

### 2.6 Thời gian

Format ISO 8601: `"2026-06-10T14:30:00Z"`

### 2.7 Metadata

Trường `metadata` là JSON object tùy ý, dùng lưu dữ liệu mở rộng:

```json
{ "metadata": { "source": "mobile", "referral_code": "ABC123" } }
```

---

## 3. User API

> **Dành cho:** App người dùng cuối (web, mobile, Zalo Mini App).
>
> **Auth:** Sau khi `POST /auth/zalo`, lưu `access_token` (Galaxy JWT) và gửi `Authorization: Bearer <access_token>`. Hết hạn thì login lại bằng Zalo `getAccessToken` (không dùng refresh token).

**Headers bắt buộc mọi request đã đăng nhập:**

```
X-Tenant-Code: sgs
Authorization: Bearer <user-jwt>
```

### 3.0 Đăng nhập / upsert Zalo

```
POST /api/v1/user/auth/zalo
```

Headers: `X-Tenant-Code` (không cần Bearer).

Body (production):
```json
{
  "access_token": "...",
  "phone_token": "...",
  "name": "...",
  "avatar_url": "..."
}
```
- `access_token` — từ `getAccessToken` (bắt buộc). Server gọi Graph `me`.
- `phone_token` — optional, từ `getPhoneNumber`. Server đổi SĐT qua `GET https://graph.zalo.me/v2.0/me/info` (cần `ZALO_APP_SECRET_KEY`).
- `name` / `avatar_url` — optional từ `getUserInfo` (bổ sung nếu Graph thiếu).

Body (dev, `ZALO_AUTH_DEV_MODE=true`): `{ "zalo_id", "name?", "avatar_url?", "phone?" }`.

Response:
```json
{
  "access_token": "<galaxy-jwt>",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user_id": 42,
  "user": { "...": "..." },
  "is_member": false
}
```

### 3.0a Đổi phone token → SĐT

```
POST /api/v1/user/auth/zalo/phone
```

Headers: `X-Tenant-Code`.  
Body: `{ "access_token", "phone_token" }` → `{ "phone": "09..." }`.  
Cần env `ZALO_APP_SECRET_KEY`.

### 3.0b Đăng ký thành viên (auto-active)

```
POST /api/v1/user/members/register
```

Headers: `X-Tenant-Code` + `Authorization: Bearer <user-jwt>`.
Body: `full_name`, `phone`, `email`, `avatar_url`, optional `city`, `ward`.  
Gán role `member` (tạo nếu chưa có), gọi ZNS stub log.

---

### 3.1 Lấy profile của tôi

```
GET /api/v1/user/me
```

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "id": 42,
    "tenant_id": 1,
    "full_name": "Nguyen Van A",
    "display_name": "Anh A",
    "avatar_url": "https://cdn.example.com/avatar.jpg",
    "gender": "male",
    "birthday": "1990-01-15T00:00:00Z",
    "email": "a@example.com",
    "phone": "0901234567",
    "address": "123 Duong ABC",
    "city": "Ho Chi Minh",
    "district": "Quan 1",
    "ward": "Phuong Ben Nghe",
    "username": "user_a",
    "status": "active",
    "metadata": {},
    "created_at": "2026-06-10T10:00:00Z",
    "updated_at": "2026-06-10T10:00:00Z",
    "identities": [
      {
        "id": 10,
        "user_id": 42,
        "provider": "local",
        "provider_user_id": "",
        "identity": "user_a",
        "metadata": {},
        "created_at": "2026-06-10T10:00:00Z",
        "updated_at": "2026-06-10T10:00:00Z"
      }
    ],
    "roles": [
      {
        "id": 2,
        "tenant_id": 1,
        "code": "member",
        "name": "Thành viên",
        "description": "",
        "is_system_role": false,
        "permissions": [
          {
            "id": 5,
            "code": "profile.read",
            "name": "Xem profile",
            "module": "profile",
            "description": "",
            "created_at": "2026-06-10T09:00:00Z"
          }
        ],
        "created_at": "2026-06-10T09:00:00Z",
        "updated_at": "2026-06-10T09:00:00Z"
      }
    ]
  }
}
```

---

### 3.2 Cập nhật profile của tôi

```
PUT /api/v1/user/me
```

**Request body** — chỉ gửi field cần thay đổi:

```json
{
  "display_name": "Anh A Updated",
  "avatar_url": "https://cdn.example.com/new-avatar.jpg",
  "phone": "0909999888",
  "address": "456 Duong XYZ",
  "city": "Ha Noi",
  "metadata": { "theme": "dark" }
}
```

| Field | Type | Ghi chú |
|-------|------|---------|
| `full_name` | string | |
| `display_name` | string | |
| `avatar_url` | string | |
| `gender` | string | |
| `birthday` | string (ISO 8601) | |
| `email` | string | |
| `phone` | string | |
| `address` | string | |
| `city` | string | |
| `district` | string | |
| `ward` | string | |
| `username` | string | |
| `status` | string | `active` \| `inactive` \| `banned` |
| `metadata` | object | |

**Response `200`:** Trả về object User đã cập nhật (không kèm identities/roles).

---

### 3.3 Lấy danh sách identities của tôi

```
GET /api/v1/user/me/identities
```

**Response `200`:**

```json
{
  "success": true,
  "data": [
    {
      "id": 10,
      "user_id": 42,
      "provider": "local",
      "provider_user_id": "",
      "identity": "user_a",
      "metadata": {},
      "created_at": "2026-06-10T10:00:00Z",
      "updated_at": "2026-06-10T10:00:00Z"
    },
    {
      "id": 11,
      "user_id": 42,
      "provider": "google",
      "provider_user_id": "google-uid-123",
      "identity": "a@gmail.com",
      "metadata": {},
      "created_at": "2026-06-10T11:00:00Z",
      "updated_at": "2026-06-10T11:00:00Z"
    }
  ]
}
```

> **Lưu ý:** `password_hash` không bao giờ trả về client.

---

### 3.4 Lấy danh sách quan hệ của tôi

```
GET /api/v1/user/me/relationships
```

**Response `200`:**

```json
{
  "success": true,
  "data": [
    {
      "id": 5,
      "tenant_id": 1,
      "from_user_id": 42,
      "to_user_id": 99,
      "relationship_type": "friend",
      "status": "active",
      "metadata": {},
      "created_at": "2026-06-10T12:00:00Z",
      "updated_at": "2026-06-10T12:00:00Z"
    }
  ]
}
```

---

## 4. Admin API

> **Dành cho:** Trang quản trị (admin panel / CMS).
>
> **Auth:** Header `X-Admin-API-Key` trên mọi request.

**Headers:**

```
X-Admin-API-Key: <admin_api_key>
X-Tenant-Code: galaxy        ← bắt buộc cho endpoint scoped theo tenant
Content-Type: application/json
```

---

### 4.1 Tenants

> Không cần `X-Tenant-Code`.

#### Tạo tenant

```
POST /api/v1/admin/tenants
```

```json
{
  "code": "galaxy",
  "name": "Galaxy Platform",
  "status": "active",
  "metadata": {}
}
```

| Field | Type | Bắt buộc | Ghi chú |
|-------|------|----------|---------|
| `code` | string | Có | 2–64 ký tự, unique |
| `name` | string | Có | |
| `status` | string | Không | Mặc định `active`. `active` \| `inactive` \| `suspended` |
| `metadata` | object | Không | |

**Response `201`:** Object Tenant.

---

#### Danh sách tenants

```
GET /api/v1/admin/tenants?page=1&page_size=20
```

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "items": [ { "id": 1, "code": "galaxy", "name": "Galaxy Platform", "status": "active", "metadata": {}, "created_at": "...", "updated_at": "..." } ],
    "meta": {
      "page": 1,
      "page_size": 20,
      "total": 1,
      "total_pages": 1
    }
  }
}
```

---

#### Chi tiết tenant

```
GET /api/v1/admin/tenants/:id
```

---

#### Cập nhật tenant

```
PUT /api/v1/admin/tenants/:id
```

```json
{
  "name": "Galaxy Platform V2",
  "status": "active",
  "metadata": { "plan": "enterprise" }
}
```

---

### 4.2 Users

> Cần `X-Tenant-Code`.

#### Tạo user

```
POST /api/v1/admin/users
```

```json
{
  "full_name": "Nguyen Van A",
  "display_name": "Anh A",
  "email": "a@example.com",
  "phone": "0901234567",
  "username": "user_a",
  "password": "secret123",
  "status": "active",
  "gender": "male",
  "birthday": "1990-01-15T00:00:00Z",
  "address": "123 Duong ABC",
  "city": "Ho Chi Minh",
  "district": "Quan 1",
  "ward": "Phuong Ben Nghe",
  "avatar_url": "",
  "metadata": {},
  "role_ids": [2, 3]
}
```

| Field | Type | Bắt buộc | Ghi chú |
|-------|------|----------|---------|
| `full_name` | string | Không | |
| `display_name` | string | Không | |
| `email` | string | Không | |
| `phone` | string | Không | |
| `username` | string | Không | Dùng làm identity nếu có `password` |
| `password` | string | Không | Tạo identity `local` kèm password |
| `status` | string | Không | Mặc định `active` |
| `role_ids` | int[] | Không | Gán roles ngay khi tạo |
| `metadata` | object | Không | |

**Response `201`:** Object User (kèm `identities`, `roles`).

---

#### Danh sách users

```
GET /api/v1/admin/users?page=1&page_size=20&status=active&search=nguyen
```

| Query | Type | Mô tả |
|-------|------|-------|
| `page` | int | Mặc định `1` |
| `page_size` | int | Mặc định `20`, tối đa `100` |
| `status` | string | Lọc: `active` \| `inactive` \| `banned` |
| `search` | string | Tìm theo full_name, display_name, email, phone, username |

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "items": [ { "id": 42, "tenant_id": 1, "full_name": "...", "status": "active", ... } ],
    "meta": { "page": 1, "page_size": 20, "total": 50, "total_pages": 3 }
  }
}
```

---

#### Chi tiết user

```
GET /api/v1/admin/users/:id
```

**Response `200`:** User kèm `identities` và `roles` (kèm `permissions`).

---

#### Cập nhật user

```
PUT /api/v1/admin/users/:id
```

```json
{
  "full_name": "Nguyen Van B",
  "status": "inactive",
  "metadata": { "note": "tạm khóa" }
}
```

---

#### Xóa user

```
DELETE /api/v1/admin/users/:id
```

**Response `204`:** Không có body.

---

#### Gán roles cho user

```
PUT /api/v1/admin/users/:id/roles
```

```json
{
  "role_ids": [2, 3, 5]
}
```

> Ghi đè toàn bộ roles hiện tại. Gửi `[]` để xóa hết roles.

**Response `204`:** Không có body.

---

### 4.3 Roles

> Cần `X-Tenant-Code`. Roles thuộc theo tenant.

#### Tạo role

```
POST /api/v1/admin/roles
```

```json
{
  "code": "editor",
  "name": "Biên tập viên",
  "description": "Quyền chỉnh sửa nội dung",
  "is_system_role": false,
  "permission_ids": [1, 2, 5]
}
```

**Response `201`:** Object Role (kèm `permissions`).

---

#### Danh sách roles

```
GET /api/v1/admin/roles
```

**Response `200`:** Array Role (kèm `permissions`).

---

#### Chi tiết role

```
GET /api/v1/admin/roles/:id
```

---

#### Cập nhật role

```
PUT /api/v1/admin/roles/:id
```

```json
{
  "name": "Biên tập viên cấp cao",
  "description": "Mô tả mới",
  "permission_ids": [1, 2, 3, 4]
}
```

> `permission_ids` ghi đè toàn bộ permissions của role.

---

#### Xóa role

```
DELETE /api/v1/admin/roles/:id
```

> Không xóa được `is_system_role = true`.

**Response `204`:** Không có body.

---

### 4.4 Permissions

Tracking permissions seeded by `scripts/migrations/002_seed_tracking_permissions.sql`:

- `tracking.view`: view campaigns, links, QR and analytics.
- `tracking.manage`: create, update and delete campaigns and links.

> Global (không theo tenant). Không cần `X-Tenant-Code`.

#### Tạo permission

```
POST /api/v1/admin/permissions
```

```json
{
  "code": "user.create",
  "name": "Tạo người dùng",
  "module": "user",
  "description": "Cho phép tạo user mới"
}
```

| Field | Type | Bắt buộc |
|-------|------|----------|
| `code` | string | Có, unique |
| `name` | string | Có |
| `module` | string | Có — nhóm permission |
| `description` | string | Không |

**Response `201`:** Object Permission.

---

#### Danh sách permissions

```
GET /api/v1/admin/permissions
```

**Response `200`:** Array Permission, sắp xếp theo `module`, `code`.

---

### 4.5 User Relationships

> Cần `X-Tenant-Code`.

#### Tạo quan hệ

```
POST /api/v1/admin/relationships
```

```json
{
  "from_user_id": 42,
  "to_user_id": 99,
  "relationship_type": "friend",
  "status": "active",
  "metadata": {}
}
```

| Field | Type | Bắt buộc | Ghi chú |
|-------|------|----------|---------|
| `from_user_id` | int | Có | > 0 |
| `to_user_id` | int | Có | > 0, khác `from_user_id` |
| `relationship_type` | string | Có | Ví dụ: `friend`, `family`, `colleague` |
| `status` | string | Không | Mặc định `active` |
| `metadata` | object | Không | |

**Response `201`:** Object UserRelationship.

---

#### Danh sách quan hệ của user

```
GET /api/v1/admin/users/:id/relationships
```

**Response `200`:** Array UserRelationship (cả from và to).

---

#### Xóa quan hệ

```
DELETE /api/v1/admin/relationships/:id
```

**Response `204`:** Không có body.

---

## 5. Internal API

> **Không dùng trực tiếp từ Frontend.** Gọi qua BFF/gateway hoặc backend service khác.

**Headers:**

```
X-Internal-API-Key: <internal_api_key>
X-Tenant-Code: galaxy
```

| Method | Endpoint | Mô tả |
|--------|----------|-------|
| `GET` | `/api/v1/internal/users/:id` | Lấy user (kèm identities, roles) |
| `GET` | `/api/v1/internal/users` | List users (cùng query như Admin) |
| `POST` | `/api/v1/internal/auth/verify` | Xác thực đăng nhập |
| `GET` | `/api/v1/internal/users/:id/relationships` | Lấy quan hệ của user |

### Verify identity (đăng nhập)

```
POST /api/v1/internal/auth/verify
```

```json
{
  "identity": "user_a",
  "password": "secret123",
  "provider": "local"
}
```

| Field | Type | Bắt buộc | Ghi chú |
|-------|------|----------|---------|
| `identity` | string | Có | username / email / phone |
| `password` | string | Có | |
| `provider` | string | Không | Mặc định `local` |

**Response `200`:** Object User (kèm identities, roles) nếu đăng nhập thành công.

**Response `401`:** Sai identity/password hoặc user không active.

---

## 6. Data Models

### Tenant

```typescript
interface Tenant {
  id: number;
  code: string;
  name: string;
  status: "active" | "inactive" | "suspended";
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}
```

### User

```typescript
interface User {
  id: number;
  tenant_id: number;
  full_name: string;
  display_name: string;
  avatar_url: string;
  gender: string;
  birthday: string | null;
  email: string;
  phone: string;
  address: string;
  city: string;
  district: string;
  ward: string;
  username: string;
  status: "active" | "inactive" | "banned";
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  identities?: UserIdentity[];  // chỉ có khi preload
  roles?: Role[];                 // chỉ có khi preload
}
```

### UserIdentity

```typescript
interface UserIdentity {
  id: number;
  user_id: number;
  provider: string;          // "local" | "google" | "zalo"
  provider_user_id: string;
  identity: string;          // username / email / phone / zalo_id
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}
```

### Role

```typescript
interface Role {
  id: number;
  tenant_id: number;
  code: string;
  name: string;
  description: string;
  is_system_role: boolean;
  permissions?: Permission[];
  created_at: string;
  updated_at: string;
}
```

### Permission

```typescript
interface Permission {
  id: number;
  code: string;
  name: string;
  module: string;
  description: string;
  created_at: string;
}
```

### UserRelationship

```typescript
interface UserRelationship {
  id: number;
  tenant_id: number;
  from_user_id: number;
  to_user_id: number;
  relationship_type: string;
  status: "active" | "inactive" | "pending";
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}
```

### Pagination

```typescript
interface PaginatedResponse<T> {
  items: T[];
  meta: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}
```

---

## 7. Enum / Constants

### Tenant status
`active` · `inactive` · `suspended`

### User status
`active` · `inactive` · `banned`

### Relationship status
`active` · `inactive` · `pending`

### Identity provider
`local` · `google` · `zalo`

### Relationship type (gợi ý, tự định nghĩa thêm)
`friend` · `family` · `colleague` · `manager` · `subordinate`

---

## 8. Luồng tích hợp gợi ý

### 8.1 Admin Panel

```
1. Cấu hình X-Admin-API-Key trong env frontend (hoặc qua BFF)
2. Tạo tenant (nếu chưa có)
3. Tạo permissions (global, 1 lần)
4. Tạo roles + gán permissions (theo tenant)
5. Tạo users + gán roles
6. Mọi request scoped: gửi X-Tenant-Code = tenant đang quản lý
```

### 8.2 User App (Mobile / Web)

```
1. App dùng tenant cố định hoặc chọn tenant trước khi đăng nhập
2. Gọi POST /api/v1/user/auth/zalo bằng Zalo access token
3. Backend trả Galaxy access_token (JWT)
4. Frontend lưu access_token, gửi kèm mọi request:
   - X-Tenant-Code: <tenant_code>
   - Authorization: Bearer <access_token>
5. GET /user/me để hiển thị profile
6. PUT /user/me để cập nhật profile
```

### 8.3 Health check

```
GET /health
```

```json
{ "status": "ok" }
```

---

## Phụ lục: Ví dụ curl

### Admin — tạo tenant

```bash
curl -X POST http://localhost:8080/api/v1/admin/tenants \
  -H "X-Admin-API-Key: change-me-admin-key" \
  -H "Content-Type: application/json" \
  -d '{"code":"galaxy","name":"Galaxy Platform"}'
```

### Admin — tạo user

```bash
curl -X POST http://localhost:8080/api/v1/admin/users \
  -H "X-Admin-API-Key: change-me-admin-key" \
  -H "X-Tenant-Code: galaxy" \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Nguyen Van A","username":"user_a","password":"secret123","email":"a@example.com"}'
```

### User — lấy profile

```bash
curl http://localhost:8080/api/v1/user/me \
  -H "X-Tenant-Code: galaxy" \
  -H "Authorization: Bearer <user-jwt>"
```

### Internal — verify login (qua BFF)

```bash
curl -X POST http://localhost:8080/api/v1/internal/auth/verify \
  -H "X-Internal-API-Key: change-me-internal-key" \
  -H "X-Tenant-Code: galaxy" \
  -H "Content-Type: application/json" \
  -d '{"identity":"user_a","password":"secret123"}'
```
