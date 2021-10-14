package clientIP

import (
	"github.com/oschwald/geoip2-golang"
	"net"
)

func GetCityByIP(ipAddr string) (string, error) {
	db, err := geoip2.Open("database/GeoLite2-City.mmdb")
	if err != nil {
		return "", err
	}
	defer db.Close()
	ip := net.ParseIP(ipAddr)
	record, err := db.City(ip)
	if err != nil {
		return "", err
	}
	if record.City.Names["zh-CN"] == "" {
		return record.Subdivisions[0].Names["zh-CN"], nil
	} else {
		return record.City.Names["zh-CN"], nil
	}
}
