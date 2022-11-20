package server

type ErrorCode int

const (
	ErrorBrowserWarning ErrorCode = 1000 + iota
	ErrorUndefinedClient
	ErrorConnectClient
	ErrorReceiveData
)

const BrowserWarningNoScriptText = "Be careful and don't input your payment or personal data on this site, this website may be used as fishing or a hacker attack. If you dont trust this site owner, please close it!"
const DefaultNoScriptText = "You need to enable JavaScript to run this app."
