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

func handleErr(conn *Connection, msg *Message, err error) {
	if msg != nil {
		// log the response
		newMessageLogger().
			messageType(msg.Type).
			message(msg.Body).
			duration(time.Since(msg.Timestamp)).
			log(err != nil)
		// send error response if we have an id
		errOther := conn.SendResponse(map[string]interface{}{
			"id":      msg.ID,
			"success": false,
			"error":   err.Error(),
		})
		// log error
		if errOther != nil {
			log.Errorf("%+v", errOther)
		}
	}
	// log error
	if err != nil {
		log.Errorf("%+v", err)
	}
}
