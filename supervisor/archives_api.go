package supervisor

import (
	"encoding/json"
	"fmt"
	"github.com/pborman/uuid"
	"github.com/starkandwayne/shield/db"
	"net/http"
	"regexp"
)

type ArchiveAPI struct {
	Data       *db.DB
	ResyncChan chan int
	AdhocChan  chan AdhocTask
}

func (self ArchiveAPI) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch {
	case match(req, `GET /v1/archives`):
		archives, err := self.Data.GetAllAnnotatedArchives(
			&db.ArchiveFilter{
				ForTarget: paramValue(req, "target", ""),
				ForStore:  paramValue(req, "store", ""),
				Before:    paramDate(req, "before"),
				After:     paramDate(req, "after"),
			},
		)
		if err != nil {
			bail(w, err)
			return
		}

		JSON(w, archives)
		return

	case match(req, `POST /v1/archive/[a-fA-F0-9-]+/restore`):
		if req.Body == nil {
			w.WriteHeader(400)
			return
		}

		var params struct {
			Target string `json:"target"`
			Owner  string `json:"owner"`
		}
		json.NewDecoder(req.Body).Decode(&params)

		if params.Owner == "" {
			params.Owner = "anon"
		}

		re := regexp.MustCompile(`^/v1/archive/([a-fA-F0-9-]+)/restore`)
		id := uuid.Parse(re.FindStringSubmatch(req.URL.Path)[1])

		// find the archive
		archive, err := self.Data.GetAnnotatedArchive(id)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		if params.Target == "" {
			params.Target = archive.TargetUUID
		}

		tid := uuid.Parse(params.Target)
		// find the target
		_, err = self.Data.GetAnnotatedTarget(id)
		if err != nil {
			w.WriteHeader(501)
			return
		}

		// tell the supervisor to schedule a task
		self.AdhocChan <- AdhocTask{
			Op:          RESTORE,
			Owner:       params.Owner,
			TargetUUID:  tid,
			ArchiveUUID: id,
		}

		w.WriteHeader(200)
		JSONLiteral(w, fmt.Sprintf(`{"ok":"scheduled"}`))
		return

	case match(req, `PUT /v1/archive/[a-fA-F0-9-]+`):
		if req.Body == nil {
			w.WriteHeader(400)
			return
		}

		var params struct {
			Notes string `json:"notes"`
		}
		json.NewDecoder(req.Body).Decode(&params)

		if params.Notes == "" {
			w.WriteHeader(400)
			return
		}

		re := regexp.MustCompile(`^/v1/archive/([a-fA-F0-9-]+)`)
		id := uuid.Parse(re.FindStringSubmatch(req.URL.Path)[1])

		_ = self.Data.AnnotateArchive(id, params.Notes)
		self.ResyncChan <- 1
		JSONLiteral(w, fmt.Sprintf(`{"ok":"updated"}`))
		return

	case match(req, `DELETE /v1/archive/[a-fA-F0-9-]+`):
		re := regexp.MustCompile(`^/v1/archive/([a-fA-F0-9-]+)`)
		id := uuid.Parse(re.FindStringSubmatch(req.URL.Path)[1])

		deleted, err := self.Data.DeleteArchive(id)

		if err != nil {
			bail(w, err)
		}
		if !deleted {
			w.WriteHeader(403)
			return
		}

		self.ResyncChan <- 1
		JSONLiteral(w, fmt.Sprintf(`{"ok":"deleted"}`))
		return
	}

	w.WriteHeader(415)
	return
}
