package internal

import "net/http"

const (
	StatusContinue           Status = http.StatusContinue
	StatusSwitchingProtocols Status = http.StatusSwitchingProtocols
	StatusProcessing         Status = http.StatusProcessing
	//StatusEarlyHints         Status = http.StatusEarlyHints

	StatusOK                   Status = http.StatusOK
	StatusCreated              Status = http.StatusCreated
	StatusAccepted             Status = http.StatusAccepted
	StatusNonAuthoritativeInfo Status = http.StatusNonAuthoritativeInfo
	StatusNoContent            Status = http.StatusNoContent
	StatusResetContent         Status = http.StatusResetContent
	StatusPartialContent       Status = http.StatusPartialContent
	StatusMultiStatus          Status = http.StatusMultiStatus
	StatusAlreadyReported      Status = http.StatusAlreadyReported
	StatusIMUsed               Status = http.StatusIMUsed

	StatusMultipleChoices   Status = http.StatusMultipleChoices
	StatusMovedPermanently  Status = http.StatusMovedPermanently
	StatusFound             Status = http.StatusFound
	StatusSeeOther          Status = http.StatusSeeOther
	StatusNotModified       Status = http.StatusNotModified
	StatusUseProxy          Status = http.StatusUseProxy
	StatusTemporaryRedirect Status = http.StatusTemporaryRedirect
	StatusPermanentRedirect Status = http.StatusPermanentRedirect

	StatusBadRequest                   Status = http.StatusBadRequest
	StatusUnauthorized                 Status = http.StatusUnauthorized
	StatusPaymentRequired              Status = http.StatusPaymentRequired
	StatusForbidden                    Status = http.StatusForbidden
	StatusNotFound                     Status = http.StatusNotFound
	StatusMethodNotAllowed             Status = http.StatusMethodNotAllowed
	StatusNotAcceptable                Status = http.StatusNotAcceptable
	StatusProxyAuthRequired            Status = http.StatusProxyAuthRequired
	StatusRequestTimeout               Status = http.StatusRequestTimeout
	StatusConflict                     Status = http.StatusConflict
	StatusGone                         Status = http.StatusGone
	StatusLengthRequired               Status = http.StatusLengthRequired
	StatusPreconditionFailed           Status = http.StatusPreconditionFailed
	StatusRequestEntityTooLarge        Status = http.StatusRequestEntityTooLarge
	StatusRequestURITooLong            Status = http.StatusRequestURITooLong
	StatusUnsupportedMediaType         Status = http.StatusUnsupportedMediaType
	StatusRequestedRangeNotSatisfiable Status = http.StatusRequestedRangeNotSatisfiable
	StatusExpectationFailed            Status = http.StatusExpectationFailed
	StatusTeapot                       Status = http.StatusTeapot
	StatusMisdirectedRequest           Status = http.StatusMisdirectedRequest
	StatusUnprocessableEntity          Status = http.StatusUnprocessableEntity
	StatusLocked                       Status = http.StatusLocked
	StatusFailedDependency             Status = http.StatusFailedDependency
	//StatusTooEarly                     Status = http.StatusTooEarly
	StatusUpgradeRequired              Status = http.StatusUpgradeRequired
	StatusPreconditionRequired         Status = http.StatusPreconditionRequired
	StatusTooManyRequests              Status = http.StatusTooManyRequests
	StatusRequestHeaderFieldsTooLarge  Status = http.StatusRequestHeaderFieldsTooLarge
	StatusUnavailableForLegalReasons   Status = http.StatusUnavailableForLegalReasons

	StatusInternalServerError           Status = http.StatusInternalServerError
	StatusNotImplemented                Status = http.StatusNotImplemented
	StatusBadGateway                    Status = http.StatusBadGateway
	StatusServiceUnavailable            Status = http.StatusServiceUnavailable
	StatusGatewayTimeout                Status = http.StatusGatewayTimeout
	StatusHTTPVersionNotSupported       Status = http.StatusHTTPVersionNotSupported
	StatusVariantAlsoNegotiates         Status = http.StatusVariantAlsoNegotiates
	StatusInsufficientStorage           Status = http.StatusInsufficientStorage
	StatusLoopDetected                  Status = http.StatusLoopDetected
	StatusNotExtended                   Status = http.StatusNotExtended
	StatusNetworkAuthenticationRequired Status = http.StatusNetworkAuthenticationRequired
)

type Status int
