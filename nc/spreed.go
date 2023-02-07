package nc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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
	} else if resp.StatusCode != 200 {
		return nil, APIUnreachable, errors.New("unknown server response")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, APIUnreachable, err
	}

	ncRes := NextcloudOCSBaseResult[[]NextcloudSpreedConversationData]{}
	if err = json.Unmarshal(body, &ncRes); err != nil {
		return nil, APIUnreachable, err
	}

	return &ncRes.OCS.Data, APISuccess, nil
}
