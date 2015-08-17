package martianlua

import (
	"net/http"
	"net/url"
	"sync"

	lua "github.com/Shopify/go-lua"
	luautil "github.com/Shopify/goluago/util"
	"github.com/google/martian"
)

type Modifier struct {
	mu     sync.RWMutex
	script string
}

func NewModifier() *Modifier {
	return &Modifier{}
}

func (m *Modifier) SetScript(script string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.script = script
}

func (m *Modifier) ModifyRequest(req *http.Request) error {
	l := lua.NewState()
	lua.OpenLibraries(l)

	m.mu.RLock()
	if err := lua.DoString(l, m.script); err != nil {
		martian.Errorf("martianlua: failed to load script: %v", err)
		return err
	}
	m.mu.RUnlock()

	l.Field(-1, "modifyrequest")
	pushRequest(l, req)

	return l.ProtectedCall(1, 0, 0)
}

var reqFuncs = map[string]func(*http.Request) lua.Function{
	"forcehttps": forceHTTPS,
}

func forceHTTPS(req *http.Request) lua.Function {
	martian.Infof("creating forcehttps function")
	return func(l *lua.State) int {
		martian.Infof("called forcehttps")
		req.URL.Scheme = "https"
		return 0
	}
}

func pushRequest(l *lua.State, req *http.Request) {
	l.NewTable()

	for name, fn := range reqFuncs {
		l.PushGoFunction(fn(req))
		l.SetField(-2, name)
	}

	l.NewTable()

	getHook := func(l *lua.State) int {
		key := lua.CheckString(l, 2)
		switch key {
		case "method":
			l.PushString(req.Method)
		case "url":
			l.PushString(req.URL.String())
		case "protomajor":
			l.PushInteger(req.ProtoMajor)
		case "protominor":
			l.PushInteger(req.ProtoMinor)
		case "contentlength":
			l.PushNumber(float64(req.ContentLength))
		case "transferencodings":
			return luautil.DeepPush(l, req.TransferEncoding)
		case "host":
			l.PushString(req.Host)
		case "remoteaddr":
			l.PushString(req.RemoteAddr)
		case "requesturi":
			l.PushString(req.RequestURI)
		default:
			return 0
		}
		return 1
	}

	l.PushGoFunction(getHook)
	l.SetField(-2, "__index")

	setHook := func(l *lua.State) int {
		martian.Infof("setter called")

		key := lua.CheckString(l, 2)
		switch key {
		case "method":
			val := lua.CheckString(l, 3)
			req.Method = val
		case "url":
			martian.Infof("martianlua: got URL setter")
			val := lua.CheckString(l, 3)
			u, err := url.Parse(val)
			if err != nil {
				lua.Errorf(l, err.Error())
				panic("unreachable")
			}

			req.URL = u
		case "protomajor":
			val := lua.CheckInteger(l, 3)
			req.ProtoMajor = val
		case "protominor":
			val := lua.CheckInteger(l, 3)
			req.ProtoMinor = val
		case "contentlength":
			val := lua.CheckInteger(l, 3)
			req.ProtoMinor = val
			l.PushNumber(float64(req.ContentLength))
		case "transferencodings":
			// How to deal with []string?
		case "host":
			l.PushString(req.Host)
		case "remoteaddr":
			l.PushString(req.RemoteAddr)
		case "requesturi":
			l.PushString(req.RequestURI)
		default:
			return 0
		}
		return 1
	}

	l.PushGoFunction(setHook)
	l.SetField(-2, "__newindex")

	l.SetMetaTable(-2)
}
