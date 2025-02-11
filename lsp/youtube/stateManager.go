package youtube

import (
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/tidwall/buntdb"
	"time"
)

type StateManager struct {
	*concern_manager.StateManager
	*extraKey
}

func (s *StateManager) AddInfo(info *Info) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		infoKey := s.InfoKey(info.ChannelId)
		_, _, err := tx.Set(infoKey, info.ToString(), localdb.ExpireOption(time.Hour*24*7))
		return err
	})
}

func (s *StateManager) GetInfo(channelId string) (*Info, error) {
	info := new(Info)
	err := s.JsonGet(s.InfoKey(channelId), info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (s *StateManager) GetVideo(channelId string, videoId string) (*VideoInfo, error) {
	var v = new(VideoInfo)
	err := s.JsonGet(s.VideoKey(channelId, videoId), v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *StateManager) AddVideo(v *VideoInfo) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		key := s.VideoKey(v.ChannelId, v.VideoId)
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(key, string(b), localdb.ExpireOption(time.Hour*24))
		return err
	})
}

func (s *StateManager) Start() error {
	for _, pattern := range []localdb.KeyPatternFunc{s.GroupConcernStateKey, s.UserInfoKey, s.FreshKey} {
		s.CreatePatternIndex(pattern, nil)
	}
	return s.StateManager.Start()
}

func NewStateManager() *StateManager {
	sm := new(StateManager)
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern_manager.NewStateManager(NewKeySet(), true)
	return sm
}
