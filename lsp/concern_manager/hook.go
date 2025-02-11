package concern_manager

import "github.com/Sora233/DDBOT/concern"

type Hook interface {
	NewsFilterHook(notify concern.Notify) *HookResult
	AtBeforeHook(notify concern.Notify) *HookResult
	ShouldSendHook(notify concern.Notify) *HookResult
}

// HookResult Pass是false的时候，要把具体失败的地方填入Reason
type HookResult struct {
	Pass   bool
	Reason string
}

func (h *HookResult) PassOrReason(pass bool, reason string) {
	if pass {
		h.Pass = pass
	} else {
		h.Reason = reason
	}
}

type defaultHook struct {
}

func (d defaultHook) NewsFilterHook(notify concern.Notify) *HookResult {
	if notify == nil {
		return defaultHookResult
	}
	switch notify.Type() {
	case concern.BilibiliNews:
		return defaultHookResult
	default:
		return HookResultPass
	}
}

func (d defaultHook) AtBeforeHook(notify concern.Notify) *HookResult {
	return defaultHookResult
}

func (d defaultHook) ShouldSendHook(notify concern.Notify) *HookResult {
	return defaultHookResult
}

var defaultHookResult = &HookResult{
	Pass:   false,
	Reason: "default hook",
}

var HookResultPass = &HookResult{
	Pass: true,
}
