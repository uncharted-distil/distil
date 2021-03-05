//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package ws

import (
	"time"

	log "github.com/unchartedsoftware/plog"
)

func handleSuccess(conn *Connection, msg *Message, response map[string]interface{}) {
	// append msg id
	response["id"] = msg.ID
	// log the response
	newMessageLogger().
		messageType(msg.Type).
		message(msg.Body).
		duration(time.Since(msg.Timestamp)).
		log(true)
	// send response
	err := conn.SendResponse(response)
	if err != nil {
		log.Errorf("%+v", err)
	}
}
