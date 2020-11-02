package pages

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"time"
)

// HomePage contains a home page
type HomePage struct {
	Page Page
}

// GoToAccountPage directs to account page
func (p *HomePage) GoToAccountPage() (*AccountPage, error) {
	widget := p.Page.FindElementByCSS(domPaths["widget_message"])
	if widget != nil {
		we, _ := widget.FindElement(selenium.ByCSSSelector, domPaths["css_text"])
		text, _ := we.Text()
		if text == sessionExpired {
			if we, err := widget.FindElement(selenium.ByCSSSelector, domPaths["ok"]); err != nil {
				we.Click()
				time.Sleep(time.Millisecond * 3000)
			}
		}
	}
	return &AccountPage{Page: p.Page}, nil
}

// LoginToAccount logins to access an account page
func (p *HomePage) LoginToAccount(login, pswd string) *AccountPage {
	title, _ := p.Page.Driver.Title()
	log.Info(fmt.Sprintf("login page: %s", title))

	p.Page.FindElementByID(domPaths["login_id"]).SendKeys(login)
	p.Page.FindElementByID(domPaths["password_id"]).SendKeys(pswd)
	loginbtn := p.Page.FindElementByCSS(domPaths["login_btn"])
	loginbtn.Click()
	//time.Sleep(time.Millisecond * 3000)

	// check if we really were redirected to account page
	err := p.Page.Driver.WaitWithTimeout(p.Page.Displayed(selenium.ByCSSSelector, domPaths["nav_logo"]), time.Second*10)
	if err != nil {
		title, _ = p.Page.Driver.Title()
		log.Info(fmt.Sprintf("current page: %s", title))
		log.Printf(err.Error())
		return nil
	}
	title, _ = p.Page.Driver.Title()
	log.Info(fmt.Sprintf("logged in as %s, page: %s", login, title))

	//check if it's a weekend
	d := time.Now().Weekday()
	if /*mode == "demo"*/ time.Sunday <= d && d <= time.Saturday {
		we := p.Page.FindElementByCSS(domPaths["alert_box"])
		if we != nil {
			we.Click()
			log.Debug("weekend trading alert-box closed")
		}
	}
	return &AccountPage{Page: p.Page}
}
