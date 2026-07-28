# Backend Required Work

> Backend implementation needed to support the frontend permissions system and
> remaining feature gaps. For general frontend-requested routes and field
> additions see [`FRONTEND-REQUESTS.md`](./FRONTEND-REQUESTS.md). For defects
> in existing routes see [`KNOWN-ISSUES.md`](./KNOWN-ISSUES.md).

---

## Priority 1: Permissions System ✅ Done

| Item | Status | Location |
|---|---|---|
| Default role seeding on company creation | ✅ | `internal/http/handler/role_seed.go` |
| `GET /api/v1/me/permissions` endpoint | ✅ | `internal/http/handler/me.go` |
| Navigation item registry (TypeScript) | ✅ | Frontend repo — `navigation.ts` |
| Model ID registry (TypeScript) | ✅ | Frontend repo — `model-ids.ts` |

---

## Priority 2: M2M Endpoint Scaffolding

### 2.1 Issue Assignees

Expose CRUD for `issue_assigned_to` join table.

```
GET    /api/v1/issues/:id/assigned-to
POST   /api/v1/issues/:id/assigned-to
DELETE /api/v1/issues/:id/assigned-to/:employee_id
```

### 2.2 Issue Watchers

Expose CRUD for `issue_watchers` join table.

```
GET    /api/v1/issues/:id/watchers
POST   /api/v1/issues/:id/watchers
DELETE /api/v1/issues/:id/watchers/:employee_id
```

### 2.3 Work Order ↔ Issue

Expose CRUD for `work_order_issues` join table.

```
GET    /api/v1/work-orders/:id/issues
POST   /api/v1/work-orders/:id/issues
DELETE /api/v1/work-orders/:id/issues/:issue_id
```

### 2.4 Work Order ↔ Fault

Expose CRUD for `work_order_faults` join table.

```
GET    /api/v1/work-orders/:id/faults
POST   /api/v1/work-orders/:id/faults
DELETE /api/v1/work-orders/:id/faults/:fault_id
```

### 2.5 Service Entry Line Item ↔ Issue

Expose CRUD for `service_entry_line_item_issues` join table.

```
GET    /api/v1/service-entry-line-items/:id/issues
POST   /api/v1/service-entry-line-items/:id/issues
DELETE /api/v1/service-entry-line-items/:id/issues/:issue_id
```

---

## Priority 3: Employee Endpoints

> Overlaps with [`FRONTEND-REQUESTS.md`](./FRONTEND-REQUESTS.md) §2.1 (list
> employee companies) and §2.2 (switch company). Items below are the
> permissions-system subset; the full employee-company story lives there.

### 3.1 `GET /api/v1/me/permissions` response — employee companies

The current `/me/permissions` response does not include the list of companies
an employee belongs to. Once the employee-companies M2M endpoint exists
(FRONTEND-REQUESTS §2.1), consider adding a `companies` field to the
permissions response so the company switcher can resolve in one request.

---

## Priority 4: Group CRUD Registration

The `Group` domain struct exists but has no handler or router registration.

**Router addition:**

```go
crud.NewHandler[dto.GroupResponse, dto.CreateGroupRequest, dto.UpdateGroupRequest](
    NewGroupStore(d.Queries)).Register(api, "/groups")
```

**Endpoints:**

```
GET/POST      /api/v1/groups
GET/PUT/DELETE /api/v1/groups/:id
```

---

## Priority 5: File Upload Improvements

> These overlap with [`KNOWN-ISSUES.md`](./KNOWN-ISSUES.md) #2 (URL column
> width), #4 (leaked objects), and #5 (no type/size validation). The items
> below are the permissions-system-adjacent concerns only.

### 5.1 Thumbnail Generation

Generate thumbnails for uploaded images. The media gallery needs thumbnails
for the grid view without loading full-size images.

---

## Priority 6: Notification System (Future)

### 6.1 `GET /api/v1/notifications`

Basic notification endpoint for the bell icon. Can be stubbed initially:

```json
{
  "data": [],
  "unread_count": 0
}
```

Real-time notifications (WebSocket/SSE) are out of scope for v1.

---

## Priority 7: Dashboard Aggregations

### 7.1 `GET /api/v1/dashboard/stats`

Aggregated KPIs for the dashboard:

```json
{
  "total_assets": 45,
  "active_work_orders": 12,
  "overdue_issues": 3,
  "upcoming_reminders": 8,
  "low_stock_parts": 5,
  "pending_inspections": 2
}
```

Avoids the frontend making 6+ separate requests on dashboard load.

---

## Implementation Order

| Priority | Item | Effort | Status |
|---|---|---|---|
| **P1** | Role seeding + /me/permissions + registries | Small | ✅ Done |
| **P2** | Issue assignees/watchers M2M | Medium | — |
| **P2** | WO ↔ Issue/Fault M2M | Medium | — |
| **P2** | SE Line Item ↔ Issue M2M | Medium | — |
| **P3** | Employee companies in /me response | Small | — |
| **P4** | Group CRUD registration | Small | — |
| **P5** | Thumbnail generation | Medium | — |
| **P6** | Notifications (stub) | Small | — |
| **P7** | Dashboard aggregations | Medium | — |

See also:
- [`FRONTEND-REQUESTS.md`](./FRONTEND-REQUESTS.md) — routes and fields the frontend needs that don't exist yet
- [`KNOWN-ISSUES.md`](./KNOWN-ISSUES.md) — defects in routes that already exist
