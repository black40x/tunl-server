package server

type ErrorCode int

const (
	ErrBrowserWarning ErrorCode = 1000 + iota
	ErrUndefinedClient
	ErrConnectClient
	ErrReceiveData
)

const (
	ErrClientResponse ErrorCode = 2000 + iota
)

const BrowserWarningNoScriptText = "Be careful and don't input your payment or personal data on this site, this website may be used as fishing or a hacker attack. If you dont trust this site owner, please close it!"
const DefaultNoScriptText = "You need to enable JavaScript to run this app."
