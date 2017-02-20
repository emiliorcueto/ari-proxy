package ariproxy

import (
	"errors"
	"strings"

	"github.com/CyCoreSystems/ari-proxy/proxy"
)

func (s *Server) applicationData(reply string, req *proxy.Request) {
	app, err := s.ari.Application.Data(req.ApplicationData.Name)
	if err != nil {
		s.sendError(reply, err)
		return
	}
	if app.Name == "" {
		s.sendNotFound(reply)
		return
	}

	s.nats.Publish(reply, &app)
}

func (s *Server) applicationList(reply string, req *proxy.Request) {
	list, err := s.ari.Application.List()
	if err != nil {
		s.sendError(reply, err)
		return
	}

	resp := proxy.EntityList{List: []*proxy.Entity{}}
	for _, i := range list {
		resp.List = append(resp.List, &proxy.Entity{
			Metadata: s.Metadata(req.Metadata.Dialog),
			ID:       i.ID(),
		})
	}
	s.nats.Publish(reply, &resp)
}

func (s *Server) applicationGet(reply string, req *proxy.Request) {
	app, err := s.ari.Application.Data(req.ApplicationGet.Name)
	if err != nil {
		s.sendError(reply, err)
		return
	}
	if app.Name == "" {
		s.sendNotFound(reply)
		return
	}

	s.nats.Publish(reply, &proxy.Entity{
		Metadata: s.Metadata(req.Metadata.Dialog),
		ID:       app.Name,
	})
}

func parseEventSource(src string) (string, string, error) {
	var err error

	pieces := strings.Split(src, ":")
	if len(pieces) != 2 {
		return "", "", errors.New("Invalid EventSource")
	}

	switch pieces[0] {
	case "channel":
	case "bridge":
	case "endpoint":
	case "deviceState":
	default:
		err = errors.New("Unhandled EventSource type")
	}
	return pieces[0], pieces[1], err
}

func (s *Server) applicationSubscribe(reply string, req *proxy.Request) {
	err := s.ari.Application.Subscribe(req.ApplicationSubscribe.Name, req.ApplicationSubscribe.EventSource)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	if req.Metadata.Dialog != "" {
		eType, eID, err := parseEventSource(req.ApplicationSubscribe.EventSource)
		if err != nil {
			Log.Warn("failed to parse event source", "error", err, "eventsource", req.ApplicationSubscribe.EventSource)
		} else {
			s.Dialog.Bind(req.Metadata.Dialog, eType, eID)
		}
	}

	s.sendError(reply, nil)
}

func (s *Server) applicationUnsubscribe(reply string, req *proxy.Request) {
	err := s.ari.Application.Subscribe(req.ApplicationSubscribe.Name, req.ApplicationSubscribe.EventSource)
	if err != nil {
		s.sendError(reply, err)
		return
	}

	s.sendError(reply, nil)
}