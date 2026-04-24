package service

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/pkg/pb/ipc"
	"google.golang.org/protobuf/proto"
)



func (s *MainService) ReceiveNATSMsg(m *nats.Msg) error {
	if m.Subject == config.TRACK_SUBJECT {
		baseMsg := ipc.BaseMessage{}
		err := proto.Unmarshal(m.Data, &baseMsg)
		if err != nil {
			log.Err(err).Msgf("Invalid Ipc msg from subj %v --> %v", m.Subject, m.Data)
			return err
		} else {
			s.ProcessIpcEvent(ipc.MessageType(baseMsg.MessageType), baseMsg.MessageDetail)
		}
	}

	return nil
}