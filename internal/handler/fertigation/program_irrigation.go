package fertigation

import (
	"errors"
	"net/http"

	"gr33n-api/internal/fertigation/programrules"
	"gr33n-api/internal/httputil"
)

func normalizeProgramFields(irrigationOnly bool, recipeID **int64) error {
	if err := programrules.ValidateCreateUpdate(irrigationOnly, *recipeID); err != nil {
		return err
	}
	if irrigationOnly {
		*recipeID = nil
	}
	return nil
}

func writeProgramValidationError(w http.ResponseWriter, err error) {
	if errors.Is(err, programrules.ErrIrrigationOnlyNoRecipe) {
		httputil.WriteError(w, http.StatusBadRequest,
			"This program is irrigation-only (plain water). Remove the nutrient recipe, or turn off irrigation-only.")
		return
	}
	httputil.WriteError(w, http.StatusBadRequest, err.Error())
}
