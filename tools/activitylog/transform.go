package activitylog

import (
	"strconv"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/activitylog/v2"
)

// CompactResponse is the shape returned by list_activitylog_events.
// It strips the Elasticsearch-style _source wrapper and other
// guaranteed-redundant fields from the SDK response to reduce
// output tokens for LLM consumers.
type CompactResponse struct {
	Events []CompactEvent `json:"events"`
	Total  int32          `json:"total"`
}

// CompactEvent is the flattened, deduplicated view of a single audit event.
// Field naming is snake_case to match the rest of the MCP tool responses.
type CompactEvent struct {
	Time      string            `json:"time,omitempty"`
	Type      string            `json:"type,omitempty"`
	Action    string            `json:"action,omitempty"`
	Status    string            `json:"status,omitempty"`
	Message   string            `json:"message,omitempty"`
	URI       string            `json:"uri,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	QueueRef  int32             `json:"queue_ref_id,omitempty"`
	User      string            `json:"user,omitempty"`
	Service   string            `json:"service,omitempty"`
	SourceIP  string            `json:"source_ip,omitempty"`
	Initiator string            `json:"initiator,omitempty"`
	ParamSvc  string            `json:"param_service,omitempty"`
	ErrorCode string            `json:"error_code,omitempty"`
	Error     *CompactError     `json:"error,omitempty"`
	Resources []CompactResource `json:"resources,omitempty"`
}

type CompactResource struct {
	Type   string   `json:"type,omitempty"`
	ID     string   `json:"id,omitempty"`
	Action []string `json:"action,omitempty"`
}

type CompactError struct {
	HTTPStatus string   `json:"http_status,omitempty"`
	Messages   []string `json:"messages,omitempty"`
}

// Compact projects a raw GetByContractResponse into CompactResponse.
//
// inputContract is the contract number that was passed to GetByContract.
// Any principal.identity.contractNumber matching this value is dropped
// from the per-event output (it is guaranteed-redundant for that endpoint).
func Compact(raw sdk.GetByContractResponse, inputContract int32) CompactResponse {
	out := CompactResponse{Events: []CompactEvent{}}
	if raw.Hits == nil {
		return out
	}
	if raw.Hits.Total != nil {
		out.Total = *raw.Hits.Total
	}
	for _, hit := range raw.Hits.Hits {
		if hit.Source == nil {
			continue
		}
		out.Events = append(out.Events, compactSource(*hit.Source, inputContract))
	}
	return out
}

func compactSource(src sdk.GetByContractResponseHitsHitsSource, inputContract int32) CompactEvent {
	ev := CompactEvent{}

	if src.Meta != nil {
		ev.Time = deref(src.Meta.Time)
		ev.RequestID = deref(src.Meta.RequestId)
		if src.Meta.QueueRefId != nil {
			ev.QueueRef = *src.Meta.QueueRefId
		}
		// auditVersion intentionally dropped — always 0.1 or 1, no semantic value.
		// transactionId dropped: rarely present and not exposed by the v6 console either.
	}

	principalSvc := ""
	if src.Principal != nil {
		ev.SourceIP = deref(src.Principal.SourceIP)
		principalSvc = deref(src.Principal.SourceService)
		ev.Service = principalSvc
		if src.Principal.Identity != nil {
			// Drop contractNumber when it equals the input contract (guaranteed-redundant per endpoint).
			if cn := src.Principal.Identity.ContractNumber; cn != nil && *cn != inputContract {
				// Different contract somehow surfaced — keep it as part of the user field for visibility.
				// This is unexpected but worth surfacing rather than silently dropping.
				ev.User = formatForeignUser(*cn, deref(src.Principal.Identity.Username))
			} else {
				ev.User = deref(src.Principal.Identity.Username)
			}
		}
		// serviceHost dropped: internal host FQDN, not useful for audit consumers.
	}

	if src.Event != nil {
		ev.Type = deref(src.Event.Type)
		ev.Message = deref(src.Event.Message)
		ev.Status = deref(src.Event.Status)
		if src.Event.Param != nil {
			ev.Action = deref(src.Event.Param.Action)
			ev.URI = deref(src.Event.Param.Uri)
			ev.ErrorCode = deref(src.Event.Param.ErrorCode)

			// Drop initiator/sourceService when they equal the principal sourceService.
			if init := deref(src.Event.Param.Initiator); init != "" && init != principalSvc {
				ev.Initiator = init
			}
			if psvc := deref(src.Event.Param.SourceService); psvc != "" && psvc != principalSvc {
				ev.ParamSvc = psvc
			}

			if src.Event.Param.Error != nil {
				ev.Error = compactError(src.Event.Param.Error)
			}
		}
		for _, r := range src.Event.Resources {
			cr := CompactResource{
				Type: deref(r.Type),
				ID:   deref(r.Id),
			}
			// Only include action when non-empty.
			if len(r.Action) > 0 {
				cr.Action = r.Action
			}
			ev.Resources = append(ev.Resources, cr)
		}
	}

	return ev
}

func compactError(e *sdk.GetByContractResponseHitsHitsSourceEventParamError) *CompactError {
	if e == nil {
		return nil
	}
	out := &CompactError{HTTPStatus: deref(e.HttpStatus)}
	for _, m := range e.Messages {
		// Messages have nested Message/ErrorCode but the user-facing string is what matters.
		if msg := messageString(m); msg != "" {
			out.Messages = append(out.Messages, msg)
		}
	}
	if out.HTTPStatus == "" && len(out.Messages) == 0 {
		return nil
	}
	return out
}

func messageString(m sdk.GetByContractResponseHitsHitsSourceEventParamErrorMessages) string {
	msg := m.GetMessage()
	code := m.GetErrorCode()
	if msg != "" && code != "" {
		return code + ": " + msg
	}
	if msg != "" {
		return msg
	}
	return code
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatForeignUser(contractNumber int32, username string) string {
	n := strconv.Itoa(int(contractNumber))
	if username == "" {
		return "contract:" + n
	}
	return username + " (contract " + n + ")"
}
