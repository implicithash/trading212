package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"trading/pages"
)

// Handler for a routing
type Handler struct {
	DB          *sql.DB
	AccountPage *pages.AccountPage
}

// Status of the query
type Status int

// Declare typed constants each with type of status
const (
	Failed Status = iota
	Success
)

// Response to Add
type Response struct {
	Message string `json:"message"`
	Status  `json:"status"`
	ID      int64 `json:"id"`
}

func (h *Handler) findID(id int) (string, error) {
	rows, err := h.DB.Query("SELECT item_key FROM items WHERE item_id = ?;", id)
	if err != nil {
		return "", err
	}
	var guid string
	for rows.Next() {
		err = rows.Scan(&guid)
		if err != nil {
			return "", err
		}
	}
	defer rows.Close()
	return guid, nil
}

func (h *Handler) findItem(id int) (*pages.DbItem, error) {
	rows, err := h.DB.Query("SELECT * FROM items WHERE item_id = ?;", id)
	if err != nil {
		return nil, err
	}
	item := pages.DbItem{}
	for rows.Next() {
		err = rows.Scan(&item.ID, &item.Instrument, &item.GUID, &item.Dir, &item.Qty, &item.Price)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()
	return &item, nil
}

func (h *Handler) edit(item *pages.DbItem) error {
	cmd := "UPDATE items SET qty = ? WHERE item_id = ?"
	itemUpdate, err := h.DB.Prepare(cmd)
	if err != nil {
		return err
	}

	args := make([]interface{}, 0)
	args = append(args, item.Qty)
	args = append(args, item.ID)
	result, err := itemUpdate.Exec(args...)

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("Edit. RowsAffected: %d, id: %d", affected, item.ID)
	log.Debug(msg)

	return nil
}

func (h *Handler) deleteID(id int) error {
	deleteItem, err := h.DB.Prepare("DELETE FROM items WHERE item_id = ?")
	if err != nil {
		return err
	}
	result, err := deleteItem.Exec(id)
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("Delete. RowsAffected: %d, id: %d", affected, id)
	log.Debug(msg)
	return nil
}

// GetPosition godoc
// @Summary Get details of the position
// @Description Get details of the position
// @Tags positions
// @Accept  json
// @Produce  json
// @Param id path int true "Position ID"
// @Success 200 {object} pages.Position
// @Router /positions/{id} [get]
func (h *Handler) GetPosition(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	result, _ := strconv.Atoi(id)
	guid, err := h.findID(result)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	position, err := h.AccountPage.GetPosition(guid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	position.ID = result
	respondWithJSON(w, http.StatusOK, position)
}

// GetPositions godoc
// @Summary Get details of all positions
// @Description Get details of all positions
func (h *Handler) GetPositions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT item_id, item_key FROM items;")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	guids := make(map[int]string, 0)
	for rows.Next() {
		var guid string
		var id int
		err = rows.Scan(&id, &guid)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		guids[id] = guid
	}
	defer rows.Close()

	positions := make([]*pages.Position, 0)
	wg := &sync.WaitGroup{}
	for id, guid := range guids {
		wg.Add(1)
		go func(id int, guid string, wg *sync.WaitGroup) {
			defer wg.Done()
			position, err := h.AccountPage.GetPosition(guid)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			position.ID = id
			positions = append(positions, position)
		}(id, guid, wg)
	}
	wg.Wait()
	respondWithJSON(w, http.StatusOK, positions)
}

// DeletePosition deletes an opened position
func (h *Handler) DeletePosition(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	data := params["id"]
	id, _ := strconv.Atoi(data)
	guid, err := h.findID(id)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if guid == "" {
		msg := fmt.Sprintf(pages.GUIDNotFound, id)
		respondWithError(w, http.StatusNotFound, msg)
		return
	}

	err = h.AccountPage.DeletePosition(guid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = h.deleteID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	response := &Response{ID: int64(id), Message: "Item is deleted", Status: Success}
	respondWithJSON(w, http.StatusOK, response)
}

// EditPosition edits an opened position
func (h *Handler) EditPosition(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	data := params["id"]
	id, _ := strconv.Atoi(data)
	item, err := h.findItem(id)

	bytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(bytes, &jsonMap)

	position, err := h.AccountPage.EditPosition(item, jsonMap)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = h.edit(position)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	response := &Response{ID: int64(position.ID), Message: "Item is edited", Status: Success}
	respondWithJSON(w, http.StatusOK, response)
}

// Add godoc
// @Summary Create a new position
// @Description Create a new position with the input data
// @Tags positions
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /positions [post]
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(bytes, &jsonMap)

	item, err := h.AccountPage.Add(jsonMap)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result, err := h.DB.Exec(
		"INSERT INTO items (`instrument`, `item_key`, `direction`, `qty`, `price`) VALUES (?, ?, ?, ?, ?)",
		item.Instrument, item.Key, item.Direction, item.Qty, item.Price,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		respondWithError(w, http.StatusNotModified, err.Error())
		return
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		respondWithError(w, http.StatusNotModified, err.Error())
		return
	}
	log.Debug(fmt.Sprintf("Insert: rowsAffected: %d, lastInsertedId: %d", affected, lastID))

	response := &Response{ID: lastID, Message: "Item is added", Status: Success}
	respondWithJSON(w, http.StatusOK, response)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// GetOrder godoc
// @Summary Get details of the order
// @Description Get details of the order
/*func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
}

// GetOrders godoc
// @Summary Get details of all orders
// @Description Get details of all orders
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
}

// DeleteOrder deletes an opened order
func (h *Handler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
}*/
