---
subcategory: "Activity Log"
page_title: "Events"
description: |-
  Tool for querying the IONOS CLOUD activity log.
---

# Events

## list_activitylog_events

Queries the IONOS CLOUD activity log: the full audit trail of API requests made against a contract (who did what, when, on which resource).

Requires the `ACCESS_ACTIVITY_LOG` privilege on the token used for authentication.

**Important:** Always pass `date_start` and `date_end` to narrow results — logs span years with thousands of events per day. Pass a small `limit` (e.g. 25) unless the user explicitly asks for bulk data.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number. Use `list_activitylog_contracts` to discover contract numbers, or read it from `get_billing_profile`. |
| `date_start` | string | No | Inclusive start date in `YYYY-MM-DD` format (e.g. `2026-05-01`). Defaults to 7 days ago. |
| `date_end` | string | No | Inclusive end date in `YYYY-MM-DD` format (e.g. `2026-05-11`). Defaults to today. Maximum range is 90 days. |
| `offset` | integer | No | 0-based pagination offset |
| `limit` | integer | No | Maximum number of events per page. Defaults to 25. |
| `user` | string | No | Filter events by username (client-side). E.g. `ionosctl-v6@cloud.ionos.com`. Drastically reduces output when investigating a specific account. |
| `event_types` | string[] | No | Filter to specific event types only (client-side). E.g. `["Error","RequestAccepted"]`. Omitting `Provision` and `RequestStatusUpdate` cuts ~65% of typical log volume. |
| `include_status_updates` | boolean | No | Include `RequestStatusUpdate` events (default `false`). These async provisioning echoes account for ~55% of log volume and are rarely useful. |

**Example:**

```json
{
  "name": "list_activitylog_events",
  "arguments": {
    "contract": 31909628,
    "date_start": "2026-05-01",
    "date_end": "2026-05-11",
    "limit": 25
  }
}
```

**Response shape (compacted):**

The raw API response wraps every event in an Elasticsearch `_source` object and repeats several guaranteed-redundant fields on every event. This tool returns a compacted shape to reduce output size for LLM consumers.

Fields stripped unconditionally:
- `_source` wrapper (Elasticsearch artifact)
- `meta.auditVersion` (always `0.1` or `1`, no semantic value)
- `principal.identity.contractNumber` when it equals the input `contract` (always true for `GetByContract`)
- `event.param.initiator` and `event.param.sourceService` when they equal `principal.sourceService`
- Empty `resources[].action` arrays

| Field | Type | Description |
|-------|------|-------------|
| `events` | array | List of compacted events |
| `total` | integer | Total events matching the query (may exceed the page size) |

**Per-event fields:**

| Field | Type | Description |
|-------|------|-------------|
| `time` | string | UTC timestamp of the event |
| `type` | string | Event type (e.g. `RequestAccepted`, `RequestStatusUpdate`, `GenerateToken`) |
| `action` | string | HTTP method or protocol action (e.g. `GET`, `POST`, `http`) |
| `status` | string | Request status for status-update events (e.g. `QUEUED`, `DONE`) |
| `message` | string | Human-readable status message when present |
| `uri` | string | Request URI when present |
| `request_id` | string | UUID grouping related events for the same request |
| `queue_ref_id` | integer | Provisioning queue reference (groups async provisioning events) |
| `user` | string | Username of the principal who triggered the event |
| `service` | string | IONOS service that processed the event (e.g. `PUBLIC_REST_V6`, `DCD`) |
| `source_ip` | string | Source IP of the request when present |
| `initiator` | string | Initiating service — only present when different from `service` |
| `param_service` | string | Param-level service — only present when different from `service` |
| `error_code` | string | Internal error code when present |
| `error` | object | Error detail (`http_status`, `messages`) when present |
| `resources` | array | Resources affected: `{type, id, action[]}` |

**API Reference:** [GetByContract](https://api.ionos.com/docs/activitylog/v1/#tag/Contracts/operation/GetByContract)
