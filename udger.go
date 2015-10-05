package udger

import (
	"database/sql"
	"errors"
	"os"
	"strings"

	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	_ "github.com/mattn/go-sqlite3"
)

func New(dbPath string) (*Udger, error) {
	u := &Udger{
		Browsers:     make(map[int]Browser),
		OS:           make(map[int]OS),
		Devices:      make(map[int]Device),
		browserTypes: make(map[int]string),
		browserOS:    make(map[int]int),
	}
	var err error

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, err
	}

	u.db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer u.db.Close()

	err = u.init()
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (this *Udger) Lookup(ua string) (*Info, error) {
	info := &Info{}

	browserId, err := this.findData(ua, this.RexBrowsers)
	if err != nil {
		return nil, err
	}

	info.Browser = this.Browsers[browserId]
	info.Browser.Type = this.browserTypes[info.Browser.typ]

	if val, ok := this.browserOS[browserId]; ok {
		info.OS = this.OS[val]
	} else {
		osId, err := this.findData(ua, this.RexOS)
		if err != nil {
			return nil, err
		}
		info.OS = this.OS[osId]
	}

	deviceId, err := this.findData(ua, this.RexDevices)
	if err != nil {
		return nil, err
	}
	if val, ok := this.Devices[deviceId]; ok {
		info.Device = val
	} else if info.Browser.typ == 3 { // if browser if mobile, we can guess its a mobile
		info.Device = Device{
			Name: "Smartphone",
			Icon: "phone.png",
		}
	}

	return info, nil
}

func (this *Udger) cleanRegex(r string) string {
	if strings.HasSuffix(r, "/si") {
		r = r[:len(r)-3]
	}
	if strings.HasPrefix(r, "/") {
		r = r[1:]
	}

	return r
}

func (this *Udger) findData(ua string, data []RexData) (int, error) {
	for i := 0; i < len(data); i++ {
		data[i].Regex = this.cleanRegex(data[i].Regex)
		r, err := pcre.Compile(data[i].Regex, pcre.CASELESS)
		if err != nil {
			return -1, errors.New(err.String())
		}
		matcher := r.MatcherString(ua, 0)
		match := matcher.MatchString(ua, 0)
		if !match {
			continue
		}
		return data[i].Id, nil
	}

	return -1, nil
}

func (this *Udger) init() error {
	rows, err := this.db.Query("SELECT browser, regstring FROM reg_browser ORDER by sequence ASC")
	if err != nil {
		return err
	}
	for rows.Next() {
		var d RexData
		rows.Scan(&d.Id, &d.Regex)
		this.RexBrowsers = append(this.RexBrowsers, d)
	}
	rows.Close()

	rows, err = this.db.Query("SELECT device, regstring FROM reg_device ORDER by sequence ASC")
	if err != nil {
		return err
	}
	for rows.Next() {
		var d RexData
		rows.Scan(&d.Id, &d.Regex)
		this.RexDevices = append(this.RexDevices, d)
	}
	rows.Close()

	rows, err = this.db.Query("SELECT os, regstring FROM reg_os ORDER by sequence ASC")
	if err != nil {
		return err
	}
	for rows.Next() {
		var d RexData
		rows.Scan(&d.Id, &d.Regex)
		this.RexOS = append(this.RexOS, d)
	}
	rows.Close()

	rows, err = this.db.Query("SELECT id,type,name,engine,company,icon FROM c_browser")
	if err != nil {
		return err
	}
	for rows.Next() {
		var d Browser
		var id int
		rows.Scan(&id, &d.typ, &d.Name, &d.Engine, &d.Company, &d.Icon)
		this.Browsers[id] = d
	}
	rows.Close()

	rows, err = this.db.Query("SELECT id, name, family, company, icon FROM c_os")
	if err != nil {
		return err
	}
	for rows.Next() {
		var d OS
		var id int
		rows.Scan(&id, &d.Name, &d.Family, &d.Company, &d.Icon)
		this.OS[id] = d
	}
	rows.Close()

	rows, err = this.db.Query("SELECT id, name, icon FROM c_device")
	if err != nil {
		return err
	}
	for rows.Next() {
		var d Device
		var id int
		rows.Scan(&id, &d.Name, &d.Icon)
		this.Devices[id] = d
	}
	rows.Close()

	rows, err = this.db.Query("SELECT type, name FROM c_browser_type")
	if err != nil {
		return err
	}
	for rows.Next() {
		var d string
		var id int
		rows.Scan(&id, &d)
		this.browserTypes[id] = d
	}
	rows.Close()

	rows, err = this.db.Query("SELECT browser, os FROM c_browser_os")
	if err != nil {
		return err
	}
	for rows.Next() {
		var browser int
		var os int
		rows.Scan(&browser, &os)
		this.browserOS[browser] = os
	}
	rows.Close()

	return nil
}