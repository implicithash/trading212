package pages

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// AccountPage represents an account page
type AccountPage struct {
	Page Page
}

const (
	// POSITIONS const
	POSITIONS = "positions"
	// ORDERS const
	ORDERS = "orders"
	// SELL const
	SELL = "sell"
	// BUY const
	BUY = "buy"
	// SortASC const
	SortASC = "sort-ascending"
)

// DbItem struct
type DbItem struct {
	ID         int
	Instrument string
	GUID       string
	Dir        string
	Qty        int
	Price      float64
}

// Limit struct
type Limit struct {
	IsUse    bool    `json:"is_use"`
	Price    float64 `json:"price"`
	Distance float64 `json:"distance"`
	Result   float64 `json:"result"`
}

// BasePosition is a base entity
type BasePosition struct {
	Instrument   string  `json:"instrument"`
	ID           int     `json:"id"`
	Quantity     int     `json:"quantity"`
	Direction    string  `json:"direction"`
	Price        float64 `json:"price"`
	CurrentPrice float64 `json:"current_price"`
	TakeProfit   string  `json:"take_profit"`
	StopLoss     string  `json:"stop_loss"`
	DateCreated  string  `json:"date_created"`
}

// Position on the Stock Exchange
type Position struct {
	BasePosition
	Margin       float64 `json:"margin"`
	TrailingStop string  `json:"trailing_stop"`
	Result       float64 `json:"result"`
}

// PositionPayload is for position editing
type PositionPayload struct {
	Direction string  `json:"direction"`
	IsPercent bool    `json:"is_percent"`
	Percent   float64 `json:"percent"`
	Value     float64 `json:"value"`
}

// Order on the Stock Exchange
type Order struct {
	BasePosition
	Type string
}

// Item represents common data
type Item struct {
	Instrument string
	Key        string
	Qty        int
	Price      float64
	Direction  string
	IsOrder    bool
	Limits     map[string]*Limit
}

type orderWindow struct {
	Page     *Page
	State    string
	Insfunds bool
	Item     *Item
}

var (
	qtyRe           = regexp.MustCompile(`\A\d+ d+@\z`)
	isAllProperties = false
)

func (p *AccountPage) checkSessionExpired() error {
	widget := p.Page.FindElementByCSS(domPaths["widget_message"])
	if widget != nil {
		we, _ := widget.FindElement(selenium.ByCSSSelector, domPaths["css_text"])
		text, _ := we.Text()
		if strings.Contains(text, sessionExpired) {
			log.Debug("Session is expired")
			if we, err := widget.FindElement(selenium.ByCSSSelector, domPaths["ok"]); err == nil {
				we.Click()
				//time.Sleep(time.Millisecond * 3000)
			}
		}
	}
	err := p.Page.Driver.WaitWithTimeout(p.Page.NotFound(selenium.ByCSSSelector, domPaths["widget_message"]), time.Second*3)
	if err != nil {
		return fmt.Errorf(sessionExpired)
	}
	return nil
}

func (p *AccountPage) checkDateSortDescending() error {
	created := p.Page.FindElementByCSS(domPaths["date_created"])
	p.checkAttr(created, SortASC)
	return nil
}

func (p *AccountPage) checkAttr(we selenium.WebElement, attr string) error {
	if we != nil {
		class, _ := we.GetAttribute("class")
		if !strings.Contains(class, attr) {
			we.Click()
			time.Sleep(time.Millisecond * 200)
		}
	}
	return nil
}

func (p *AccountPage) switchAll() error {
	if isAllProperties {
		return nil
	}

	tableCtxMenu := p.Page.FindElementByCSS(domPaths["settings"])
	if tableCtxMenu == nil {
		return nil
	}
	tableCtxMenu.Click()
	time.Sleep(time.Millisecond * 300)

	ctxItems := []string{"ctx_name", "ctx_qty", "ctx_dir", "ctx_price", "ctx_curprice", "ctx_tp", "ctx_sl", "cxt_ts", "ctx_margin", "ctx_datecreated", "ctx_result"}

	ctxMenu := p.Page.FindElementByXpath(domPaths["dt_ctxmenu"])
	if ctxMenu == nil {
		return nil
	}

	wg := &sync.WaitGroup{}
	for _, ctxItem := range ctxItems {
		wg.Add(1)
		go func(ctxItem string, wg *sync.WaitGroup) {
			defer wg.Done()
			if menuItem, err := ctxMenu.FindElement(selenium.ByCSSSelector, domPaths[ctxItem]); err == nil && menuItem != nil {
				p.checkAttr(menuItem, "selected")
			}
		}(ctxItem, wg)
	}
	wg.Wait()
	isAllProperties = true

	return nil
}

func (p *AccountPage) findItem(target string, id string) (selenium.WebElement, error) {
	itemPath := fmt.Sprintf("#item-%s", id)
	item := p.Page.FindElementByCSS(itemPath)
	if item == nil {
		err := p.switchTab(POSITIONS)
		if err != nil {
			return nil, err
		}
		item := p.Page.FindElementByCSS(itemPath)
		if item == nil {
			return nil, fmt.Errorf(fmt.Sprintf(positionNotFound, id))
		}
	}
	return item, nil
}

// GetPosition returns an opened position
func (p *AccountPage) GetPosition(id string) (*Position, error) {
	p.checkSessionExpired()
	p.switchAll()

	log.Infof(fmt.Sprintf("Get a position: %s", id))
	wePosition, err := p.findItem(POSITIONS, id)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf(positionNotFound, id))
	}
	wePosition.Click()
	time.Sleep(time.Millisecond * 300)

	dlg := &orderWindow{Page: &p.Page, Item: nil, State: "init"}
	err = dlg.info()
	if err != nil {
		return nil, err
	}
	position, err := dlg.getInfo()
	if err != nil {
		return nil, err
	}

	return position, nil
}

// DeletePosition deletes an opened position
func (p *AccountPage) DeletePosition(id string) error {
	p.checkSessionExpired()
	err := p.Delete(id, POSITIONS)
	return err
}

// EditPosition edits an opened position
func (p *AccountPage) EditPosition(item *DbItem, args interface{}) (*DbItem, error) {
	p.checkSessionExpired()
	payload, err := p.initEdit(args)
	if payload == nil {
		return nil, err
	}
	log.Infof(fmt.Sprintf("Edit: %#v", args))
	position, err := p.findItem(POSITIONS, item.GUID)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf(positionNotFound, item.GUID))
	}
	position.Click()

	dlg := &orderWindow{Page: &p.Page, Item: nil, State: "init"}
	err = dlg.edit()
	if err != nil {
		return nil, err
	}
	qty, err := dlg.editQuantity(payload)
	if err != nil {
		return nil, err
	}
	err = dlg.confirm()
	if err != nil {
		return nil, err
	}
	name := POSITIONS
	log.WithFields(log.Fields{
		"id":         item.ID,
		"instrument": item.Instrument,
		"quantity":   item.Qty,
	}).Info(fmt.Sprintf("edited a %s", name[:len(name)-1]))
	item.Qty = item.Qty + qty

	return item, nil
}

// DeleteOrder deletes an opened order
func (p *AccountPage) DeleteOrder(id string) error {
	p.checkSessionExpired()
	err := p.Delete(id, ORDERS)
	return err
}

// Delete deletes an item
func (p *AccountPage) Delete(id string, target string) error {
	itemPath := fmt.Sprintf("#item-%s", id)
	_, err := p.findItem(target, id)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf(positionNotFound, id))
	}

	p.Page.MouseHoverToElement(itemPath)
	time.Sleep(time.Millisecond * 100)
	p.Page.Driver.Click(selenium.RightButton)
	time.Sleep(time.Millisecond * 200)

	we := p.Page.FindElementByCSS(domPaths["cxtmenu"])
	if we == nil {
		return fmt.Errorf(fmt.Sprintf(cssError, domPaths["ctxmenu"]))
	}
	menuItem := fmt.Sprintf(domPaths["rm_item"], target)
	rm, _ := we.FindElement(selenium.ByCSSSelector, menuItem)
	if rm == nil {
		return fmt.Errorf(fmt.Sprintf(cssError, menuItem))
	}
	rm.MoveTo(0, 0)
	rm.Click()
	time.Sleep(time.Millisecond * 100)
	widget := p.Page.FindElementByCSS(domPaths["widget_message"])
	if okBtn, err := widget.FindElement(selenium.ByCSSSelector, domPaths["ok_btn"]); err == nil {
		okBtn.Click()
	}
	widget = p.Page.FindElementByCSS(domPaths["widget_message"])
	if widget != nil {
		if txt, err := widget.FindElement(selenium.ByCSSSelector, domPaths["css_text"]); err == nil {
			return fmt.Errorf(txt.Text())
		}
	}
	return nil
}

func (p *AccountPage) initEdit(args interface{}) (*PositionPayload, error) {
	in := &PositionPayload{}
	data, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf(wrongTypeInput, args))
	}
	jsonbody, _ := json.Marshal(data["quantity"])
	err := json.Unmarshal(jsonbody, &in)
	if err != nil {
		return nil, fmt.Errorf("error while parsing args")
	}
	return in, nil
}

func (p *AccountPage) initAdd(args interface{}) (*Item, error) {
	in := &Item{}
	data, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf(wrongTypeInput, args))
	}
	instrument, ok := data["instrument"].(string)
	if !ok {
		return nil, fmt.Errorf(instrumentNotDefined)
	}
	direction, ok := data["direction"].(string)
	if !ok || direction == "" {
		return nil, fmt.Errorf(directionNotDefined)
	}
	isOrder, ok := data["is_order"].(bool)
	if !ok {
		return nil, fmt.Errorf(typeNotDefined)
	}
	qty, ok := data["qty"].(float64)
	if !ok {
		return nil, fmt.Errorf("qty must be int")
	}
	price, ok := data["price"].(float64)
	if !ok && isOrder {
		return nil, fmt.Errorf("price not found")
	}

	var limits map[string]interface{}
	if data["limits"] != nil {
		limits = data["limits"].(map[string]interface{})
	}
	tp := Limit{}
	jsonbody, _ := json.Marshal(limits["tp"])
	err := json.Unmarshal(jsonbody, &tp)
	if err != nil {
		return nil, fmt.Errorf("error while parsing limits")
	}
	in.Limits = make(map[string]*Limit, 0)
	in.Limits["tp"] = &tp

	sl := Limit{}
	jsonbody, _ = json.Marshal(limits["sl"])
	err = json.Unmarshal(jsonbody, &sl)
	if err != nil {
		return nil, fmt.Errorf("error while parsing limits")
	}
	in.Limits["sl"] = &sl

	in.Instrument = instrument
	in.Direction = direction
	in.Qty = int(qty)
	in.IsOrder = isOrder
	in.Price = price

	return in, nil
}

// Add adds a new position/order
func (p *AccountPage) Add(args interface{}) (*Item, error) {
	p.checkSessionExpired()
	p.checkDateSortDescending()

	log.Infof(fmt.Sprintf("Add: %#v", args))

	item, err := p.initAdd(args)
	if item == nil {
		return nil, fmt.Errorf(inputDataErrors)
	}

	dlg := &orderWindow{Page: &p.Page, Item: item, State: "init"}
	err = dlg.open()
	if err != nil {
		return nil, err
	}
	err = dlg.setDirection(item.Direction)
	if err != nil {
		return nil, err
	}
	// set quantity
	if item.Qty != 0 {
		dlg.setQuantity(item.Qty)
	}

	// set limits
	if item.Limits != nil {
		dlg.setLimit(item.Limits)
	}
	err = dlg.confirm()
	if err != nil {
		return nil, err
	}
	var name string
	if item.IsOrder {
		name = ORDERS
	} else {
		name = POSITIONS
	}
	id, err := p.findKey(name)
	item.Key = id

	log.WithFields(log.Fields{
		"instrument": item.Instrument,
		"quantity":   item.Qty,
	}).Info(fmt.Sprintf("added a new %s", name[:len(name)-1]))

	return item, nil
}

// findID finds a real guid id
func (p *AccountPage) findKey(name string) (string, error) {
	var guid string
	p.switchTab(name)
	rowPath := fmt.Sprintf("%s_last_row", name)
	lastRow := p.Page.FindElementByCSS(domPaths[rowPath])
	if lastRow == nil {
		return guid, fmt.Errorf(positionTableEmpty)
	}
	id, _ := lastRow.GetAttribute("id")
	guid = strings.Replace(id, "item-", "", 1)
	return guid, nil
}

func (p *AccountPage) switchTab(name string) error {
	tabName := fmt.Sprintf("tab_%s", name)
	tab := p.Page.FindElementByCSS(domPaths[tabName])
	if tab == nil {
		return fmt.Errorf(cssError, tabName)
	}
	tab.Click()
	return nil
}

func (w *orderWindow) open() error {
	if ok, _ := w.Page.FindElementByCSS(domPaths["add_order"]).IsDisplayed(); ok {
		w.Page.FindElementByCSS(domPaths["add_order"]).Click()
	} else {
		w.Page.FindElementByCSS(domPaths["dt_no_data"]).Click()
	}
	log.Infof("opened window")
	time.Sleep(time.Millisecond * 200)

	we := w.Page.FindElementByCSS(domPaths["search_box"])
	if we == nil {
		return fmt.Errorf(fmt.Sprintf(cssError, domPaths["search_box"]))
	}
	we.SendKeys(w.Item.Instrument)
	if w.getResult(0) == nil {
		w.Page.FindElementByCSS(domPaths["close"]).Click()
		return fmt.Errorf(instrumentNotDefined)
	}
	result, _, err := w.searchResult(w.Item.Instrument)
	if err != nil {
		return err
	}
	result.Click()
	widgetMsg := w.Page.FindElementByCSS(domPaths["widget_message"])
	if widgetMsg != nil {
		w.decode(widgetMsg)
	}
	w.State = "open"
	return nil
}

func (w *orderWindow) edit() error {
	if dlg := w.Page.FindElementByCSS(domPaths["dlg"]); dlg == nil {
		return fmt.Errorf(fmt.Sprintf(cssError, domPaths["dlg"]))
	}
	w.State = "edit"
	return nil
}

func (w *orderWindow) info() error {
	if dlg := w.Page.FindElementByCSS(domPaths["dlg"]); dlg == nil {
		return fmt.Errorf(fmt.Sprintf(cssError, domPaths["dlg"]))
	}
	w.State = "info"
	return nil
}

func (w *orderWindow) close() error {
	err := w.checkOpen()
	if err != nil {
		return err
	}
	w.Page.FindElementByCSS(domPaths["close"]).Click()
	w.State = "closed"
	log.Debug("closed window")
	return nil
}

func (w *orderWindow) confirm() error {
	err := w.checkOpen()
	if err != nil {
		return err
	}
	confirmBtn := w.Page.FindElementByCSS(domPaths["confirm_btn"])
	if confirmBtn == nil {
		return fmt.Errorf(fmt.Sprintf(cssError, domPaths["confirm_btn"]))
	}
	confirmBtn.Click()
	time.Sleep(time.Millisecond * 1000)

	var txt string
	widget := w.Page.FindElementByCSS(domPaths["widget_message"])
	if widget != nil {
		we, _ := widget.FindElement(selenium.ByCSSSelector, domPaths["css_text"])
		txt, _ = we.Text()
	}
	confirmBtn = w.Page.FindElementByCSS(domPaths["confirm_btn"])

	if confirmBtn != nil {
		log.Error(txt)
		return fmt.Errorf(txt)
	}
	log.Info("Operation is confirmed")
	w.State = "confirmed"
	return nil
}

func (w *orderWindow) getResearchName(result selenium.WebElement) (string, error) {
	if result == nil {
		return "", nil
	}
	we, err := result.FindElement(selenium.ByCSSSelector, domPaths["instrument_name"])
	txt, _ := we.Text()
	if err != nil {
		log.Errorf(instrumentNotFound)
		return "", fmt.Errorf(instrumentNotFound)
	}
	return txt, nil
}

// decode text pop-up
func (w *orderWindow) decode(we selenium.WebElement) error {
	titleWebElem, _ := we.FindElement(selenium.ByCSSSelector, domPaths["css_title"])
	title, _ := titleWebElem.Text()
	textWebElem, _ := we.FindElement(selenium.ByCSSSelector, domPaths["css_text"])
	text, _ := textWebElem.Text()

	if title == "Insufficient Funds" {
		w.Insfunds = true
	} else if title == "Maximum Quantity Limit" {
		qty, _ := strconv.Atoi(text)
		return fmt.Errorf(fmt.Sprintf(maxQtyLimit, qty))
	} else if title == "Minimum Quantity Limit" {
		qty, _ := strconv.Atoi(text)
		return fmt.Errorf(fmt.Sprintf(minQtyLimit, qty))
	}
	log.Debug("decoded message")
	return nil
}

func (w *orderWindow) checkName(what, where string) bool {
	what = strings.ToLower(what)
	where = strings.ToLower(where)
	return strings.Contains(where, what)
}

func (w *orderWindow) searchResult(product string) (selenium.WebElement, string, error) {
	log.Infof("searching result...")
	result := w.getResult(0)
	name, err := w.getResearchName(result)
	if err != nil {
		return nil, "", err
	}
	pos := 0
	for {
		if ok := w.checkName(product, name); !ok {
			name, err = w.getResearchName(w.getResult(pos))
			if err != nil {
				return nil, "", err
			}
			if name == "" {
				w.Page.FindElementByCSS(domPaths["close"]).Click()
				msg := fmt.Sprintf(instrumentNotFound, product)
				return nil, "", errors.New(msg)
			}
			log.Debug(name)

			if w.checkName(product, name) {
				return w.getResult(pos), name, nil
			}
			pos++
		} else {
			break
		}
	}
	msg := fmt.Sprintf(foundProductAtPos, (pos + 1))
	log.Debug(msg)
	return result, name, nil
}

func (w *orderWindow) getResult(pos int) selenium.WebElement {
	// get pos result, where 0 is first
	evalXPath := fmt.Sprintf("%s[%d]", domPaths["result_instrument"], pos+1)

	if result := w.Page.FindElementByXpath(evalXPath); result != nil {
		return result
	}
	return nil
}

func (w *orderWindow) checkOpen() error {
	if w.State == "open" || w.State == "edit" || w.State == "info" {
		return nil
	}
	return fmt.Errorf(dialogNotOpened)
}

// Set direction (buy or sell)
func (w *orderWindow) setDirection(direction string) error {
	err := w.checkOpen()
	if err != nil {
		return err
	}
	if direction != BUY && direction != SELL {
		return fmt.Errorf(fmt.Sprintf(unacceptableValue, direction))
	}
	cssMode := fmt.Sprintf(domPaths["mode-btn"], direction)
	modeBtn := w.Page.FindElementByCSS(cssMode)
	modeBtn.Click()
	log.Debug("direction set")
	return nil
}

// Get current quantity
func (w *orderWindow) getQuantity() (int, error) {
	err := w.checkOpen()
	if err != nil {
		return 0, err
	}
	we := w.Page.FindElementByCSS(domPaths["qty_value"])
	if we == nil {
		return 0, fmt.Errorf(fmt.Sprintf(cssError, domPaths["qty_value"]))
	}
	txt, err := we.Text()
	results := strings.Split(txt, "@")
	if len(results) == 0 {
		return 0, fmt.Errorf(fmt.Sprintf(cssError, domPaths["qty_value"]))
	}
	sQty := strings.Replace(results[0], " ", "", -1)
	qty, err := strconv.Atoi(sQty)
	if err != nil {
		return 0, err
	}
	return qty, err
}

func (w *orderWindow) calcQuantity(params *PositionPayload, currQty int) (int, error) {
	var qty int
	if params.IsPercent {
		if params.Percent != 0 {
			qty = int(float64(currQty) * params.Percent / 100)
			if params.Direction == SELL {
				return qty, nil
			}
			if qty > currQty {
				return 0, fmt.Errorf(buyNotAllowed)
			}
			return qty, nil
		}
	} else {
		if params.Value != 0 {
			qty = int(params.Value)
			if params.Direction == SELL {
				return qty, nil
			}
			if qty > currQty {
				return 0, fmt.Errorf(buyNotAllowed)
			}
			return qty, nil
		}
	}
	return qty, nil
}

func (w *orderWindow) getInfo() (*Position, error) {
	err := w.checkOpen()

	if err != nil {
		return nil, err
	}
	// set INFO tab
	infoTab := w.Page.FindElementByCSS(domPaths["info_tab"])
	if infoTab == nil {
		log.Debug("Info tab is not found")
		return nil, fmt.Errorf(fmt.Sprintf(cssError, domPaths["info_tab"]))
	}
	infoTab.Click()
	log.Debug("Start get info...")

	position := &Position{}
	position.Instrument = w.Page.FindCSSValue(selenium.ByCSSSelector, domPaths["name"])
	position.DateCreated = w.Page.FindCSSValue(selenium.ByXPATH, domPaths["created"])

	str := w.Page.FindCSSValue(selenium.ByXPATH, domPaths["qty"])
	str = strings.Replace(str, " ", "", 1)
	if qty, err := strconv.Atoi(str); err == nil {
		position.Quantity = qty
	}

	position.Direction = w.Page.FindCSSValue(selenium.ByXPATH, domPaths["direction"])

	str = w.Page.FindCSSValue(selenium.ByXPATH, domPaths["avg_price"])
	str = strings.Replace(str, " ", "", 1)
	if price, err := strconv.ParseFloat(str, 32); err == nil {
		position.Price = price
	}
	str = w.Page.FindCSSValue(selenium.ByXPATH, domPaths["cur_price"])
	str = strings.Replace(str, " ", "", 1)
	if price, err := strconv.ParseFloat(str, 32); err == nil {
		position.CurrentPrice = price
	}
	position.TakeProfit = w.Page.FindCSSValue(selenium.ByXPATH, domPaths["take_profit"])
	position.StopLoss = w.Page.FindCSSValue(selenium.ByXPATH, domPaths["stop_loss"])
	position.TrailingStop = w.Page.FindCSSValue(selenium.ByXPATH, domPaths["trailing_stop"])

	str = w.Page.FindCSSValue(selenium.ByXPATH, domPaths["margin"])
	str = strings.Replace(str, " ", "", 1)
	if limit, err := strconv.ParseFloat(str, 32); err == nil {
		position.Margin = limit
	}

	// close a dialog
	if close := w.Page.FindElementByCSS(domPaths["info_close"]); close != nil {
		close.Click()
	}

	return position, nil
}

// Edit quantity
func (w *orderWindow) editQuantity(params *PositionPayload) (int, error) {
	err := w.checkOpen()

	if err != nil {
		return 0, err
	}
	if params.Direction != SELL && params.Direction != BUY {
		return 0, fmt.Errorf(directionNotDefined)
	}
	// set MARKET ORDER tab
	marketOrderTab := w.Page.FindElementByCSS(domPaths["market_order_tab"])
	if marketOrderTab == nil {
		return 0, fmt.Errorf(fmt.Sprintf(cssError, domPaths["market_order_tab"]))
	}
	marketOrderTab.Click()
	// set the direction
	err = w.setDirection(params.Direction)
	if err != nil {
		return 0, err
	}

	currQty, err := w.getQuantity()
	if err != nil {
		return 0, err
	}

	// calc qty according input parameters
	qty, err := w.calcQuantity(params, currQty)
	if err != nil {
		return 0, err
	}

	we := w.Page.FindElementByXpath(domPaths["qty_input_xpath"])
	if we == nil {
		return 0, fmt.Errorf(fmt.Sprintf(cssError, domPaths["qty_input_xpath"]))
	}
	we.Click()

	for _, el := range strconv.Itoa(qty) {
		w.Page.FindElementByXpath(domPaths["qty_input_xpath"]).SendKeys(string(el))
		time.Sleep(time.Millisecond * 50)
	}

	cssQty := w.Page.FindElementByCSS(domPaths["input_qty_val"])
	txt, err := cssQty.Text()

	log.Debug(fmt.Sprintf("Edit. added qty: %s", txt))
	return qty, nil
}

// Set quantity
func (w *orderWindow) setQuantity(qty int) error {
	err := w.checkOpen()
	if err != nil {
		return err
	}
	we := w.Page.FindElementByXpath(domPaths["qty_input_xpath"])
	if we == nil {
		return fmt.Errorf(fmt.Sprintf(cssError, domPaths["qty_input_xpath"]))
	}
	we.Click()

	for _, el := range strconv.Itoa(qty) {
		w.Page.FindElementByXpath(domPaths["qty_input_xpath"]).SendKeys(string(el))
		time.Sleep(time.Millisecond * 50)
	}

	cssQty := w.Page.FindElementByCSS(domPaths["input_qty_val"])
	txt, err := cssQty.Text()

	log.Debug(fmt.Sprintf("Add. quantity set: %s", txt))
	return nil
}

// set limit in order window
func (w *orderWindow) setLimit(limits map[string]*Limit) error {
	err := w.checkOpen()
	if err != nil {
		return err
	}
	tpUse := limits["tp"].IsUse
	slUse := limits["sl"].IsUse
	if !w.Item.IsOrder {
		if tpUse {
			w.buttonToggle("market_tp_toggle")
		}
		if slUse {
			w.buttonToggle("market_sl_toggle")
		}
	} else {
		if tpUse {
			w.buttonToggle("ls_tp_toggle")
		}
		if slUse {
			w.buttonToggle("ls_sl_toggle")
		}
	}

	log.WithFields(log.Fields{
		"tp.Use":      limits["tp"].IsUse,
		"tp.Distance": limits["tp"].Distance,
		"tp.Price":    limits["tp"].Price,
		"tp.Result":   limits["tp"].Result,
		"sl.Use":      limits["sl"].IsUse,
		"sl.Distance": limits["sl"].Distance,
		"sl.Price":    limits["sl"].Price,
		"sl.Result":   limits["sl"].Result,
	}).Info("set limit")
	return nil
}

func (w *orderWindow) buttonToggle(btnPath string) {
	toggle := w.Page.FindElementByCSS(domPaths[btnPath])
	if toggle != nil {
		toggle.Click()
	}
}

// Get current price
func (w *orderWindow) getPrice() (float64, error) {
	direction := w.Item.Direction
	if direction != BUY && direction != SELL {
		return 0, fmt.Errorf(fmt.Sprintf(unacceptableValue, direction))
	}
	err := w.checkOpen()
	if err != nil {
		return 0, err
	}

	tradeboxPrice := fmt.Sprintf(domPaths["tradebox_price"], direction, direction)
	s := w.Page.FindElementByCSS(tradeboxPrice)
	txt, _ := s.Text()
	price, err := strconv.ParseFloat(txt, 32)
	if err != nil {
		return 0, err
	}
	return price, nil
}
