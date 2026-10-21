package nc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

func (i *Instance) GetUserConversations() (*[]NextcloudSpreedConversationData, APIResponse, error) {
	req, err := i.NewOCSRequest(http.MethodGet, i.baseUrl+"/ocs/v2.php/apps/spreed/api/v4/room", bytes.NewReader([]byte("")))
	if err != nil {
		return nil, APIUnreachable, err
	}
	req.SetBasicAuth(i.credentials.LoginName, i.credentials.AppPassword)

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, APIUnreachable, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, APILoginExpired, nil
	} else if resp.StatusCode == 503 {
		return nil, APIMaintenance, nil
	} else if resp.StatusCode != 200 {
		return nil, APIUnreachable, errors.New("unknown server response")
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, APIUnreachable, err
	}

	// Decode every conversation on its own, so that a single conversation
	// with unexpected data doesn't prevent the others from being processed.
	ncRes := NextcloudOCSBaseResult[[]json.RawMessage]{}
	if err = json.Unmarshal(body, &ncRes); err != nil {
		return nil, APIUnreachable, err
	}

	conversations := make([]NextcloudSpreedConversationData, 0, len(ncRes.OCS.Data))
	for _, rawConversation := range ncRes.OCS.Data {
		conv := NextcloudSpreedConversationData{}
		if err = json.Unmarshal(rawConversation, &conv); err != nil {
			// A field changing type between Nextcloud versions only leaves that field empty:
			// the rest of the conversation is still decoded and usable.
			var typeErr *json.UnmarshalTypeError
			if !errors.As(err, &typeErr) {
				log.Printf("%s: skipping a conversation that could not be decoded: %v", i.instanceName, err)
				continue
			}
		}
		conversations = append(conversations, conv)
	}

	return &conversations, APISuccess, nil
}
