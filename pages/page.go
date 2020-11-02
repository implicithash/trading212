package pages

import "github.com/tebeka/selenium"

// Page struct
type Page struct {
	Driver selenium.WebDriver
}

func (s *Page) driver() selenium.WebDriver {
	return s.Driver
}

// FindElementByID finds by ID
func (s *Page) FindElementByID(locator string) selenium.WebElement {
	element, _ := s.Driver.FindElement(selenium.ByID, locator)
	return element
}

// FindElementByXpath finds by Xpath
func (s *Page) FindElementByXpath(locator string) selenium.WebElement {
	element, _ := s.Driver.FindElement(selenium.ByXPATH, locator)
	return element
}

// FindElementByLinkText finds an element by link text
func (s *Page) FindElementByLinkText(locator string) selenium.WebElement {
	element, _ := s.Driver.FindElement(selenium.ByLinkText, locator)
	return element
}

// FindElementByPartialLink finds an element by partial link
func (s *Page) FindElementByPartialLink(locator string) selenium.WebElement {
	element, _ := s.Driver.FindElement(selenium.ByPartialLinkText, locator)
	return element
}

// FindElementByName finds an element by name
func (s *Page) FindElementByName(locator string) selenium.WebElement {
	element, _ := s.Driver.FindElement(selenium.ByName, locator)
	return element
}

// FindElementByTag finds an element by tag
func (s *Page) FindElementByTag(locator string) selenium.WebElement {
	element, _ := s.Driver.FindElement(selenium.ByTagName, locator)
	return element
}

// FindElementByClass finds an element by class
func (s *Page) FindElementByClass(locator string) selenium.WebElement {
	element, _ := s.Driver.FindElement(selenium.ByClassName, locator)
	return element
}

// FindElementByCSS finds an element by css
func (s *Page) FindElementByCSS(locator string) selenium.WebElement {
	element, _ := s.Driver.FindElement(selenium.ByCSSSelector, locator)
	return element
}

// FindElementsByID finds elements by ID
func (s *Page) FindElementsByID(locator string) []selenium.WebElement {
	element, _ := s.Driver.FindElements(selenium.ByID, locator)
	return element
}

// FindElementsByXpath finds elements by Xpath
func (s *Page) FindElementsByXpath(locator string) []selenium.WebElement {
	element, _ := s.Driver.FindElements(selenium.ByXPATH, locator)
	return element
}

// FindElementsByLinkText finds elements by link text
func (s *Page) FindElementsByLinkText(locator string) []selenium.WebElement {
	element, _ := s.Driver.FindElements(selenium.ByLinkText, locator)
	return element
}

// FindElementsByPartialLink finds elements by partial link
func (s *Page) FindElementsByPartialLink(locator string) []selenium.WebElement {
	element, _ := s.Driver.FindElements(selenium.ByPartialLinkText, locator)
	return element
}

// FindElementsByName finds elements by name
func (s *Page) FindElementsByName(locator string) []selenium.WebElement {
	element, _ := s.Driver.FindElements(selenium.ByName, locator)
	return element
}

// FindElementsByTag finds elements by tag
func (s *Page) FindElementsByTag(locator string) []selenium.WebElement {
	element, _ := s.Driver.FindElements(selenium.ByTagName, locator)
	return element
}

// FindElementsByClass finds elements by class
func (s *Page) FindElementsByClass(locator string) []selenium.WebElement {
	element, _ := s.Driver.FindElements(selenium.ByClassName, locator)
	return element
}

// FindElementsByCSS finds elements by css
func (s *Page) FindElementsByCSS(locator string) []selenium.WebElement {
	element, _ := s.Driver.FindElements(selenium.ByCSSSelector, locator)
	return element
}

// MouseHoverToElement hovers a mouse over
func (s *Page) MouseHoverToElement(locator string) selenium.WebElement {
	element, _ := s.Driver.FindElement(selenium.ByCSSSelector, locator)
	element.MoveTo(0, 0)
	return element
}

// FindCSSValue returns a value
func (s *Page) FindCSSValue(by, name string) string {
	we, _ := s.Driver.FindElement(by, name)
	var text string
	if we != nil {
		text, _ = we.Text()
	}
	return text
}

// Displayed function
func (s *Page) Displayed(by, elementName string) func(selenium.WebDriver) (bool, error) {
	return func(wd selenium.WebDriver) (bool, error) {
		elem, err := wd.FindElement(by, elementName)
		if err != nil {
			return false, nil
		}
		displayed, err := elem.IsEnabled()
		if err != nil {
			return false, nil
		}
		if !displayed {
			return false, nil
		}
		return true, nil
	}
}

// NotFound function
func (s *Page) NotFound(by, elementName string) func(selenium.WebDriver) (bool, error) {
	return func(wd selenium.WebDriver) (bool, error) {
		elem, err := wd.FindElement(by, elementName)
		/*if elem != nil {
			return false, nil
		}
		return true, nil*/
		if err != nil {
			return false, nil
		}
		displayed, err := elem.IsEnabled()
		if err != nil {
			return false, nil
		}
		if !displayed {
			return false, nil
		}
		return true, nil
	}
}
