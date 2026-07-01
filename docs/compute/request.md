---
subcategory: "Compute Engine"
page_title: "Request"
description: |-
  Tools for listing and inspecting API requests in IONOS CLOUD.
---

# Requests

## list_requests

Lists all API requests in your IONOS CLOUD account.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
| `filters` | object | No | Server-side property filters; e.g. `{"method":"POST","requestStatus":"DONE"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_requests",
  "arguments": {}
}
```

**API Reference:** [requestsGet](https://api.ionos.com/docs/cloud/v6/#tag/Requests/operation/requestsGet)

---

## get_request

Gets detailed information about a specific API request.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `request_id` | string | Yes | The ID of the request |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_request",
  "arguments": {
    "request_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [requestsFindById](https://api.ionos.com/docs/cloud/v6/#tag/Requests/operation/requestsFindById)

---

## get_request_status

Gets the status of a specific API request.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `request_id` | string | Yes | The ID of the request |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_request_status",
  "arguments": {
    "request_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [requestsStatusGet](https://api.ionos.com/docs/cloud/v6/#tag/Requests/operation/requestsStatusGet)
