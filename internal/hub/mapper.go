package hub

import "google.golang.org/protobuf/proto"

func NewHubMessage(version int32, id string, apiUrl string) *HubMessage {
	return &HubMessage{
		Version: proto.Int32(version),
		Id:      proto.String(id),
		ApiUrl:  proto.String(apiUrl),
	}
}
