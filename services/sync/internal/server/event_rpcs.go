package server

import (
	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) SubscribeEvents(req *syncv1.SubscribeEventsRequest, stream syncv1.SyncService_SubscribeEventsServer) error {
	sub := s.eventBus.Subscribe(req.EventTypes)
	defer s.eventBus.Unsubscribe(sub)

	for {
		select {
		case event, ok := <-sub.Ch:
			if !ok {
				return nil
			}
			if err := stream.Send(eventToProto(event)); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func eventToProto(e service.Event) *syncv1.SyncEvent {
	return &syncv1.SyncEvent{
		Type:      e.Type,
		Timestamp: timestamppb.New(e.Timestamp),
		Payload:   e.Payload,
	}
}
