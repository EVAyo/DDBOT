package huya

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLiveInfo(t *testing.T) {
	l := &LiveInfo{
		RoomId:   test.NAME1,
		Name:     test.NAME2,
		RoomName: test.NAME2,
	}
	assert.Equal(t, test.NAME2, l.GetName())
	assert.Equal(t, Live, l.Type())
	notify := NewConcernLiveNotify(test.G1, l)
	assert.NotNil(t, notify)
	assert.NotNil(t, notify.Logger())
	assert.Equal(t, test.G1, notify.GetGroupCode())
	assert.Equal(t, test.NAME1, notify.GetUid())
	assert.Equal(t, concern.HuyaLive, notify.Type())
}
