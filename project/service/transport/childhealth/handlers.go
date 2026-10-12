package childhealthgt

import (
	"encoding/json"
	"errors"
	"net/http"

	hp "github.com/Johncode248/Rodzic-Chat/project/common/childhealth"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/childhealth"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func account(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || id == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return "", false
	}
	return id, true
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, childhealth.ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, childhealth.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, childhealth.ErrInvalid), errors.Is(err, childhealth.ErrEpisodeOpen):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func decode[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return v, false
	}
	return v, true
}

func reply(w http.ResponseWriter, v any, err error) {
	if err != nil {
		fail(w, err)
		return
	}
	json.NewEncoder(w).Encode(v)
}

func ok(w http.ResponseWriter, err error) {
	if err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func Overview(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		v, err := childhealth.GetOverview(r.Context(), a, r.URL.Query().Get("child_id"))
		reply(w, v, err)
	}
}

func SaveProfile(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		if req, good := decode[hp.ProfileRequest](w, r); good {
			ok(w, childhealth.SaveProfile(r.Context(), a, req))
		}
	}
}

func AddAllergy(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		if req, good := decode[hp.AllergyRequest](w, r); good {
			v, err := childhealth.AddAllergy(r.Context(), a, req)
			reply(w, v, err)
		}
	}
}

func AddDoctor(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		if req, good := decode[hp.DoctorRequest](w, r); good {
			v, err := childhealth.AddDoctor(r.Context(), a, req)
			reply(w, v, err)
		}
	}
}

func AddMeasurement(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		if req, good := decode[hp.MeasurementRequest](w, r); good {
			v, err := childhealth.AddMeasurement(r.Context(), a, req)
			reply(w, v, err)
		}
	}
}

func deleter(table string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a, good := account(w, r); good {
			ok(w, childhealth.Delete(r.Context(), a, table, r.URL.Query().Get("id")))
		}
	}
}

var (
	DeleteAllergy     = deleter("child_allergies")
	DeleteDoctor      = deleter("child_doctors")
	DeleteMeasurement = deleter("child_measurements")
)

func Illness(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		v, err := childhealth.GetIllness(r.Context(), a, r.URL.Query().Get("child_id"))
		reply(w, v, err)
	}
}

func Episode(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		v, err := childhealth.GetEpisode(r.Context(), a, r.URL.Query().Get("id"))
		reply(w, v, err)
	}
}

func StartEpisode(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		if req, good := decode[hp.EpisodeRequest](w, r); good {
			v, err := childhealth.StartEpisode(r.Context(), a, req)
			reply(w, v, err)
		}
	}
}

func EndEpisode(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		ok(w, childhealth.EndEpisode(r.Context(), a, r.URL.Query().Get("id")))
	}
}

func AddEntry(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		if req, good := decode[hp.EntryRequest](w, r); good {
			v, err := childhealth.AddEntry(r.Context(), a, req)
			reply(w, v, err)
		}
	}
}

func DeleteEntry(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		ok(w, childhealth.DeleteEntry(r.Context(), a, r.URL.Query().Get("id")))
	}
}

func AddMedication(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		if req, good := decode[hp.MedicationRequest](w, r); good {
			v, err := childhealth.AddMedication(r.Context(), a, req)
			reply(w, v, err)
		}
	}
}

func StopMedication(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		ok(w, childhealth.StopMedication(r.Context(), a, r.URL.Query().Get("id")))
	}
}

func GiveDose(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		v, err := childhealth.GiveDose(r.Context(), a, r.URL.Query().Get("id"))
		reply(w, v, err)
	}
}

func SkipDose(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		ok(w, childhealth.SkipDose(r.Context(), a, r.URL.Query().Get("id")))
	}
}

func AdHocDose(w http.ResponseWriter, r *http.Request) {
	if a, good := account(w, r); good {
		v, err := childhealth.GiveAdHoc(r.Context(), a, r.URL.Query().Get("id"))
		reply(w, v, err)
	}
}
